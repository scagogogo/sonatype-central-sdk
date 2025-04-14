package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"
)

// executeWithRetry 执行HTTP请求并包含重试逻辑
func (c *Client) executeWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// 如果不是第一次尝试，等待退避时间
		if attempt > 0 {
			// 使用指数退避算法
			backoffTime := time.Duration(c.retryBackoffMs*int(math.Pow(2, float64(attempt-1)))) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffTime):
				// 继续重试
			}
		}

		// 克隆请求以避免重用请求体
		reqClone := req.Clone(ctx)
		resp, err = c.httpClient.Do(reqClone)

		// 判断是否需要重试
		if err != nil {
			// 网络错误，继续重试
			continue
		}

		// 根据状态码判断是否需要重试
		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			// 速率限制，需要重试
			if resp.Body != nil {
				resp.Body.Close()
			}
			continue
		case http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusBadGateway:
			// 服务器错误，需要重试
			if resp.Body != nil {
				resp.Body.Close()
			}
			continue
		default:
			// 其他情况不需要重试
			return resp, nil
		}
	}

	// 所有重试都失败
	if err != nil {
		return nil, err
	}

	// 处理HTTP错误
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return nil, ErrRateLimited
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case http.StatusForbidden:
		return nil, ErrForbidden
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	}

	if resp.StatusCode >= 500 {
		return nil, ErrServerError
	}

	return resp, nil
}

// doRequest 执行HTTP请求并处理响应
func (c *Client) doRequest(ctx context.Context, method, targetUrl string, body io.Reader, result interface{}) ([]byte, error) {
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, targetUrl, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "sonatype-central-sdk/1.0")
	if method == "POST" || method == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// 使用RetryWithBackoff执行带重试的请求
	var responseBody []byte
	err = RetryWithBackoff(
		ctx,
		c.maxRetries,
		c.retryBackoffMs,
		2.0,   // backoffFactor
		10000, // maxBackoffMs
		func() error {
			// 检查上下文是否已取消
			if err := ctx.Err(); err != nil {
				return err
			}

			// 执行请求
			resp, reqErr := c.httpClient.Do(req)
			if reqErr != nil {
				return reqErr
			}
			defer resp.Body.Close()

			// 读取响应体
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return readErr
			}

			// 处理HTTP错误
			if resp.StatusCode >= 400 {
				responseBody = body // 保存响应体以便外部函数可以使用
				return handleHttpError(resp.StatusCode, body)
			}

			// 成功，保存响应并返回
			responseBody = body
			return nil
		},
	)

	// 如果请求成功且需要解析响应
	if err == nil && result != nil && len(responseBody) > 0 {
		if jsonErr := json.Unmarshal(responseBody, result); jsonErr != nil {
			return responseBody, fmt.Errorf("解析JSON响应失败: %w", jsonErr)
		}
	}

	return responseBody, err
}

// downloadWithCache 从仓库下载文件，支持缓存
//
// 这是实现下载功能的核心内部方法，由Download方法调用。它处理URL构建、缓存查找、
// HTTP请求执行、重试逻辑和响应处理等细节。该方法首先尝试从缓存中获取内容（如果启用了缓存），
// 否则执行HTTP请求并处理各种可能的错误情况。成功下载后，如果启用了缓存，会将内容添加到缓存中。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求的超时和取消
//   - filePath: 文件在Maven仓库中的相对路径
//
// 返回:
//   - []byte: 下载文件的二进制内容
//   - error: 如果下载过程中出现以下错误，将返回相应的错误信息：
//     1. URL构建失败
//     2. 创建HTTP请求失败
//     3. 执行HTTP请求时网络错误
//     4. 读取响应体失败
//     5. 服务器返回HTTP错误状态码
//     6. 上下文取消或超时
//
// 错误处理:
//   - 对于某些HTTP错误（如429、500、502、503、504），会自动进行重试
//   - 重试次数和退避策略由Client配置决定
//   - 所有重试都失败后，返回最后一次尝试的错误
//
// 缓存行为:
//   - 如果启用了缓存且缓存中存在对应的内容，直接返回缓存内容而不发起HTTP请求
//   - 如果启用了缓存且成功下载文件，会将文件内容添加到缓存中，TTL由Client配置决定
func (c *Client) downloadWithCache(ctx context.Context, filePath string) ([]byte, error) {
	// 构建完整URL
	targetUrl, err := url.JoinPath(c.repoBaseURL, filePath)
	if err != nil {
		return nil, fmt.Errorf("URL构建失败: %w", err)
	}

	// 如果启用了缓存，尝试从缓存获取
	if c.cacheEnabled {
		cacheKey := "download:" + targetUrl
		if data, found := getFromCache(cacheKey); found {
			return data, nil
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", targetUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "sonatype-central-sdk/1.0")

	// 使用RetryWithBackoff执行带重试的请求
	var responseBody []byte
	err = RetryWithBackoff(
		ctx,
		c.maxRetries,
		c.retryBackoffMs,
		2.0,   // backoffFactor
		10000, // maxBackoffMs
		func() error {
			// 检查上下文是否已取消
			if err := ctx.Err(); err != nil {
				return err
			}

			// 执行请求
			resp, reqErr := c.httpClient.Do(req)
			if reqErr != nil {
				return reqErr
			}
			defer resp.Body.Close()

			// 读取响应体
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return readErr
			}

			// 处理HTTP错误
			if resp.StatusCode >= 400 {
				return handleHttpError(resp.StatusCode, body)
			}

			// 成功，保存响应
			responseBody = body
			return nil
		},
	)

	// 如果请求成功且启用了缓存，添加到缓存
	if err == nil && c.cacheEnabled {
		cacheKey := "download:" + targetUrl
		addToCache(cacheKey, responseBody, c.cacheTTLSeconds)
	}

	return responseBody, err
}

// isRetriableError 判断是否为可重试的错误
//
// 该方法用于确定HTTP响应状态码是否表示一个应该进行重试的暂时性错误。
// 在自动重试逻辑中，只有符合特定条件的错误才会触发重试机制，以避免对不可恢复的错误
// （如客户端错误、资源不存在等）进行无意义的重试，从而浪费资源。
//
// 被视为可重试的HTTP状态码包括:
//   - 429 Too Many Requests: 表示客户端发送了太多请求，服务器实施了限流
//   - 500 Internal Server Error: 服务器内部错误
//   - 502 Bad Gateway: 网关或代理服务器从上游服务器收到无效响应
//   - 503 Service Unavailable: 服务器暂时不可用（过载或维护）
//   - 504 Gateway Timeout: 网关或代理服务器等待上游服务器响应超时
//
// 参数:
//   - statusCode: HTTP响应状态码
//
// 返回:
//   - bool: 如果状态码表示可重试错误，返回true；否则返回false
func isRetriableError(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusInternalServerError ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}
