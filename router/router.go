/*
 * @Author: flwfdd
 * @Date: 2023-03-13 10:39:47
 * @LastEditTime: 2023-09-24 00:20:52
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
		user.PUT("/info", middleware.CheckLogin(true), controller.UserSetInfo)
		user.POST("/follow", middleware.CheckLogin(true), controller.ReactionFollow)
		user.GET("/followings", middleware.CheckLogin(true), controller.GetFollowList)
		user.GET("/followers", middleware.CheckLogin(true), controller.GetFansList)
	}
	// 成绩模块
	score := router.Group("/score")
	{
		score.GET("", middleware.Proxy(), controller.Score)
		score.GET("/report", middleware.Proxy(), controller.Report)
	}
	// 上传模块
	upload := router.Group("/upload")
	{
		upload.POST("/image", middleware.CheckLogin(true), controller.ImageUpload)
		upload.POST("/image/url", middleware.CheckLogin(true), controller.ImageUploadByUrl)
	}
	// 帖子模块
	post := router.Group("/posters")
	{
		post.GET("/:id", middleware.CheckLogin(true), controller.PostGet)
		post.GET("", middleware.CheckLogin(true), controller.PostList)
		post.POST("", middleware.CheckLogin(true), controller.PostSubmit)
		post.PUT("/:id", middleware.CheckLogin(true), controller.PostPut)
		post.DELETE("/:id", middleware.CheckLogin(true), controller.PostDelete)
		post.GET("/claims", controller.ClaimList)
	}
	// 文章模块
	paper := router.Group("/papers")
	{
		paper.GET("/:id", middleware.CheckLogin(false), controller.PaperGet)
		paper.GET("", controller.PaperList)
		paper.POST("", middleware.CheckLogin(true), controller.PaperPost)
		paper.PUT("/:id", middleware.CheckLogin(true), controller.PaperPut)
		paper.DELETE("/:id", middleware.CheckLogin(true), controller.PaperDelete)
	}
	// 操作反馈模块
	reaction := router.Group("/reaction")
	{
		reaction.POST("/like", middleware.CheckLogin(true), controller.ReactionLike)
		reaction.POST("/comments", middleware.CheckLogin(true), controller.ReactionComment)
		reaction.GET("/comments", middleware.CheckLogin(false), controller.ReactionCommentList)
		reaction.DELETE("/comments/:id", middleware.CheckLogin(true), controller.ReactionCommentDelete)
	}
	// 课程模块
	course := router.Group("/courses")
	{
		course.GET("", controller.CourseList)
		course.GET("/:id", middleware.CheckLogin(false), controller.CourseInfo)
		course.GET("/upload/url", middleware.CheckLogin(true), controller.CourseUploadUrl)
		course.POST("/upload/log", middleware.CheckLogin(true), controller.CourseUploadLog)
		course.GET("/schedule", controller.CourseSchedule)
		course.GET("/histories/:number", controller.CourseHistory)
	}
	// 变量模块
	variable := router.Group("/variables")
	{
		variable.GET("", controller.VariableGet)
		variable.POST("", middleware.CheckLogin(true), controller.VariablePost)
	}
	// 消息模块
	message := router.Group("/messages")
	{
		message.GET("", middleware.CheckLogin(true), controller.MessageGetList)
		message.GET("/unread_num", middleware.CheckLogin(true), controller.MessageGetUnreadNum)
		message.GET("/unread_likes_num", middleware.CheckLogin(true), controller.MessageGetUnreadLikeNum)
		message.GET("/unread_comments_num", middleware.CheckLogin(true), controller.MessageGetUnreadCommentNum)
	}
}
