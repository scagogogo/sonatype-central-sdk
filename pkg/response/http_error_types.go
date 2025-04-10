package response

import "fmt"

// HTTPError 表示HTTP请求过程中发生的错误
type HTTPError struct {
	StatusCode int    `json:"statusCode"` // HTTP状态码
	Message    string `json:"message"`    // 错误信息
	URL        string `json:"url"`        // 请求的URL
}

// Error 实现error接口
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP错误 %d: %s (URL: %s)", e.StatusCode, e.Message, e.URL)
}

// APIError 表示API调用过程中发生的错误
type APIError struct {
	Code    string `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误消息
}

// Error 实现error接口
func (e *APIError) Error() string {
	return fmt.Sprintf("API错误 [%s]: %s", e.Code, e.Message)
}

// ErrorResponse 表示API错误响应
type ErrorResponse struct {
	Status  int       `json:"status"`            // HTTP状态码
	Error   string    `json:"error"`             // 错误类型
	Message string    `json:"message"`           // 错误消息
	Details *APIError `json:"details,omitempty"` // 详细错误信息
}
