package api

import (
	"context"
	"errors"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/scagogogo/sonatype-central-sdk/pkg/response"
)

// RateLimitConfig 定义速率限制配置参数
type RateLimitConfig struct {
	MaxRetries       int     // 最大重试次数
	InitialBackoffMs int     // 初始退避时间(毫秒)
	MaxBackoffMs     int     // 最大退避时间(毫秒)
	BackoffFactor    float64 // 退避因子，用于计算下一次退避时间

	// 新增配置项
	SearchRequestsPerSecond   int // 每秒允许的搜索请求数
	DownloadRequestsPerSecond int // 每秒允许的下载请求数
	DefaultRequestsPerSecond  int // 每秒允许的默认请求数

	// 是否启用请求计数统计
	EnableStats bool
}

// DefaultRateLimitConfig 默认的速率限制配置
var DefaultRateLimitConfig = RateLimitConfig{
	MaxRetries:       3,
	InitialBackoffMs: 500,
	MaxBackoffMs:     10000,
	BackoffFactor:    2.0,

	SearchRequestsPerSecond:   2,
	DownloadRequestsPerSecond: 1,
	DefaultRequestsPerSecond:  5,

	EnableStats: true,
}

// RateLimiter 速率限制器，管理API请求速率
type RateLimiter struct {
	mu              sync.Mutex                  // 用于并发安全
	lastRequestTime map[string]time.Time        // 记录每个主机的最后请求时间
	requestCounts   map[string]map[string]int64 // 记录每个主机每个时间窗口的请求数
	waitTimes       map[string]map[string]int64 // 记录每个主机每个操作类型的等待时间(毫秒)
	totalRequests   map[string]int64            // 记录总请求数
	config          RateLimitConfig             // 速率限制配置
}

// NewRateLimiter 创建一个新的速率限制器
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		lastRequestTime: make(map[string]time.Time),
		requestCounts:   make(map[string]map[string]int64),
		waitTimes:       make(map[string]map[string]int64),
		totalRequests:   make(map[string]int64),
		config:          DefaultRateLimitConfig,
	}
}

// NewRateLimiterWithConfig 创建一个带有自定义配置的速率限制器
func NewRateLimiterWithConfig(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		lastRequestTime: make(map[string]time.Time),
		requestCounts:   make(map[string]map[string]int64),
		waitTimes:       make(map[string]map[string]int64),
		totalRequests:   make(map[string]int64),
		config:          config,
	}
}

// RetryWithBackoff 使用指数退避策略重试操作
// ctx: 上下文，用于取消操作
// maxRetries: 最大重试次数
// initialBackoffMs: 初始退避时间(毫秒)
// backoffFactor: 退避因子，用于计算下一次退避时间
// maxBackoffMs: 最大退避时间(毫秒)
// operation: 要执行的操作函数
func RetryWithBackoff(
	ctx context.Context,
	maxRetries int,
	initialBackoffMs int,
	backoffFactor float64,
	maxBackoffMs int,
	operation func() error,
) error {
	backoff := float64(initialBackoffMs)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 第一次尝试前不等待
		if attempt > 0 {
			// 指数退避重试
			sleepTime := time.Duration(backoff) * time.Millisecond

			// 使用带超时的上下文等待
			timer := time.NewTimer(sleepTime)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
				// 时间到，继续执行
			}

			// 计算下一次退避时间，但不超过最大值
			backoff = math.Min(backoff*backoffFactor, float64(maxBackoffMs))
		}

		// 执行操作
		err := operation()

		// 如果操作成功或者是不应该重试的错误，则返回
		if err == nil || !shouldRetryError(err) {
			return err
		}

		// 如果是最后一次尝试，返回错误
		if attempt == maxRetries {
			return err
		}

		// 否则继续重试
	}

	// 不应该到达这里
	return errors.New("重试失败: 已超过最大重试次数")
}

// shouldRetryError 检查是否应该重试错误
func shouldRetryError(err error) bool {
	// 检查是否是HTTP错误
	var httpErr *response.HTTPError
	if errors.As(err, &httpErr) {
		return isRetriableStatusCode(httpErr.StatusCode)
	}

	// 检查是否是网络连接错误或者超时错误
	// 这里可以添加更多可重试的错误类型
	return false
}

// isRetriableStatusCode 检查HTTP状态码是否表示可重试的错误
func isRetriableStatusCode(statusCode int) bool {
	// 429: 请求过多 (Too Many Requests)
	// 500, 502, 503, 504: 服务器错误
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusInternalServerError ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}

