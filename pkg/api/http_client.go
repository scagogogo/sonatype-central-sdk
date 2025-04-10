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

	// 执行请求，带重试逻辑
	var resp *http.Response
	var responseBody []byte
	maxRetries := c.maxRetries
	backoff := float64(c.retryBackoffMs)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避重试
			sleepTime := time.Duration(backoff) * time.Millisecond
			time.Sleep(sleepTime)
			backoff = math.Min(backoff*2, 10000) // 最大10秒
		}

		// 检查上下文是否已取消
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// 执行请求
		resp, err = c.httpClient.Do(req)
		if err != nil {
			if attempt < maxRetries {
				continue // 重试
			}
			return nil, fmt.Errorf("请求执行失败: %w", err)
		}

		// 读取响应体
		responseBody, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			if attempt < maxRetries {
				continue // 重试
			}
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		// 处理HTTP错误
		if resp.StatusCode >= 400 {
			err = handleHttpError(resp.StatusCode, responseBody)
			if isRetriableError(resp.StatusCode) && attempt < maxRetries {
				continue // 重试
			}
			return responseBody, err
		}

		// 成功
		break
	}

	// 如果需要解析响应
	if result != nil && len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, result); err != nil {
			return responseBody, fmt.Errorf("解析JSON响应失败: %w", err)
		}
	}

	return responseBody, nil
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

	// 执行请求，带重试逻辑
	var resp *http.Response
	var responseBody []byte
	maxRetries := c.maxRetries
	backoff := float64(c.retryBackoffMs)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避重试
			sleepTime := time.Duration(backoff) * time.Millisecond
			time.Sleep(sleepTime)
			backoff = math.Min(backoff*2, 10000) // 最大10秒
		}

		// 检查上下文是否已取消
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// 执行请求
		resp, err = c.httpClient.Do(req)
		if err != nil {
			if attempt < maxRetries {
				continue // 重试
			}
			return nil, fmt.Errorf("请求执行失败: %w", err)
		}

		// 读取响应体
		responseBody, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			if attempt < maxRetries {
				continue // 重试
			}
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		// 处理HTTP错误
		if resp.StatusCode >= 400 {
			err = handleHttpError(resp.StatusCode, responseBody)
			if isRetriableError(resp.StatusCode) && attempt < maxRetries {
				continue // 重试
			}
			return nil, err
		}

		// 成功
		break
	}

	// 如果启用了缓存，将结果添加到缓存
	if c.cacheEnabled {
		cacheKey := "download:" + targetUrl
		addToCache(cacheKey, responseBody, c.cacheTTLSeconds)
	}

	return responseBody, nil
}

// isRetriableError 判断是否为可重试的错误
func isRetriableError(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusInternalServerError ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}
