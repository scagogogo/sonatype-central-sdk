package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
)

var (
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

// cachedResponse 表示一个缓存的HTTP响应
type cachedResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// 内存缓存实例，全局共享，默认5分钟过期时间，10分钟清理一次
var memoryCache = cache.New(5*time.Minute, 10*time.Minute)

// APIError 表示API返回的错误
type APIError struct {
	StatusCode int
	Message    string
	Details    string
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("API错误 [%d]: %s - %s", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("API错误 [%d]: %s", e.StatusCode, e.Message)
}

// ErrorResponse API错误响应的JSON结构
type ErrorResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Status  int    `json:"status,omitempty"`
}

// handleHttpError 根据HTTP状态码处理错误
func handleHttpError(statusCode int, responseBody []byte) error {
	var message, details string

	// 尝试解析错误响应
	var errResp ErrorResponse
	if err := json.Unmarshal(responseBody, &errResp); err == nil {
		if errResp.Message != "" {
			message = errResp.Message
		} else if errResp.Error != "" {
			message = errResp.Error
		}
	}

	// 如果没有解析到错误消息，使用默认消息
	if message == "" {
		message = http.StatusText(statusCode)
		details = string(responseBody)
	}

	// 根据状态码处理特定错误
	switch statusCode {
	case http.StatusTooManyRequests:
		return &APIError{
			StatusCode: statusCode,
			Message:    "请求频率过高，已被限流",
			Details:    message,
		}
	case http.StatusNotFound:
		return &APIError{
			StatusCode: statusCode,
			Message:    "资源不存在",
			Details:    message,
		}
	case http.StatusUnauthorized:
		return &APIError{
			StatusCode: statusCode,
			Message:    "未授权访问",
			Details:    message,
		}
	case http.StatusForbidden:
		return &APIError{
			StatusCode: statusCode,
			Message:    "禁止访问",
			Details:    message,
		}
	case http.StatusBadRequest:
		return &APIError{
			StatusCode: statusCode,
			Message:    "请求参数错误",
			Details:    message,
		}
	default:
		if statusCode >= 500 {
			return &APIError{
				StatusCode: statusCode,
				Message:    "服务器错误",
				Details:    message,
			}
		}
		return &APIError{
			StatusCode: statusCode,
			Message:    message,
			Details:    details,
		}
	}
}
