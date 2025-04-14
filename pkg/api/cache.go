package api

import (
	"sync"
	"time"
)

// cacheItem 缓存条目
//
// 表示存储在内存缓存中的单个数据项。每个缓存项包含实际数据内容和过期时间。
// 当访问缓存项时，会检查当前时间是否已超过过期时间，以确定该项是否仍然有效。
//
// 字段说明:
//   - data: 缓存的二进制数据内容，通常是API响应的正文
//   - expiration: 过期时间点，表示该缓存项在何时应被视为无效
type cacheItem struct {
	data       []byte
	expiration time.Time
}

// memCache 内存缓存
//
// 提供简单的内存中键值存储，用于缓存下载的文件和API响应，以减少网络请求。
// 该实现使用标准的Go sync.RWMutex来保证并发安全，允许多个goroutine同时
// 读取缓存，但写入操作会阻塞所有其他访问。
//
// 缓存使用惰性过期检查策略，即只有在尝试访问某个缓存项时才检查它是否过期，
// 而不是主动清理过期项。这种方法简化了实现，但可能导致过期项在内存中长期存在，
// 直到被再次访问或整个缓存被清除。
//
// 字段说明:
//   - entries: 存储缓存项的映射，键为缓存标识符，值为包含数据和过期时间的结构
//   - mutex: 读写锁，用于保证并发安全
type memCache struct {
	entries map[string]cacheItem
	mutex   sync.RWMutex
}

// 创建全局缓存实例
//
// globalCache是包级别的单例缓存实例，被整个SDK共享使用。这种设计允许
// 不同的Client实例共享同一个缓存，从而最大化缓存命中率和资源复用。
//
// 全局缓存主要用于存储下载的文件内容、搜索结果和其他API响应，避免重复请求
// 相同的资源。缓存内容的生命周期由各个Client实例的cacheTTLSeconds参数控制。
//
// 注意事项:
//   - 缓存仅在内存中，应用程序重启后会丢失
//   - 在使用大量数据或长时间运行的应用中，应定期调用ClearCache()以释放内存
//   - 缓存是进程内的，不在多个应用实例间共享
var globalCache = &memCache{
	entries: make(map[string]cacheItem),
}

// ClearCache 清除客户端的内存缓存
//
// 该方法用于立即清空全局内存缓存中的所有条目。这在需要强制刷新缓存或释放内存时非常有用，
// 例如当怀疑缓存数据已过时或在测试环境中需要确保每次请求都获取最新数据时。
// 清除操作使用互斥锁保证并发安全。
//
// 使用示例:
//
//	client := api.NewClient()
//
//	// 执行一些使用缓存的操作...
//
//	// 然后清除缓存以确保获取最新数据
//	client.ClearCache()
//
//	// 后续请求将获取新的数据而不是使用缓存
func (c *Client) ClearCache() {
	globalCache.mutex.Lock()
	defer globalCache.mutex.Unlock()

	// 重置缓存
	globalCache.entries = make(map[string]cacheItem)
}

