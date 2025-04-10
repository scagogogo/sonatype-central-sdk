package api

import (
	"testing"
)

// createRealClient 创建一个连接到真实Maven Central API的客户端
// 该客户端已配置适当的重试和速率限制参数，以避免对API服务器造成过大负担
func createRealClient(t *testing.T) *Client {
	// 创建默认客户端实例（使用真实API地址）
	client := NewClient(
		WithMaxRetries(3),     // 设置更多重试次数以应对临时网络问题
		WithRetryBackoff(800), // 较长的重试间隔，避免过快重试
		WithCache(true, 3600), // 启用长时间缓存以减少对API的请求
	)

	// 在测试结束时清除缓存
	t.Cleanup(func() {
		client.ClearCache()
	})

	return client
}

// minInt 返回两个整数中较小的一个（用于测试中限制数组索引）
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
