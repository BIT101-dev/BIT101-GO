/*
 * @Author: flwfdd
 * @Date: 2023-03-13 10:39:47
 * @LastEditTime: 2023-03-18 11:54:19
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
	// 用户模块
	user := router.Group("/user")
	{
		// user.POST("/check", controller.UserCheck)
		user.POST("/login", controller.UserLogin)
		user.POST("/webvpn_verify_init", controller.UserWebvpnVerifyInit)
		user.POST("/webvpn_verify", controller.UserWebvpnVerify)
	}
	// 成绩模块
	score := router.Group("/score")
	{
		score.GET("", controller.Score)
		score.GET("/report", controller.Report)
	}
}
