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
func isRetriableError(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusInternalServerError ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}
