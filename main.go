/*
 * @Author: flwfdd
 * @Date: 2023-03-13 10:20:13
 * @LastEditTime: 2023-03-15 16:53:51
 * @Description: _(:з」∠)_
 */
package main

import (
	"BIT101-GO/router"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// webvpn.Test()
	app := gin.Default()
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://127.0.0.1:3000"},
		AllowHeaders: []string{"Content-Type", "fake-cookie"},
		// ExposeHeaders:    []string{"Content-Length"},
		// AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// 	return true
		// },
		MaxAge: 12 * time.Hour,
	}))
	router.SetRouter(app)
	app.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
