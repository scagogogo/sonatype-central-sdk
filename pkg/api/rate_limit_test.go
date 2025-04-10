package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
	"github.com/stretchr/testify/assert"
)

// TestRateLimiterBasic 测试基本的速率限制器功能
func TestRateLimiterBasic(t *testing.T) {
	rateLimiter := NewRateLimiter()
	ctx := context.Background()

	// 第一次请求不应该等待
	waitTime, err := rateLimiter.WaitForRateLimit(ctx, "test.example.com", "search")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), waitTime)

	// 第二次请求应该等待
	waitTime, err = rateLimiter.WaitForRateLimit(ctx, "test.example.com", "search")
	assert.NoError(t, err)
	assert.Greater(t, waitTime, int64(0))
}

// TestRateLimiterWithCustomConfig 测试自定义配置的速率限制器
func TestRateLimiterWithCustomConfig(t *testing.T) {
	config := RateLimitConfig{
		SearchRequestsPerSecond:   10, // 每秒10次搜索请求
		DownloadRequestsPerSecond: 5,  // 每秒5次下载请求
		DefaultRequestsPerSecond:  20, // 每秒20次默认请求
		EnableStats:               true,
	}
	rateLimiter := NewRateLimiterWithConfig(config)
	ctx := context.Background()

	// 测试搜索请求速率
	waitTime, err := rateLimiter.WaitForRateLimit(ctx, "test.example.com", "search")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), waitTime)

	// 由于设置了每秒10次搜索请求，理论上两次请求之间的间隔是100ms
	// 但第一次请求不等待，所以第二次请求应该等待的时间应该接近100ms
	waitTime, err = rateLimiter.WaitForRateLimit(ctx, "test.example.com", "search")
	assert.NoError(t, err)
	assert.Greater(t, waitTime, int64(0))
	assert.LessOrEqual(t, waitTime, int64(100))
}

// TestRetryWithBackoff 测试指数退避重试机制
func TestRetryWithBackoff(t *testing.T) {
	ctx := context.Background()
	maxRetries := 3
	initialBackoffMs := 10 // 短间隔便于测试
	backoffFactor := 2.0
	maxBackoffMs := 1000

	// 测试成功情况
	attempts := 0
	err := RetryWithBackoff(
		ctx,
		maxRetries,
		initialBackoffMs,
		backoffFactor,
		maxBackoffMs,
		func() error {
			attempts++
			return nil // 第一次就成功
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts, "应该只尝试一次")

	// 测试多次重试后成功
	attempts = 0
	err = RetryWithBackoff(
		ctx,
		maxRetries,
		initialBackoffMs,
		backoffFactor,
		maxBackoffMs,
		func() error {
			attempts++
			if attempts < 3 {
				// 模拟可重试的错误
				return &response.HTTPError{
					StatusCode: http.StatusTooManyRequests,
					Message:    "Rate limited",
					URL:        "https://example.com",
				}
			}
			return nil // 第三次成功
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts, "应该尝试3次才成功")

	// 测试超过最大重试次数
	attempts = 0
	err = RetryWithBackoff(
		ctx,
		maxRetries,
		initialBackoffMs,
		backoffFactor,
		maxBackoffMs,
		func() error {
			attempts++
			// 始终返回可重试的错误
			return &response.HTTPError{
				StatusCode: http.StatusTooManyRequests,
				Message:    "Rate limited",
				URL:        "https://example.com",
			}
		},
	)
	assert.Error(t, err)
	assert.Equal(t, maxRetries+1, attempts, "应该尝试最大重试次数+1次")

	// 测试非可重试的错误
	attempts = 0
	nonRetryableErr := &response.HTTPError{
		StatusCode: http.StatusBadRequest,
		Message:    "Bad request",
		URL:        "https://example.com",
	}
	err = RetryWithBackoff(
		ctx,
		maxRetries,
		initialBackoffMs,
		backoffFactor,
		maxBackoffMs,
		func() error {
			attempts++
			return nonRetryableErr
		},
	)
	assert.Error(t, err)
	assert.Equal(t, nonRetryableErr, err)
	assert.Equal(t, 1, attempts, "非可重试错误应该只尝试一次")

	// 测试上下文取消
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消
	attempts = 0
	err = RetryWithBackoff(
		cancelCtx,
		maxRetries,
		initialBackoffMs,
		backoffFactor,
		maxBackoffMs,
		func() error {
			attempts++
			return &response.HTTPError{
				StatusCode: http.StatusTooManyRequests,
				Message:    "Rate limited",
				URL:        "https://example.com",
			}
		},
	)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestRateLimitWithCache 测试结合缓存的速率限制
func TestRateLimitWithCache(t *testing.T) {
	// 这里可以添加测试用例来验证结合缓存的速率限制功能
	// 暂时跳过
	t.Skip("暂未实现")
}
