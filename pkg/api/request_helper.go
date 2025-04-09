package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/scagogogo/sonatype-central-sdk/pkg/request"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

type cachedResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

var (
	// 内存缓存实例
	memoryCache = cache.New(5*time.Minute, 10*time.Minute)

	// ErrRateLimited 速率限制错误
	ErrRateLimited = errors.New("rate limited by Sonatype Central API")

	// ErrNotFound 资源不存在错误
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthorized 未授权错误
	ErrUnauthorized = errors.New("unauthorized request")

	// ErrForbidden 禁止访问错误
	ErrForbidden = errors.New("forbidden request")

	// ErrBadRequest 请求格式错误
	ErrBadRequest = errors.New("bad request")

	// ErrServerError 服务器错误
	ErrServerError = errors.New("server error")
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

// searchRequestWithCache 执行搜索请求并应用缓存
func (c *Client) searchRequestWithCache(ctx context.Context, searchRequest *request.SearchRequest) (*http.Response, error) {
	targetUrl := fmt.Sprintf("%s/solrsearch/select?%s", c.baseURL, searchRequest.ToRequestParams())

	// 检查是否启用缓存
	if c.cacheEnabled {
		// 使用URL作为缓存键
		if cached, found := memoryCache.Get(targetUrl); found {
			cachedData := cached.(*cachedResponse)
			// 创建新的响应对象，带有缓存的数据
			resp := &http.Response{
				StatusCode: cachedData.StatusCode,
				Header:     cachedData.Headers,
				Body:       io.NopCloser(bytes.NewReader(cachedData.Body)),
			}
			return resp, nil
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}

	// 添加标准头部
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "sonatype-central-sdk/go")

	// 执行请求
	resp, err := c.executeWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if c.cacheEnabled && resp.StatusCode == http.StatusOK {
		// 读取并保存响应体，同时创建一个新的响应对象
		bodyData, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("读取响应体失败: %w", err)
		}
		resp.Body.Close() // 关闭原始响应体

		// 缓存响应数据
		cachedData := &cachedResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			Body:       bodyData,
		}
		memoryCache.Set(targetUrl, cachedData, time.Duration(c.cacheTTLSeconds)*time.Second)

		// 返回新的响应对象
		resp = &http.Response{
			StatusCode: resp.StatusCode,
			Header:     resp.Header,
			Body:       io.NopCloser(bytes.NewReader(bodyData)),
		}
	}

	return resp, nil
}

// SearchRequest 底层API，使用自定义客户端构造查询参数进行列表查询
func (c *Client) SearchRequest(ctx context.Context, searchRequest *request.SearchRequest) (*http.Response, error) {
	return c.searchRequestWithCache(ctx, searchRequest)
}

// SearchRequestJsonDoc 执行搜索请求并解析JSON响应
func SearchRequestJsonDoc[Doc any](c *Client, ctx context.Context, searchRequest *request.SearchRequest) (*response.Response[Doc], error) {
	resp, err := c.searchRequestWithCache(ctx, searchRequest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析JSON响应
	var result response.Response[Doc]
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("解析JSON响应失败: %w", err)
	}

	return &result, nil
}

// downloadWithCache 执行下载请求并应用缓存
func (c *Client) downloadWithCache(ctx context.Context, filePath string) ([]byte, error) {
	targetUrl := fmt.Sprintf("%s/remotecontent?filepath=%s", c.baseURL, filePath)

	// 检查是否启用缓存
	if c.cacheEnabled {
		// 使用URL作为缓存键
		if cachedData, found := memoryCache.Get(targetUrl); found {
			return cachedData.([]byte), nil
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}

	// 添加标准头部
	req.Header.Add("User-Agent", "sonatype-central-sdk/go")

	// 执行请求
	resp, err := c.executeWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 缓存结果
	if c.cacheEnabled && resp.StatusCode == http.StatusOK {
		memoryCache.Set(targetUrl, body, time.Duration(c.cacheTTLSeconds)*time.Second)
	}

	return body, nil
}

// 清除缓存
func (c *Client) ClearCache() {
	memoryCache.Flush()
}
