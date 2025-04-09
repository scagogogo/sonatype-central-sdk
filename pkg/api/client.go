package api

import (
	"net/http"
	"time"
)

// ClientOption 客户端配置选项函数
type ClientOption func(*Client)

// Client Sonatype Central 客户端
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
func WithProxy(proxy string) ClientOption {
	return func(c *Client) {
		c.proxy = proxy
	}
}

// WithBaseURL 设置自定义基础URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithRepoBaseURL 设置自定义仓库基础URL
func WithRepoBaseURL(repoBaseURL string) ClientOption {
	return func(c *Client) {
		c.repoBaseURL = repoBaseURL
	}
}

// WithHTTPClient 设置自定义HTTP客户端
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithRetryBackoff 设置重试间隔（毫秒）
func WithRetryBackoff(retryBackoffMs int) ClientOption {
	return func(c *Client) {
		c.retryBackoffMs = retryBackoffMs
	}
}

// WithCache 设置缓存选项
func WithCache(enabled bool, ttlSeconds int) ClientOption {
	return func(c *Client) {
		c.cacheEnabled = enabled
		c.cacheTTLSeconds = ttlSeconds
	}
}

// NewClient 创建一个新的Sonatype Central客户端
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
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// GetRepoBaseURL 获取当前使用的仓库基础URL
func (c *Client) GetRepoBaseURL() string {
	return c.repoBaseURL
}
