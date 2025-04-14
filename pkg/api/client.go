package api

import (
	"net/http"
	"time"
)

// ClientOption 客户端配置选项函数
//
// ClientOption是一个函数类型，用于实现选项模式来配置Client实例。
// 该模式允许使用者以灵活的方式初始化客户端，只指定需要自定义的选项，
// 而使用默认值处理其他配置。
//
// 使用选项模式的好处:
//   - 避免了多个构造函数或复杂的配置结构体
//   - 支持未来扩展新选项而不破坏API兼容性
//   - 提高了代码可读性，使配置过程更直观
//
// 使用示例:
//
//	// 创建一个使用默认配置的客户端
//	client := NewClient()
//
//	// 创建一个带自定义选项的客户端
//	customClient := NewClient(
//	    WithProxy("http://my-proxy:8080"),
//	    WithMaxRetries(5),
//	    WithCache(true, 600),
//	)
type ClientOption func(*Client)

// Client Sonatype Central 客户端
//
// Client是SDK的核心结构体，提供与Sonatype Central Maven仓库交互的所有功能。
// 它封装了HTTP请求管理、缓存、重试机制和响应处理，为上层API提供统一的访问接口。
//
// 主要功能:
//   - 搜索: 按照各种条件搜索Maven制品
//   - 下载: 获取制品的JAR、POM、源码等文件
//   - 类搜索: 按类名、包名或接口查找制品
//   - 版本管理: 获取制品的版本信息和元数据
//
// 客户端使用可选配置项进行初始化，同时提供合理的默认值以简化使用。
// 它是线程安全的，可以在多个goroutine中共享同一个实例。
//
// 使用示例:
//
//	// 创建客户端
//	client := api.NewClient()
//
//	// 搜索制品
//	artifacts, err := client.SearchByGroupAndArtifact(ctx, "org.apache.commons", "commons-lang3")
//
//	// 下载JAR文件
//	jarData, err := client.DownloadJar(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
type Client struct {
	// 代理服务器地址
	proxy string

	// 基础URL，默认为 https://search.maven.org
	baseURL string

	// 下载文件时使用的基础URL，默认为 https://repo1.maven.org/maven2
	repoBaseURL string

	// HTTP客户端，可自定义
	httpClient *http.Client

	// 最大重试次数
	maxRetries int

	// 重试间隔基准时间（毫秒）
	retryBackoffMs int

	// 是否启用缓存
	cacheEnabled bool

	// 缓存过期时间（秒）
	cacheTTLSeconds int
}

// WithProxy 设置代理服务器
//
// 该选项用于配置客户端通过HTTP代理服务器访问Maven仓库。
// 在某些环境中，如企业网络或特定国家/地区，可能需要通过代理才能访问外部资源。
// 指定的代理URL应该包含协议、主机名和端口号。
//
// 参数:
//   - proxy: 代理服务器URL，格式如"http://proxy-host:port"或"socks5://proxy-host:port"
//
// 返回:
//   - ClientOption: 一个可以应用到NewClient的配置函数
//
// 使用示例:
//
//	client := api.NewClient(
//	    api.WithProxy("http://corporate-proxy.example.com:8080"),
//	)
func WithProxy(proxy string) ClientOption {
	return func(c *Client) {
		c.proxy = proxy
	}
}

// WithBaseURL 设置自定义基础URL
//
// 该选项用于配置客户端使用的搜索API基础URL。这在以下情况下特别有用:
// - 使用私有Maven仓库或镜像
// - 通过CDN或负载均衡器访问官方仓库
// - 测试或开发环境中需要模拟API响应
//
// 参数:
//   - baseURL: 搜索API的基础URL，如"https://search.maven.org"
//
// 返回:
//   - ClientOption: 一个可以应用到NewClient的配置函数
//
// 使用示例:
//
//	client := api.NewClient(
//	    api.WithBaseURL("https://maven.mycompany.com/search"),
//	)
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithRepoBaseURL 设置自定义仓库基础URL
//
// 该选项用于配置客户端下载制品时使用的仓库基础URL。
// 下载URL与搜索URL通常是不同的，因此需要单独配置。
// 这对于指向自定义Maven镜像或私有仓库特别有用。
//
// 参数:
//   - repoBaseURL: 仓库的基础URL，如"https://repo1.maven.org/maven2"
//
// 返回:
//   - ClientOption: 一个可以应用到NewClient的配置函数
//
// 使用示例:
//
//	client := api.NewClient(
//	    api.WithRepoBaseURL("https://maven-mirror.mycompany.com/maven2"),
//	)
//
//	// 当下载制品时，将使用指定的镜像
//	jarData, err := client.DownloadJar(ctx, "org.apache.commons", "commons-lang3", "3.12.0")
func WithRepoBaseURL(repoBaseURL string) ClientOption {
	return func(c *Client) {
		c.repoBaseURL = repoBaseURL
	}
}

