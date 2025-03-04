/*
 * @Author: flwfdd
 * @Date: 2023-03-13 10:39:47
 * @LastEditTime: 2023-09-24 00:20:52
 * @Description: 路由配置
 */
package api

import (
	"BIT101-GO/api/middleware"
	"BIT101-GO/config"
	"BIT101-GO/service"
	"time"

	"github.com/gin-contrib/cors"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
)

// 配置路由
func RegisterRouter(router *gin.Engine, cfg *config.Config) {
	router.Use(limits.RequestSizeLimiter(cfg.Saver.MaxSize << 20))
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Content-Type", "fake-cookie", "webvpn-cookie"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		MaxAge:       12 * time.Hour,
	}))

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "Hello BIT101!"})
	})
	// 用户模块
	user := router.Group("/user")
	{
		user.GET("/check", middleware.CheckLogin(true))
		user.POST("/login", service.UserLogin)
		user.POST("/webvpn_verify_init", service.UserWebvpnVerifyInit)
		user.POST("/webvpn_verify", service.UserWebvpnVerify)
		user.POST("/mail_verify", service.UserMailVerify)
		user.POST("/register", service.UserRegister)
		user.GET("/info", middleware.CheckLogin(false), service.OldUserGetInfo)
		user.GET("/info/:id", middleware.CheckLogin(false), service.UserGetInfo)
		user.PUT("/info", middleware.CheckLogin(true), service.UserSetInfo)
		user.POST("/follow/:id", middleware.CheckLogin(true), service.FollowPost)
		user.GET("/followings", middleware.CheckLogin(true), service.FollowListGet)
		user.GET("/followers", middleware.CheckLogin(true), service.FansListGet)
	}
	// 成绩模块
	score := router.Group("/score")
	{
		score.GET("", middleware.Proxy(), service.Score)
		score.GET("/report", middleware.Proxy(), service.Report)
	}
	// 上传模块
	upload := router.Group("/upload")
	{
		upload.POST("/image", middleware.CheckLogin(true), service.ImageUpload)
		upload.POST("/image/url", middleware.CheckLogin(true), service.ImageUploadByUrl)
	}
	// 帖子模块
	poster := router.Group("/posters")
	{
		poster.GET("/:id", middleware.CheckLogin(true), service.PosterGet)
		poster.GET("", middleware.CheckLogin(true), service.PostList)
		poster.POST("", middleware.CheckLogin(true), service.PosterPost)
		poster.PUT("/:id", middleware.CheckLogin(true), service.PosterPut)
		poster.DELETE("/:id", middleware.CheckLogin(true), service.PosterDelete)
		poster.GET("/claims", service.ClaimList)
	}
	// 文章模块
	paper := router.Group("/papers")
	{
		paper.GET("/:id", middleware.CheckLogin(false), service.PaperGet)
		paper.GET("", service.PaperList)
		paper.POST("", middleware.CheckLogin(true), service.PaperPost)
		paper.PUT("/:id", middleware.CheckLogin(true), service.PaperPut)
		paper.DELETE("/:id", middleware.CheckLogin(true), service.PaperDelete)
	}
	// 操作反馈模块
	reaction := router.Group("/reaction")
	{
		reaction.POST("/like", middleware.CheckLogin(true), service.ReactionLike)
		reaction.POST("/comments", middleware.CheckLogin(true), service.ReactionComment)
		reaction.GET("/comments", middleware.CheckLogin(false), service.ReactionCommentList)
		reaction.DELETE("/comments/:id", middleware.CheckLogin(true), service.ReactionCommentDelete)
		reaction.POST("/stay", middleware.CheckLogin(true), service.ReactionStay)
	}
	// 课程模块
	course := router.Group("/courses")
	{
		course.GET("", middleware.CheckLogin(true), service.CourseList)
		course.GET("/:id", middleware.CheckLogin(true), service.CourseInfo)
		course.GET("/upload/url", middleware.CheckLogin(true), service.CourseUploadUrl)
		course.POST("/upload/log", middleware.CheckLogin(true), service.CourseUploadLog)
		course.GET("/schedule", service.CourseSchedule)
		course.GET("/histories/:number", service.CourseHistory)
	}
	// 变量模块
	variable := router.Group("/variables")
	{
		variable.GET("", service.VariableGet)
		variable.POST("", middleware.CheckLogin(true), middleware.CheckSuper(), service.VariablePost)
	}
	// 消息模块
	message := router.Group("/messages")
	{
		message.GET("", middleware.CheckLogin(true), service.MessageGetList)
		message.GET("/unread_num", middleware.CheckLogin(true), service.MessageGetUnreadNum)
		message.GET("/unread_nums", middleware.CheckLogin(true), service.MessageGetUnreadNums)
		message.POST("/system", middleware.CheckLogin(true), middleware.CheckAdmin(), service.SystemMessagePost)
		message.GET("/push", middleware.CheckLogin(true), service.PushMessageRequestKey)
		message.POST("/push", middleware.CheckLogin(true), service.PushMessageSubscribe)
		message.DELETE("/push", middleware.CheckLogin(true), service.PushMessageUnsubscribe)
	}
	// 治理模块
	manage := router.Group("/manage")
	{
		manage.GET("/report_types", service.ReportTypeListGet)
		manage.POST("/reports", middleware.CheckLogin(true), service.ReportPost)
		manage.GET("/reports", middleware.CheckLogin(true), middleware.CheckAdmin(), service.ReportList)
		manage.PUT("reports/:id", middleware.CheckLogin(true), middleware.CheckAdmin(), service.ReportPut)
		manage.POST("/bans", middleware.CheckLogin(true), middleware.CheckAdmin(), service.BanPost)
		manage.GET("/bans", middleware.CheckLogin(true), middleware.CheckAdmin(), service.BanList)
	}
}
