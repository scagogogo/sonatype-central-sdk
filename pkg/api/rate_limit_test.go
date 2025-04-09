package api

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRateLimitHandling 测试客户端处理速率限制的能力
func TestRateLimitHandling(t *testing.T) {
	// 此测试可能会花费较长时间，跳过常规测试
	if testing.Short() {
		t.Skip("跳过速率限制测试")
	}

	// 创建一个低重试间隔的客户端
	client := NewClient(
		WithMaxRetries(3),
		WithRetryBackoff(100), // 100ms初始退避时间
	)

	// 并发发送多个请求，可能触发速率限制
	ctx := context.Background()
	var wg sync.WaitGroup
	requestCount := 20 // 尝试触发速率限制的请求数
	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 查询一个常见的依赖
			_, err := client.SearchByClassName(ctx, "guice", 5)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				t.Logf("请求错误: %v", err)
				errorCount++
			} else {
				successCount++
			}
		}()

		// 短暂暂停，使得请求不会太过快速
		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()

	t.Logf("成功请求: %d, 失败请求: %d", successCount, errorCount)

	// 即使在高并发时，也应该有一定比例的请求成功
	assert.True(t, successCount > 0, "至少应该有一些请求成功")
}

// TestRateLimitWithCache 测试缓存如何帮助减轻速率限制问题
func TestRateLimitWithCache(t *testing.T) {
	// 此测试可能会花费较长时间，跳过常规测试
	if testing.Short() {
		t.Skip("跳过速率限制缓存测试")
	}

	// 创建一个启用缓存的客户端
	client := NewClient(
		WithMaxRetries(2),
		WithRetryBackoff(300),
		WithCache(true, 600), // 缓存10分钟
	)

	// 清除之前的缓存
	client.ClearCache()

	ctx := context.Background()

	// 第一次请求，将结果缓存
	t.Log("第一次请求 - 应该从API获取数据")
	startTime := time.Now()
	result1, err := client.SearchByClassName(ctx, "guice", 5)
	duration1 := time.Since(startTime)
	assert.NoError(t, err)
	assert.NotEmpty(t, result1)

	// 第二次请求，应该从缓存中获取
	t.Log("第二次请求 - 应该从缓存获取数据")
	startTime = time.Now()
	result2, err := client.SearchByClassName(ctx, "guice", 5)
	duration2 := time.Since(startTime)
	assert.NoError(t, err)
	assert.NotEmpty(t, result2)

	// 缓存的请求应该明显快于API请求
	t.Logf("API请求用时: %v, 缓存请求用时: %v", duration1, duration2)
	assert.True(t, duration2 < duration1/2, "缓存请求应该至少快两倍")
}
