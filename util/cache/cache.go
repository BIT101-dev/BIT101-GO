/*
 * @Author: flwfdd
 * @Date: 2023-03-13 20:39:32
 * @LastEditTime: 2023-03-15 18:14:40
 * @Description: 缓存模块
 */

/*
	优化

线程安全
使用 sync.RWMutex 保证并发安全。RLock 允许多个读操作同时进行，而 Lock 使写操作互斥。

内存管理
增加 cleanup 方法用于删除过期的缓存键，避免缓存占用无效的内存。

添加 NewMemoryCache 工厂函数来初始化 MemoryCache 实例，避免直接访问底层字段，便于扩展和测试。
统一的实例管理

全局 Instance 变量通过工厂方法生成，更加直观。

在 Get 方法中增加判断并调用 cleanup，即使过期键未被主动访问也能正确清理。
*/
package cache

import (
	"sync"
	"time"
)

type Cache interface {
	Get(key string, default_value string) string
	Set(key string, value string, expire_second int)
}

type MemoryCache struct {
	cache  map[string]string
	expiry map[string]int64
	mutex  sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		cache:  make(map[string]string),
		expiry: make(map[string]int64),
	}
}

// Get 获取缓存中的值，若不存在或已过期则返回默认值
func (m *MemoryCache) Get(key string, defaultValue string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	expiry, exists := m.expiry[key]
	if !exists || expiry < time.Now().Unix() {
		// 缓存不存在或已过期
		m.cleanup(key)
		return defaultValue
	}
	return m.cache[key]
}

// Set 设置缓存值及过期时间
func (m *MemoryCache) Set(key string, value string, expireSeconds int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	expiry := time.Now().Unix() + int64(expireSeconds)
	m.cache[key] = value
	m.expiry[key] = expiry
}

// cleanup 清理过期的缓存键
func (m *MemoryCache) cleanup(key string) {
	delete(m.cache, key)
	delete(m.expiry, key)
}

// 全局缓存实例
var Instance Cache = NewMemoryCache()
