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
//
// 该结构体用于配置速率限制器的行为，包括重试策略和请求频率限制。
// 它允许开发者根据应用程序的具体需求和目标API的限制，调整客户端的请求模式。
// 合理的速率限制配置可以防止请求被服务器拒绝，同时优化资源利用和响应时间。
//
// 字段说明:
//   - MaxRetries: 操作失败时的最大重试次数（不包括首次尝试）
//   - InitialBackoffMs: 首次重试前等待的时间（毫秒）
//   - MaxBackoffMs: 重试间隔的最大值（毫秒），防止退避时间无限增长
//   - BackoffFactor: 每次重试后等待时间的增长因子，通常为2.0表示指数增长
//   - SearchRequestsPerSecond: 针对搜索操作，每秒允许的最大请求数
//   - DownloadRequestsPerSecond: 针对下载操作，每秒允许的最大请求数
//   - DefaultRequestsPerSecond: 针对其他操作，每秒允许的最大请求数
//   - EnableStats: 是否启用请求统计收集功能，用于监控和分析
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
//
// 这是一组预设的速率限制参数，在没有提供自定义配置时使用。
// 这些默认值经过调整，适合大多数一般用途场景，在保持良好性能的同时
// 避免触发常见API提供商的限流措施。
//
// 默认配置值:
//   - MaxRetries: 3 - 在放弃前最多重试3次，适合大多数临时性错误
//   - InitialBackoffMs: 500 - 首次重试前等待500毫秒，给服务器足够恢复时间
//   - MaxBackoffMs: 10000 - 最大等待10秒，避免过长等待影响用户体验
//   - BackoffFactor: 2.0 - 标准指数退避增长率
//   - SearchRequestsPerSecond: 2 - 适合大多数搜索API的温和请求频率
//   - DownloadRequestsPerSecond: 1 - 下载操作通常更占资源，因此频率更低
//   - DefaultRequestsPerSecond: 5 - 其他操作的适中频率
//   - EnableStats: true - 默认启用统计收集，便于监控和调试
//
// 如果这些默认值不满足特定需求（例如高吞吐量场景或严格限流的API），
// 请使用NewRateLimiterWithConfig创建自定义配置的速率限制器。
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
//
// RateLimiter是一个客户端侧的速率限制实现，用于控制对API端点的请求频率，
// 防止触发服务器的限流措施。它通过追踪每个主机的请求历史和操作类型，根据配置的
// 频率限制自动调整请求时机，必要时延迟请求的发送。
//
// 关键功能:
//   - 按主机和操作类型分别控制请求速率
//   - 支持不同操作类型的差异化速率限制（如搜索、下载等）
//   - 收集统计数据用于监控和分析
//   - 提供可配置的重试机制处理临时性错误
//
// 主要方法:
//   - WaitForRateLimit: 在发送请求前调用，自动等待适当时间
//   - GetStats: 获取请求统计信息
//   - GetTotalRequestCount/GetRequestCountByType: 获取特定请求计数
//   - ResetStats: 重置统计数据
//
// 并发安全性:
//   - 该结构体使用互斥锁(mutex)保证在并发环境中的数据一致性
//   - 可安全地在多个goroutine中使用同一个RateLimiter实例
type RateLimiter struct {
	mu              sync.Mutex                  // 用于并发安全
	lastRequestTime map[string]time.Time        // 记录每个主机的最后请求时间
	requestCounts   map[string]map[string]int64 // 记录每个主机每个时间窗口的请求数
	waitTimes       map[string]map[string]int64 // 记录每个主机每个操作类型的等待时间(毫秒)
	totalRequests   map[string]int64            // 记录总请求数
	config          RateLimitConfig             // 速率限制配置
}

