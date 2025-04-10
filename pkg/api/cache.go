package api

import (
	"sync"
	"time"
)

// cacheItem 缓存条目
type cacheItem struct {
	data       []byte
	expiration time.Time
}

// memCache 内存缓存
type memCache struct {
	entries map[string]cacheItem
	mutex   sync.RWMutex
}

// 创建全局缓存实例
var globalCache = &memCache{
	entries: make(map[string]cacheItem),
}

// ClearCache 清除缓存
func (c *Client) ClearCache() {
	globalCache.mutex.Lock()
	defer globalCache.mutex.Unlock()

	// 重置缓存
	globalCache.entries = make(map[string]cacheItem)
}

// 从缓存中获取内容
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
func (c *Client) IsCacheEnabled() bool {
	return c.cacheEnabled
}

// GetCacheTTL 获取缓存TTL设置（秒）
func (c *Client) GetCacheTTL() int {
	return c.cacheTTLSeconds
}

// SetCacheTTL 设置缓存TTL
func (c *Client) SetCacheTTL(seconds int) {
	c.cacheTTLSeconds = seconds
}

// EnableCache 启用缓存
func (c *Client) EnableCache() {
	c.cacheEnabled = true
}

// DisableCache 禁用缓存
func (c *Client) DisableCache() {
	c.cacheEnabled = false
}