// WaitForRateLimit 根据主机名和操作类型等待适当的时间以遵守速率限制
// 返回等待的时间（毫秒）
func (rl *RateLimiter) WaitForRateLimit(ctx context.Context, host string, operationType string) (int64, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// 初始化该主机的计数器（如果不存在）
	if _, exists := rl.requestCounts[host]; !exists {
		rl.requestCounts[host] = make(map[string]int64)
	}

	// 初始化该主机的等待时间统计（如果不存在且启用了统计）
	if rl.config.EnableStats {
		if _, exists := rl.waitTimes[host]; !exists {
			rl.waitTimes[host] = make(map[string]int64)
		}
	}

	// 获取上次请求时间
	lastTime, exists := rl.lastRequestTime[host]

	// 计算应该等待的时间（毫秒）
	var waitTimeMs int64 = 0

	if exists {
		// 不同的操作类型可能有不同的速率限制规则
		switch operationType {
		case "search":
			// 对搜索操作限制
			requestsPerSecond := rl.config.SearchRequestsPerSecond
			if requestsPerSecond <= 0 {
				requestsPerSecond = 1 // 默认每秒允许1个请求
			}

			// 计算每个请求应该间隔的毫秒数
			intervalMs := int64(1000 / requestsPerSecond)

			// 计算已经过去的时间
			elapsedMs := now.Sub(lastTime).Milliseconds()

			// 如果过去的时间小于间隔，则需要等待
			if elapsedMs < intervalMs {
				waitTimeMs = intervalMs - elapsedMs

				// 监听上下文取消信号
				timer := time.NewTimer(time.Duration(waitTimeMs) * time.Millisecond)
				select {
				case <-ctx.Done():
					timer.Stop()
					return 0, ctx.Err()
				case <-timer.C:
					// 等待完成，继续
				}
			}

		case "download":
			// 对下载操作限制
			requestsPerSecond := rl.config.DownloadRequestsPerSecond
			if requestsPerSecond <= 0 {
				requestsPerSecond = 1 // 默认每秒允许1个请求
			}

			// 计算每个请求应该间隔的毫秒数
			intervalMs := int64(1000 / requestsPerSecond)

			// 计算已经过去的时间
			elapsedMs := now.Sub(lastTime).Milliseconds()

			// 如果过去的时间小于间隔，则需要等待
			if elapsedMs < intervalMs {
				waitTimeMs = intervalMs - elapsedMs

				// 监听上下文取消信号
				timer := time.NewTimer(time.Duration(waitTimeMs) * time.Millisecond)
				select {
				case <-ctx.Done():
					timer.Stop()
					return 0, ctx.Err()
				case <-timer.C:
					// 等待完成，继续
				}
			}

		default:
			// 对其他操作限制
			requestsPerSecond := rl.config.DefaultRequestsPerSecond
			if requestsPerSecond <= 0 {
				requestsPerSecond = 5 // 默认每秒允许5个请求
			}

			// 计算每个请求应该间隔的毫秒数
			intervalMs := int64(1000 / requestsPerSecond)

			// 计算已经过去的时间
			elapsedMs := now.Sub(lastTime).Milliseconds()

			// 如果过去的时间小于间隔，则需要等待
			if elapsedMs < intervalMs {
				waitTimeMs = intervalMs - elapsedMs

				// 监听上下文取消信号
				timer := time.NewTimer(time.Duration(waitTimeMs) * time.Millisecond)
				select {
				case <-ctx.Done():
					timer.Stop()
					return 0, ctx.Err()
				case <-timer.C:
					// 等待完成，继续
				}
			}
		}
	}

	// 更新最后请求时间和计数
	rl.lastRequestTime[host] = time.Now()
	if rl.config.EnableStats {
		rl.totalRequests[host]++
		rl.requestCounts[host][operationType]++
		rl.waitTimes[host][operationType] += waitTimeMs
	}

	return waitTimeMs, nil
}

// GetStats 获取速率限制器的统计信息
func (rl *RateLimiter) GetStats() map[string]interface{} {
	if !rl.config.EnableStats {
		return map[string]interface{}{
			"stats_enabled": false,
		}
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	stats := make(map[string]interface{})
	stats["stats_enabled"] = true

	// 统计每个主机信息
	hostStats := make(map[string]interface{})
	for host, totalRequests := range rl.totalRequests {
		hostInfo := make(map[string]interface{})
		hostInfo["total_requests"] = totalRequests

		// 操作类型统计
		opStats := make(map[string]map[string]int64)
		for opType, count := range rl.requestCounts[host] {
			opData := make(map[string]int64)
			opData["count"] = count
			if waitTime, ok := rl.waitTimes[host][opType]; ok {
				opData["total_wait_ms"] = waitTime
				if count > 0 {
					opData["avg_wait_ms"] = waitTime / count
				}
			}
			opStats[opType] = opData
		}
		hostInfo["operations"] = opStats

		hostStats[host] = hostInfo
	}
	stats["hosts"] = hostStats

	return stats
}

// GetTotalRequestCount 获取特定主机的总请求数
func (rl *RateLimiter) GetTotalRequestCount(host string) int64 {
	if !rl.config.EnableStats {
		return 0
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	return rl.totalRequests[host]
}

// GetRequestCountByType 获取特定主机和操作类型的请求数
func (rl *RateLimiter) GetRequestCountByType(host string, operationType string) int64 {
	if !rl.config.EnableStats {
		return 0
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if counts, ok := rl.requestCounts[host]; ok {
		return counts[operationType]
	}
	return 0
}

// ResetStats 重置统计信息
func (rl *RateLimiter) ResetStats() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.totalRequests = make(map[string]int64)
	rl.requestCounts = make(map[string]map[string]int64)
	rl.waitTimes = make(map[string]map[string]int64)
}