// NewRateLimiter 创建一个新的速率限制器
//
// 该方法使用默认配置初始化一个新的速率限制器实例。默认配置设置了适当的重试策略和
// 速率限制参数，适用于大多数常规使用场景。默认配置包括:
// - 最大重试次数: 3
// - 初始退避时间: 500毫秒
// - 最大退避时间: 10000毫秒(10秒)
// - 退避因子: 2.0
// - 搜索请求限制: 每秒2个请求
// - 下载请求限制: 每秒1个请求
// - 默认请求限制: 每秒5个请求
// - 启用统计功能
//
// 返回:
//   - *RateLimiter: 配置好的速率限制器实例
//
// 使用示例:
//
//	// 创建一个使用默认配置的速率限制器
//	rateLimiter := NewRateLimiter()
//
//	// 在发送请求前使用它
//	waitTimeMs, err := rateLimiter.WaitForRateLimit(ctx, "repo1.maven.org", "search")
//	if err != nil {
//	    return err
//	}
//
//	// 发送请求...
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
//
// 该方法使用自定义配置初始化一个新的速率限制器实例。这允许开发者根据特定需求调整
// 速率限制的行为，例如在高性能环境中增加请求速率，或在受限环境中降低请求速率。
//
// 参数:
//   - config: 自定义速率限制配置，包含以下字段:
//   - MaxRetries: 最大重试次数
//   - InitialBackoffMs: 初始退避时间(毫秒)
//   - MaxBackoffMs: 最大退避时间(毫秒)
//   - BackoffFactor: 退避因子
//   - SearchRequestsPerSecond: 每秒允许的搜索请求数
//   - DownloadRequestsPerSecond: 每秒允许的下载请求数
//   - DefaultRequestsPerSecond: 每秒允许的默认请求数
//   - EnableStats: 是否启用请求统计
//
// 返回:
//   - *RateLimiter: 使用自定义配置的速率限制器实例
//
// 使用示例:
//
//	// 创建自定义配置
//	config := RateLimitConfig{
//	    MaxRetries:                5,    // 更多重试次数
//	    InitialBackoffMs:          200,  // 更短的初始等待时间
//	    MaxBackoffMs:              30000, // 更长的最大等待时间
//	    BackoffFactor:             1.5,  // 更平缓的退避增长
//	    SearchRequestsPerSecond:   5,    // 更高的搜索速率
//	    DownloadRequestsPerSecond: 2,    // 更高的下载速率
//	    DefaultRequestsPerSecond:  10,   // 更高的默认速率
//	    EnableStats:               true,
//	}
//
//	// 创建使用自定义配置的速率限制器
//	rateLimiter := NewRateLimiterWithConfig(config)
//
//	// 使用自定义限制器...
func NewRateLimiterWithConfig(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		lastRequestTime: make(map[string]time.Time),
		requestCounts:   make(map[string]map[string]int64),
		waitTimes:       make(map[string]map[string]int64),
		totalRequests:   make(map[string]int64),
		config:          config,
	}
}

