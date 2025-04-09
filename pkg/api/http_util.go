package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// doRequest 执行 HTTP 请求并解析 JSON 响应
func (c *Client) doRequest(req *http.Request, result interface{}) error {
	// 执行请求
	resp, err := c.executeWithRetry(req.Context(), req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
