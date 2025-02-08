/*
 * @Author: flwfdd
 * @Date: 2025-02-08 15:39:40
 * @LastEditTime: 2025-02-08 16:02:35
 * @Description: _(:з」∠)_
 */
package cache

import (
	"BIT101-GO/util/config"
	"context"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var Context = context.Background()

func Init() {
	// 初始化redis
	rdb := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	RDB = rdb
}
