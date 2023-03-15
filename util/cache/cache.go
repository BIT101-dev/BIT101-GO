/*
 * @Author: flwfdd
 * @Date: 2023-03-13 20:39:32
 * @LastEditTime: 2023-03-15 18:14:40
 * @Description: 缓存模块
 */
package cache

import "time"

type Cache interface {
	Get(key string, default_value string) string
	Set(key string, value string, expire_second int)
}

type MemoryCache struct {
	cache map[string]string
	time  map[string]int64
}

func (m *MemoryCache) Get(key string, default_value string) string {
	if value, ok := m.cache[key]; ok {
		if m.time[key] < time.Now().Unix() {
			delete(m.cache, key)
			delete(m.time, key)
			return default_value
		}
		return value
	}
	return default_value
}

func (m *MemoryCache) Set(key string, value string, expire_second int) {
	m.time[key] = time.Now().Unix() + int64(expire_second)
	m.cache[key] = value
}

var cacheInstance = &MemoryCache{map[string]string{}, map[string]int64{}}

var Instance Cache = cacheInstance