// RetryWithBackoff 实现带指数退避的重试机制
//
// 这是SDK中所有重试逻辑的核心方法。它使用指数退避策略在操作失败时进行重试，
// 即每次重试的等待时间会逐渐增加，直到达到最大重试次数或操作成功。
// 在重试过程中，方法会检查上下文是否已取消，以便及时响应取消请求。
//
// 参数:
//   - ctx: 上下文对象，用于控制重试过程的取消和超时
//   - maxRetries: 最大重试次数，不包括首次尝试
//   - initialBackoffMs: 初始退避时间(毫秒)，第一次重试前等待的时间
//   - backoffFactor: 退避因子，用于计算每次重试的等待时间，通常设为2.0表示指数增长
//   - maxBackoffMs: 最大退避时间(毫秒)，重试等待时间不会超过此值
//   - operation: 要执行的操作函数，该函数应返回error，如果为nil表示操作成功
//
// 返回:
//   - error: 如果操作最终成功，返回nil；否则返回最后一次尝试的错误。
//     如果上下文被取消，返回ctx.Err()。
//
// 使用示例:
//
//	ctx := context.Background()
//
//	// 尝试执行可能失败的操作，最多重试3次
//	err := RetryWithBackoff(
//	    ctx,
//	    3,                // 最大重试3次
//	    100,              // 初始等待100毫秒
//	    2.0,              // 退避因子2.0，即等待时间会变为200ms, 400ms, 800ms...
//	    5000,             // 最大等待5000毫秒
//	    func() error {
//	        // 执行可能失败的操作
//	        resp, err := http.Get("https://example.com/api/data")
//	        if err != nil {
//	            return err // 操作失败，将进行重试
//	        }
//	        if resp.StatusCode == 429 {
//	            return errors.New("rate limited") // 被限流，将进行重试
//	        }
//	        // 操作成功，不再重试
//	        return nil
//	    },
//	)
//
//	if err != nil {
//	    log.Fatalf("操作失败: %v", err)
//	}
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
//
// 该方法用于判断特定错误是否可以进行重试。在重试逻辑中，并非所有的错误都适合重试，
// 例如参数验证错误或资源不存在等永久性错误就不适合重试。该方法会识别HTTP错误中的
// 可重试状态码（如429、500系列）以及网络连接错误等临时性错误。
//
// 参数:
//   - err: 要检查的错误对象
//
// 返回:
//   - bool: 如果错误是可重试的，返回true；否则返回false
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
//
// 该方法用于判断HTTP响应状态码是否表示一个应该进行重试的暂时性问题。
// 通常，客户端错误（4xx）中除了429（请求过多）外大多不适合重试，而服务器错误（5xx）
// 通常是临时性的，适合进行重试。这种区分有助于避免对不可恢复的错误进行无意义的重试，
// 同时确保在服务器临时故障时能够自动恢复。
//
// 参数:
//   - statusCode: HTTP响应状态码
//
// 返回:
//   - bool: 如果状态码表示可重试条件，返回true；否则返回false
//
// 可重试的状态码包括:
//   - 429 Too Many Requests: 请求频率超过限制
//   - 500 Internal Server Error: 服务器内部错误
//   - 502 Bad Gateway: 网关错误
//   - 503 Service Unavailable: 服务不可用（通常是临时的）
//   - 504 Gateway Timeout: 网关超时
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
//
// 该方法实现了客户端侧的速率限制，用于控制对特定主机和特定操作类型的请求频率。
// 它会根据配置的每秒请求数限制，计算需要等待的时间，并在发送请求前进行等待，
// 从而避免服务器端限流。同时，该方法还支持收集速率限制统计信息，用于监控和分析。
//
// 参数:
//   - ctx: 上下文对象，用于控制等待过程的取消
//   - host: 主机名，用于区分不同的目标服务器
//   - operationType: 操作类型，如"search"、"download"等，用于应用不同的限制策略
//
// 返回:
//   - int64: 实际等待的时间(毫秒)
//   - error: 如果等待过程被上下文取消，返回取消错误；否则返回nil
//
// 速率限制策略:
//   - "search": 使用SearchRequestsPerSecond配置值限制搜索类请求
//   - "download": 使用DownloadRequestsPerSecond配置值限制下载类请求
//   - 其他: 使用DefaultRequestsPerSecond配置值限制其他类型请求
//
// 使用示例:
//
//	// 创建一个速率限制器
//	rateLimiter := NewRateLimiter()
//
//	// 在发送搜索请求前调用
//	waitTimeMs, err := rateLimiter.WaitForRateLimit(ctx, "repo1.maven.org", "search")
//	if err != nil {
//	    // 处理等待被取消的情况
//	    return err
//	}
//
//	// 如果需要，可以记录等待时间
//	if waitTimeMs > 0 {
//	    log.Printf("为遵守速率限制等待了 %d 毫秒", waitTimeMs)
//	}
//
//	// 现在可以安全地发送请求
//	response, err := sendSearchRequest()
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
//
// 该方法返回速率限制器收集的所有统计数据，包括每个主机的请求总数、各操作类型的请求数
// 和等待时间。这些数据可用于监控和分析API使用情况，识别瓶颈或异常模式。
// 如果未启用统计功能（EnableStats=false），将返回一个仅包含stats_enabled=false的映射。
//
// 返回:
//   - map[string]interface{}: 包含统计信息的嵌套映射，结构如下:
//   - stats_enabled: 布尔值，表示统计功能是否启用
//   - hosts: 映射，每个主机的详细统计信息
//   - [主机名]: 主机统计信息
//   - total_requests: 对该主机的总请求数
//   - operations: 映射，各操作类型的统计
//   - [操作类型]: 操作统计信息
//   - count: 操作的请求数
//   - total_wait_ms: 为该操作类型等待的总毫秒数
//   - avg_wait_ms: 每个请求的平均等待毫秒数
//
// 使用示例:
//
//	rateLimiter := NewRateLimiter()
//
//	// 执行一些API请求...
//
//	// 获取统计信息
//	stats := rateLimiter.GetStats()
//
//	// 访问统计数据
//	if stats["stats_enabled"].(bool) {
//	    hosts := stats["hosts"].(map[string]interface{})
//	    for host, hostData := range hosts {
//	        hostInfo := hostData.(map[string]interface{})
//	        totalReqs := hostInfo["total_requests"].(int64)
//	        fmt.Printf("主机 %s 的总请求数: %d\n", host, totalReqs)
//
//	        // 查看各操作类型的数据
//	        operations := hostInfo["operations"].(map[string]map[string]int64)
//	        for opType, opData := range operations {
//	            fmt.Printf("  操作类型 %s: 请求数 %d, 总等待时间 %d ms\n",
//	                opType, opData["count"], opData["total_wait_ms"])
//	        }
//	    }
//	}
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
//
// 该方法返回针对指定主机发送的请求总数。这对于监控单个主机的使用情况非常有用，
// 特别是当应用程序与多个不同的API端点交互时。如果未启用统计功能或该主机没有记录，
// 将返回0。
//
// 参数:
//   - host: 主机名，如"repo1.maven.org"
//
// 返回:
//   - int64: 对指定主机的请求总数
//
// 使用示例:
//
//	rateLimiter := NewRateLimiter()
//
//	// 执行一些API请求...
//
//	// 检查特定主机的请求数
//	requestCount := rateLimiter.GetTotalRequestCount("repo1.maven.org")
//	fmt.Printf("已向Maven中央仓库发送%d个请求\n", requestCount)
func (rl *RateLimiter) GetTotalRequestCount(host string) int64 {
	if !rl.config.EnableStats {
		return 0
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	return rl.totalRequests[host]
}

// GetRequestCountByType 获取特定主机和操作类型的请求数
//
// 该方法返回针对指定主机和操作类型发送的请求数。这对于分析不同类型操作的使用模式
// 和负载分布非常有用，例如监控搜索操作与下载操作的比例。如果未启用统计功能或
// 找不到指定的主机或操作类型记录，将返回0。
//
// 参数:
//   - host: 主机名，如"repo1.maven.org"
//   - operationType: 操作类型，如"search"、"download"等
//
// 返回:
//   - int64: 针对指定主机和操作类型的请求数
//
// 使用示例:
//
//	rateLimiter := NewRateLimiter()
//
//	// 执行一些API请求...
//
//	// 检查特定操作类型的请求数
//	searchCount := rateLimiter.GetRequestCountByType("repo1.maven.org", "search")
//	downloadCount := rateLimiter.GetRequestCountByType("repo1.maven.org", "download")
//
//	fmt.Printf("搜索请求数: %d, 下载请求数: %d\n", searchCount, downloadCount)
//	fmt.Printf("下载/搜索比例: %.2f\n", float64(downloadCount)/float64(searchCount))
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

// ResetStats 重置所有统计信息
//
// 该方法清除速率限制器收集的所有统计数据，包括请求计数和等待时间。这在开始新的
// 测试周期、清除历史数据或解决内存占用问题时非常有用。重置操作使用互斥锁保证
// 并发安全。请注意，重置只影响统计数据，不会影响速率限制功能本身。
//
// 使用示例:
//
//	rateLimiter := NewRateLimiter()
//
//	// 执行一些API请求...
//
//	// 在测试完成后重置统计信息
//	rateLimiter.ResetStats()
//
//	// 确认统计已重置
//	stats := rateLimiter.GetStats()
//	// 如果启用了统计功能，此时应该显示所有主机的请求数为0
func (rl *RateLimiter) ResetStats() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.totalRequests = make(map[string]int64)
	rl.requestCounts = make(map[string]map[string]int64)
	rl.waitTimes = make(map[string]map[string]int64)
}
