/*
 * @Author: flwfdd
 * @Date: 2023-03-13 10:39:47
 * @LastEditTime: 2023-05-17 17:24:48
 * @Description: 路由配置
 */
package router

import (
	"github.com/gin-gonic/gin"

	"BIT101-GO/controller"
)

// 配置路由
func SetRouter(router *gin.Engine) {
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "Hello BIT101!"})
	})
	// 成绩模块
	score := router.Group("/score")
	{
		score.GET("", controller.Score)
		score.GET("/report", controller.Report)
	}
}
