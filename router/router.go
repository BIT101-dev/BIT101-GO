/*
 * @Author: flwfdd
 * @Date: 2023-03-13 10:39:47
 * @LastEditTime: 2023-03-21 08:56:30
 * @Description: 路由配置
 */
package router

import (
	"github.com/gin-gonic/gin"

	"BIT101-GO/controller"
	"BIT101-GO/middleware"
)

// 配置路由
func SetRouter(router *gin.Engine) {
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "Hello BIT101!"})
	})
	// 用户模块
	user := router.Group("/user")
	{
		user.GET("/check", middleware.CheckLogin(true))
		user.POST("/login", controller.UserLogin)
		user.POST("/webvpn_verify_init", controller.UserWebvpnVerifyInit)
		user.POST("/webvpn_verify", controller.UserWebvpnVerify)
		user.POST("/mail_verify", controller.UserMailVerify)
		user.POST("/register", controller.UserRegister)
		user.GET("/info", middleware.CheckLogin(false), controller.UserGetInfo)
		user.POST("/info", middleware.CheckLogin(true), controller.UserSetInfo)
	}
	// 成绩模块
	score := router.Group("/score")
	{
		score.GET("", controller.Score)
		score.GET("/report", controller.Report)
	}
	// 上传模块
	upload := router.Group("/upload")
	{
		upload.POST("/image", middleware.CheckLogin(true), controller.ImageUpload)
		upload.POST("/image/url", middleware.CheckLogin(true), controller.ImageUploadByUrl)
	}
}