// WithHTTPClient 设置自定义HTTP客户端
//
// 该选项允许提供一个完全自定义的HTTP客户端，用于所有网络请求。
// 这在需要高级HTTP配置时特别有用，例如:
// - 自定义SSL/TLS设置
// - 特定的超时配置
// - 添加请求拦截器或中间件
// - 使用特殊的传输层实现
//
// 参数:
//   - httpClient: 一个配置好的http.Client实例
//
// 返回:
//   - ClientOption: 一个可以应用到NewClient的配置函数
//
// 使用示例:
//
//	// 创建自定义HTTP客户端
//	customTransport := &http.Transport{
//	    TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 警告：仅用于测试
//	    MaxIdleConns: 100,
//	    IdleConnTimeout: 90 * time.Second,
//	}
//	httpClient := &http.Client{
//	    Transport: customTransport,
//	    Timeout: 60 * time.Second,
//	}
//
//	// 使用自定义HTTP客户端创建SDK客户端
//	client := api.NewClient(
//	    api.WithHTTPClient(httpClient),
//	)
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithMaxRetries 设置最大重试次数
//
// 该选项配置客户端在遇到临时性错误（如网络故障、服务器过载等）时
// 自动重试的最大次数。合理的重试策略可以提高请求成功率，但过多的重试
// 可能会加剧服务器负载或导致客户端等待时间过长。
//
// 参数:
//   - maxRetries: 最大重试次数（不包括首次尝试）
//
// 返回:
//   - ClientOption: 一个可以应用到NewClient的配置函数
//
// 使用示例:
//
//	// 在不稳定网络环境中，可能需要更多重试
//	clientForUnstableNetwork := api.NewClient(
//	    api.WithMaxRetries(5),
//	)
//
//	// 或者在需要快速失败的场景中减少重试
//	clientForFailFast := api.NewClient(
//	    api.WithMaxRetries(1),
//	)
func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithRetryBackoff 设置重试间隔（毫秒）
//
// 该选项配置重试操作之间的初始等待时间（毫秒）。在实际重试过程中，
// 等待时间会根据指数退避策略增加，直到达到最大退避时间限制。
// 合理的退避时间有助于避免对服务器造成额外负担，并为其恢复提供时间。
//
// 参数:
//   - retryBackoffMs: 初始重试等待时间(毫秒)
//
// 返回:
//   - ClientOption: 一个可以应用到NewClient的配置函数
//
// 使用示例:
//
//	// 设置较短的初始退避时间，适合对延迟敏感的应用
//	client := api.NewClient(
//	    api.WithRetryBackoff(100), // 100毫秒初始等待
//	    api.WithMaxRetries(3),     // 配合重试次数使用
//	)
//
//	// 注意：实际等待时间会随重试次数增加
//	// 第一次重试：100ms
//	// 第二次重试：200ms（假设退避因子为2）
//	// 第三次重试：400ms
func WithRetryBackoff(retryBackoffMs int) ClientOption {
	return func(c *Client) {
		c.retryBackoffMs = retryBackoffMs
	}
}

