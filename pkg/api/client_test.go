package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	// 测试默认客户端创建
	client := NewClient()
	assert.NotNil(t, client)
	assert.Equal(t, "https://search.maven.org", client.GetBaseURL())
	assert.Equal(t, "https://repo1.maven.org/maven2", client.GetRepoBaseURL())

	// 测试自定义选项
	customClient := NewClient(
		WithBaseURL("https://custom-maven.org"),
		WithRepoBaseURL("https://custom-repo.org"),
		WithProxy("http://proxy.example.com"),
		WithMaxRetries(5),
		WithRetryBackoff(1000),
		WithCache(true, 600),
	)
	assert.NotNil(t, customClient)
	assert.Equal(t, "https://custom-maven.org", customClient.GetBaseURL())
	assert.Equal(t, "https://custom-repo.org", customClient.GetRepoBaseURL())
	assert.Equal(t, "http://proxy.example.com", customClient.proxy)
	assert.Equal(t, 5, customClient.maxRetries)
	assert.Equal(t, 1000, customClient.retryBackoffMs)
	assert.True(t, customClient.cacheEnabled)
	assert.Equal(t, 600, customClient.cacheTTLSeconds)
}

func TestClientOptions(t *testing.T) {
	// 测试自定义HTTP客户端选项
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	client := NewClient(WithHTTPClient(httpClient))
	assert.Equal(t, httpClient, client.httpClient)

	// 测试缓存操作
	cacheClient := NewClient(WithCache(true, 100))
	assert.True(t, cacheClient.cacheEnabled)

	// 清除缓存
	cacheClient.ClearCache()
	// 由于缓存在内部实现，我们只能测试函数调用不会崩溃
}

func TestClientMethods(t *testing.T) {
	client := NewClient()

	// 测试URL获取方法
	assert.Equal(t, "https://search.maven.org", client.GetBaseURL())
	assert.Equal(t, "https://repo1.maven.org/maven2", client.GetRepoBaseURL())
}
