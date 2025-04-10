package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// parseJsonResponse 解析HTTP响应中的JSON内容到指定结构体
// 注意：此函数替代了原来的doRequest方法，避免命名冲突
func parseJsonResponse(resp *http.Response, result interface{}) error {
	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	// 解析 JSON 响应
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(result); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}