// 从缓存中获取内容
//
// 该方法用于从全局内存缓存中检索数据。它会先检查指定键的数据是否存在，然后验证数据是否已过期。
// 如果数据存在且未过期，则返回缓存的内容；否则表示缓存未命中。整个过程使用读锁保证并发安全。
//
// 参数:
//   - key: 缓存的键，通常由请求路径或资源标识符构成
//
// 返回:
//   - []byte: 缓存的数据内容，当缓存命中且未过期时返回
//   - bool: 如果缓存命中且未过期返回true，否则返回false
func getFromCache(key string) ([]byte, bool) {
	globalCache.mutex.RLock()
	defer globalCache.mutex.RUnlock()

	entry, exists := globalCache.entries[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.data, true
}

// 添加内容到缓存
//
// 该方法用于将数据存储到全局内存缓存中，并设置其过期时间。如果指定的TTL小于或等于0，
// 则不会进行缓存。缓存操作使用互斥锁保证并发安全。缓存项包括数据内容和基于当前时间
// 计算的过期时间点。
//
// 参数:
//   - key: 缓存的键，用于后续检索数据
//   - data: 要缓存的数据内容
//   - ttlSeconds: 生存时间(秒)，指定数据在缓存中保留的时长
//
// 实现细节:
//   - 数据以byte数组形式存储，适合缓存二进制内容如下载的文件
//   - 过期机制基于时间戳比较，而非定时删除
//   - 在访问时才检查过期，没有主动清理机制
func addToCache(key string, data []byte, ttlSeconds int) {
	if ttlSeconds <= 0 {
		return
	}

	globalCache.mutex.Lock()
	defer globalCache.mutex.Unlock()

	// 添加带过期时间的条目
	globalCache.entries[key] = cacheItem{
		data:       data,
		expiration: time.Now().Add(time.Duration(ttlSeconds) * time.Second),
	}
}

// IsCacheEnabled 判断客户端是否启用了缓存
//
// 该方法返回当前客户端的缓存状态。当缓存启用时，客户端会尝试从内存缓存获取响应，
// 减少网络请求；当缓存禁用时，每次请求都会直接访问网络。
//
// 返回:
//   - bool: 如果缓存已启用，返回true；否则返回false
//
// 使用示例:
//
//	client := api.NewClient()
//
//	if client.IsCacheEnabled() {
//	    fmt.Println("客户端缓存已启用，将优先使用缓存数据")
//	} else {
//	    fmt.Println("客户端缓存已禁用，每次请求都将访问网络")
//	}
func (c *Client) IsCacheEnabled() bool {
	return c.cacheEnabled
}

// GetCacheTTL 获取缓存条目的生存时间(TTL)
//
// 该方法返回当前客户端设置的缓存条目生存时间（以秒为单位）。所有新添加到缓存中的条目
// 都会使用这个TTL值设置过期时间。此设置仅影响新添加的缓存项，不会改变已存在的缓存项的TTL。
//
// 返回:
//   - int: 缓存条目的生存时间(秒)
//
// 使用示例:
//
//	client := api.NewClient()
//
//	ttl := client.GetCacheTTL()
//	fmt.Printf("当前缓存TTL设置为%d秒\n", ttl)
func (c *Client) GetCacheTTL() int {
	return c.cacheTTLSeconds
}

// SetCacheTTL 设置缓存条目的生存时间(TTL)
//
// 该方法用于调整缓存条目在内存中保留的时间长度（以秒为单位）。较短的TTL会使缓存更快过期，
// 更频繁地从网络获取最新数据；较长的TTL可以减少网络请求，但可能导致使用过时的数据。
// 该设置仅影响新添加的缓存项，不会改变已存在缓存项的过期时间。
//
// 参数:
//   - seconds: 缓存条目的生存时间(秒)，如果设为0或负数，新的缓存项将不会被存储
//
// 使用示例:
//
//	client := api.NewClient()
//
//	// 设置较短的TTL，确保数据不会太旧
//	client.SetCacheTTL(60) // 缓存项在60秒后过期
//
//	// 或设置较长的TTL，减少网络请求
//	client.SetCacheTTL(3600) // 缓存项在1小时后过期
func (c *Client) SetCacheTTL(seconds int) {
	c.cacheTTLSeconds = seconds
}

// EnableCache 启用客户端缓存
//
// 该方法用于启用客户端的内存缓存功能。启用缓存后，客户端会尝试从缓存中获取之前
// 请求过的数据，从而减少网络请求，提高性能和响应速度。对于频繁访问相同资源的场景，
// 启用缓存可以显著减少API请求次数和网络带宽消耗。
//
// 使用示例:
//
//	client := api.NewClient()
//
//	// 默认可能缓存已启用，如果要确保启用
//	client.EnableCache()
//
//	// 设置适当的TTL
//	client.SetCacheTTL(300) // 5分钟
//
//	// 现在发起的请求会使用缓存
func (c *Client) EnableCache() {
	c.cacheEnabled = true
}

// DisableCache 禁用客户端缓存
//
// 该方法用于禁用客户端的内存缓存功能。禁用缓存后，每次请求都会直接访问网络获取最新数据，
// 而不会使用任何缓存的响应。这在需要确保始终获取最新数据的场景下非常有用，例如在开发或测试环境中，
// 或者当数据变化频繁且实时性要求高的应用场景。
//
// 使用示例:
//
//	client := api.NewClient()
//
//	// 禁用缓存以确保获取最新数据
//	client.DisableCache()
//
//	// 现在发起的请求将始终访问网络而不使用缓存
func (c *Client) DisableCache() {
	c.cacheEnabled = false
}
