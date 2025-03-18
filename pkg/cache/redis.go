/*
 * @Author: flwfdd
 * @Date: 2025-02-08 15:39:40
 * @LastEditTime: 2025-03-18 14:31:56
 * @Description: _(:з」∠)_
 */
package cache

import (
	"BIT101-GO/config"
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	Context = context.Background()
	rdb     *redis.Client
	once    sync.Once
)

func RDB() *redis.Client {
	once.Do(func() {
		rdb = redis.NewClient(&redis.Options{
			Addr: config.Get().Redis.Addr,
		})
	})
	return rdb
}
