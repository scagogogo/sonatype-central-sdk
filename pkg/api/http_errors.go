package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

var (
	// ErrRateLimited 速率限制错误
	//
	// 表示请求因超过了服务器允许的频率而被拒绝。这通常是临时性错误，
	// 客户端应该降低请求频率或稍后重试。在使用SDK时，内置的速率限制器
	// 通常会自动处理这类错误，减少其发生频率。
	ErrRateLimited = errors.New("rate limited by Sonatype Central API")

	// ErrNotFound 资源不存在错误
	//
	// 表示请求的资源（如制品、类或版本）在服务器上不存在。这通常是永久性错误，
	// 重试相同的请求不太可能成功。客户端应该检查请求参数（如groupId, artifactId, version）
	// 是否正确，或考虑资源确实不存在的可能性。
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthorized 未授权错误
	//
	// 表示请求需要身份验证或提供的凭据无效。客户端应检查认证信息，
	// 或确认是否有权限访问请求的资源。某些API端点可能需要特定的访问令牌
	// 或订阅级别。
	ErrUnauthorized = errors.New("unauthorized request")

	// ErrForbidden 禁止访问错误
	//
	// 表示服务器理解请求但拒绝执行。这与未授权错误不同，未授权表示身份验证问题，
	// 而禁止访问表示即使通过了身份验证，也没有权限执行特定操作。客户端应检查
	// 是否有足够的权限或是否违反了使用政策。
	ErrForbidden = errors.New("forbidden request")

	// ErrBadRequest 请求格式错误
	//
	// 表示服务器无法理解请求，通常是由于请求参数格式错误、缺少必要参数或参数值无效。
	// 客户端应检查请求参数并确保符合API要求。在使用SDK时，可能需要检查所提供的
	// 搜索条件、过滤器或其他输入值。
	ErrBadRequest = errors.New("bad request")

	// ErrServerError 服务器错误
	//
	// 表示服务器在处理请求时遇到了意外情况。这通常是临时性错误，
	// 可能在稍后重试时解决。如果错误持续存在，可能表明服务器存在更严重的问题。
	// SDK会自动重试适合的服务器错误，但连续失败后仍会返回此错误。
	ErrServerError = errors.New("server error")
)

// cachedResponse 表示一个缓存的HTTP响应
//
// 该结构体用于在内存中临时存储HTTP响应的内容，包括状态码、响应体和头信息。
// 通过缓存常用请求的响应，可以减少重复的网络请求，提高应用程序性能，
// 特别是对于那些频繁请求但内容变化不大的API端点。
//
// 字段说明:
//   - StatusCode: HTTP响应状态码
//   - Body: 响应体的二进制内容
//   - Headers: 响应头信息，可用于检查内容类型、日期等元数据
type cachedResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// 内存缓存实例，全局共享，默认5分钟过期时间，10分钟清理一次
//
// 该缓存使用github.com/patrickmn/go-cache库实现，提供线程安全的内存缓存功能。
// 缓存项默认在5分钟后过期，过期的项目会在每10分钟进行一次自动清理。
// 这种机制确保了缓存的数据不会无限增长占用内存，同时保持了合理的新鲜度。
//
// 缓存键通常由请求URL组成，而值为cachedResponse结构体，包含完整的响应内容。
// 此缓存主要用于减少对相同资源的重复请求，特别适合于搜索结果和元数据信息等
// 短期内不太可能发生变化的数据。
var memoryCache = cache.New(5*time.Minute, 10*time.Minute)

// handleHttpError 根据HTTP状态码处理错误
//
// 这是SDK中错误处理的核心方法，用于将HTTP错误转换为对客户端友好的错误对象。
// 它尝试从响应体中解析错误信息，如果无法解析，则使用HTTP状态文本作为默认错误消息。
// 该方法对常见的HTTP错误状态码进行特殊处理，生成包含详细错误描述的APIError对象。
//
// 参数:
//   - statusCode: HTTP响应状态码
//   - responseBody: HTTP响应体的原始内容，可能包含错误信息
//
// 返回:
//   - error: 返回封装了HTTP错误信息的APIError对象，该对象实现了error接口
//
// 错误处理逻辑:
//   - 429 Too Many Requests: 转换为限流错误，提示客户端请求频率过高
//   - 404 Not Found: 转换为资源不存在错误
//   - 401 Unauthorized: 转换为未授权访问错误
//   - 403 Forbidden: 转换为禁止访问错误
//   - 400 Bad Request: 转换为请求参数错误
//   - 5xx: 转换为服务器错误
//   - 其他: 生成包含状态码和消息的通用错误
//
// 响应体解析:
//   - 尝试以JSON格式解析响应体，寻找标准错误字段("message"或"error")
//   - 如果解析失败，使用HTTP状态文本作为错误消息，原始响应体作为详细信息
func handleHttpError(statusCode int, responseBody []byte) error {
	var message, details string

	// 尝试解析错误响应
	var errResp response.ErrorResponse
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
		return &response.APIError{
			Code:    fmt.Sprintf("%d", statusCode),
			Message: "请求频率过高，已被限流: " + message,
		}
	case http.StatusNotFound:
		return &response.APIError{
			Code:    fmt.Sprintf("%d", statusCode),
			Message: "资源不存在: " + message,
		}
	case http.StatusUnauthorized:
		return &response.APIError{
			Code:    fmt.Sprintf("%d", statusCode),
			Message: "未授权访问: " + message,
		}
	case http.StatusForbidden:
		return &response.APIError{
			Code:    fmt.Sprintf("%d", statusCode),
			Message: "禁止访问: " + message,
		}
	case http.StatusBadRequest:
		return &response.APIError{
			Code:    fmt.Sprintf("%d", statusCode),
			Message: "请求参数错误: " + message,
		}
	default:
		if statusCode >= 500 {
			return &response.APIError{
				Code:    fmt.Sprintf("%d", statusCode),
				Message: "服务器错误: " + message,
			}
		}

		msg := message
		if details != "" {
			msg = message + " - " + details
		}

		return &response.APIError{
			Code:    fmt.Sprintf("%d", statusCode),
			Message: msg,
		}
	}
}