// WithCache 设置缓存选项
//
// 该选项用于配置客户端的内存缓存功能。启用缓存可以减少对相同资源的重复请求，
// 提高应用性能并减轻服务器负担。缓存特别适合于搜索结果、元数据和不频繁变化的制品。
//
// 参数:
//   - enabled: 是否启用缓存(true/false)
//   - ttlSeconds: 缓存条目的生存时间(秒)，在此时间后缓存项将过期
//
// 返回:
//   - ClientOption: 一个可以应用到NewClient的配置函数
//
// 使用示例:
//
//	// 启用缓存，设置10分钟过期时间
//	client := api.NewClient(
//	    api.WithCache(true, 600),
//	)
//
//	// 查询相同资源时，在缓存有效期内将直接返回缓存结果而不发起网络请求
//	results1, _ := client.SearchByGroupAndArtifact(ctx, "org.apache.commons", "commons-lang3")
//	results2, _ := client.SearchByGroupAndArtifact(ctx, "org.apache.commons", "commons-lang3")
//	// results2将来自缓存，无网络请求延迟
func WithCache(enabled bool, ttlSeconds int) ClientOption {
	return func(c *Client) {
		c.cacheEnabled = enabled
		c.cacheTTLSeconds = ttlSeconds
	}
}

// NewClient 创建一个新的Sonatype Central客户端
//
// 该方法初始化一个配置完善的客户端实例，可通过可选参数自定义配置。
// 如果不提供任何选项，将使用以下默认值:
//   - baseURL: "https://search.maven.org" - 官方搜索API地址
//   - repoBaseURL: "https://repo1.maven.org/maven2" - 官方仓库地址
//   - httpClient: 30秒超时的标准HTTP客户端
//   - maxRetries: 3 - 失败时最多重试3次
//   - retryBackoffMs: 500 - 初始重试延迟500毫秒
//   - cacheEnabled: false - 默认不启用缓存
//   - cacheTTLSeconds: 300 - 缓存项有效期5分钟(如果启用)
//
// 参数:
//   - options: 可变数量的ClientOption函数，用于自定义客户端配置
//
// 返回:
//   - *Client: 配置完成并准备使用的客户端实例
//
// 使用示例:
//
//	// 基本用法 - 使用默认配置
//	client := api.NewClient()
//
//	// 高级用法 - 自定义配置
//	client := api.NewClient(
//	    api.WithBaseURL("https://custom-maven-mirror.com"),
//	    api.WithMaxRetries(5),
//	    api.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
//	    api.WithCache(true, 1800), // 启用缓存，30分钟过期
//	)
func NewClient(options ...ClientOption) *Client {
	// 默认配置
	client := &Client{
		baseURL:         "https://search.maven.org",
		repoBaseURL:     "https://repo1.maven.org/maven2",
		httpClient:      &http.Client{Timeout: 30 * time.Second},
		maxRetries:      3,
		retryBackoffMs:  500,
		cacheEnabled:    false,
		cacheTTLSeconds: 300, // 5分钟
	}

	// 应用自定义选项
	for _, option := range options {
		option(client)
	}

	return client
}

// GetBaseURL 获取当前使用的基础URL
//
// 该方法返回客户端当前配置的搜索API基础URL。这对于调试、日志记录或需要了解
// 当前客户端配置的场景很有用。基础URL是在创建客户端时通过WithBaseURL选项设置的，
// 如果未指定，则使用默认值"https://search.maven.org"。
//
// 返回:
//   - string: 当前配置的搜索API基础URL
//
// 使用示例:
//
//	client := api.NewClient(
//	    api.WithBaseURL("https://custom-maven-repo.example.com"),
//	)
//
//	// 获取并记录基础URL
//	baseURL := client.GetBaseURL()
//	fmt.Printf("当前使用的Maven搜索API: %s\n", baseURL)
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// GetRepoBaseURL 获取当前使用的仓库基础URL
//
// 该方法返回客户端当前配置的下载仓库基础URL。下载URL与搜索API URL通常不同，
// 它被用于构建完整的制品下载路径。该URL是在创建客户端时通过WithRepoBaseURL选项设置的，
// 如果未指定，则使用默认值"https://repo1.maven.org/maven2"。
//
// 返回:
//   - string: 当前配置的下载仓库基础URL
//
// 使用示例:
//
//	client := api.NewClient(
//	    api.WithRepoBaseURL("https://maven-mirror.example.com/maven2"),
//	)
//
//	// 获取并使用仓库URL
//	repoURL := client.GetRepoBaseURL()
//	fmt.Printf("制品将从 %s 下载\n", repoURL)
//
//	// 也可以用于手动构建完整下载URL
//	fullDownloadURL := repoURL + "/com/example/library/1.0.0/library-1.0.0.jar"
func (c *Client) GetRepoBaseURL() string {
	return c.repoBaseURL
}
