/*
 * @Author: flwfdd
 * @Date: 2023-03-13 10:39:47
 * @LastEditTime: 2025-03-18 01:04:05
 * @Description: 路由配置
 */
package api

import (
	"BIT101-GO/api/handler"
	"BIT101-GO/api/middleware"
	"BIT101-GO/api/service"
	"BIT101-GO/config"
	"time"

	"github.com/gin-contrib/cors"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

func SetupContainer() {
	// 注册配置
	do.Provide(nil, func(do.Injector) (*config.Config, error) {
		return config.Get(), nil
	})

	// 注册变量服务
	do.Provide(nil, func(do.Injector) (*service.VariableService, error) {
		return service.NewVariableService(), nil
	})

	// 注册图片服务
	do.Provide(nil, func(do.Injector) (*service.ImageService, error) {
		return service.NewImageService(), nil
	})

	// 解决用户和消息服务的循环依赖
	// 注册消息服务(初始为空)
	do.Provide(nil, func(do.Injector) (*service.MessageService, error) {
		return service.NewMessageService(nil), nil
	})

	// 注册用户服务
	do.Provide(nil, func(i do.Injector) (*service.UserService, error) {
		imageSvc := do.MustInvoke[*service.ImageService](i)
		messageSvc := do.MustInvoke[*service.MessageService](i)
		return service.NewUserService(imageSvc, messageSvc), nil
	})

	// 更新消息服务中的用户服务引用
	userSvc := do.MustInvoke[*service.UserService](nil)
	messageSvc := do.MustInvoke[*service.MessageService](nil)
	messageSvc.SetUserService(userSvc) // 更新消息服务中的用户服务引用

	// 注册管理服务
	do.Provide(nil, func(do.Injector) (*service.ManageService, error) {
		userSvc := do.MustInvoke[*service.UserService](nil)
		return service.NewManageService(userSvc), nil
	})

	// 注册Gorse服务
	do.Provide(nil, func(do.Injector) (*service.GorseService, error) {
		return service.NewGorseService(), nil
	})

	// 注册Meilisearch服务
	do.Provide(nil, func(do.Injector) (*service.MeilisearchService, error) {
		return service.NewMeilisearchService(), nil
	})

	// 注册交互服务
	do.Provide(nil, func(do.Injector) (*service.ReactionService, error) {
		userSvc := do.MustInvoke[*service.UserService](nil)
		imageSvc := do.MustInvoke[*service.ImageService](nil)
		messageSvc := do.MustInvoke[*service.MessageService](nil)
		gorseSvc := do.MustInvoke[*service.GorseService](nil)
		return service.NewReactionService(userSvc, imageSvc, messageSvc, gorseSvc), nil
	})

	// 注册课程服务
	do.Provide(nil, func(do.Injector) (*service.CourseService, error) {
		reactionSvc := do.MustInvoke[*service.ReactionService](nil)
		meilisearchSvc := do.MustInvoke[*service.MeilisearchService](nil)
		return service.NewCourseService(reactionSvc, meilisearchSvc), nil
	})

	// 注册文章服务
	do.Provide(nil, func(do.Injector) (*service.PaperService, error) {
		userSvc := do.MustInvoke[*service.UserService](nil)
		reactionSvc := do.MustInvoke[*service.ReactionService](nil)
		meilisearchSvc := do.MustInvoke[*service.MeilisearchService](nil)
		return service.NewPaperService(userSvc, reactionSvc, meilisearchSvc), nil
	})

	// 注册帖子服务
	do.Provide(nil, func(do.Injector) (*service.PosterService, error) {
		userSvc := do.MustInvoke[*service.UserService](nil)
		imageSvc := do.MustInvoke[*service.ImageService](nil)
		reactionSvc := do.MustInvoke[*service.ReactionService](nil)
		messageSvc := do.MustInvoke[*service.MessageService](nil)
		gorseSvc := do.MustInvoke[*service.GorseService](nil)
		meilisearchSvc := do.MustInvoke[*service.MeilisearchService](nil)
		return service.NewPosterService(userSvc, imageSvc, reactionSvc, messageSvc, gorseSvc, meilisearchSvc), nil
	})
}

// 配置路由
func RegisterRouter(router *gin.Engine) {
	cfg := do.MustInvoke[*config.Config](nil)
	router.Use(middleware.PrometheusMiddleware())
	router.Use(limits.RequestSizeLimiter(cfg.Saver.MaxSize << 20))
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Content-Type", "fake-cookie", "webvpn-cookie"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		MaxAge:       12 * time.Hour,
	}))

	// 添加Prometheus指标路由
	router.GET("/metrics", middleware.PrometheusHandler())

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "Hello BIT101!"})
	})
	// 用户模块
	userSvc := do.MustInvoke[*service.UserService](nil)
	userHandler := handler.NewUserHandler(userSvc)
	userRouter := router.Group("/user")
	{
		userRouter.GET("/check", middleware.CheckLogin(true))
		userRouter.POST("/login", userHandler.LoginHandler)
		userRouter.POST("/webvpn_verify_init", userHandler.WebvpnVerifyInitHandler)
		userRouter.POST("/webvpn_verify", userHandler.WebvpnVerifyHandler)
		userRouter.POST("/mail_verify", userHandler.MailVerifyHandler)
		userRouter.POST("/register", userHandler.RegisterHandler)
		userRouter.GET("/info", middleware.CheckLogin(false), userHandler.OldUserGetInfo)
		userRouter.GET("/info/:id", middleware.CheckLogin(false), userHandler.GetInfoHandler)
		userRouter.PUT("/info", middleware.CheckLogin(true), userHandler.SetInfoHandler)
		userRouter.POST("/follow/:id", middleware.CheckLogin(true), userHandler.FollowHandler)
		userRouter.GET("/followings", middleware.CheckLogin(true), userHandler.GetFollowListHandler)
		userRouter.GET("/followers", middleware.CheckLogin(true), userHandler.GetFansListHandler)
	}
	// 成绩模块
	scoreHandler := handler.NewScoreHandler()
	scoreRouter := router.Group("/score")
	{
		scoreRouter.GET("", middleware.Proxy(), scoreHandler.GetScoreHandler)
		scoreRouter.GET("/report", middleware.Proxy(), scoreHandler.GetReportHandler)
	}
	// 上传模块
	imageSvc := do.MustInvoke[*service.ImageService](nil)
	imageHandler := handler.NewImageHandler(imageSvc)
	uploadRouter := router.Group("/upload")
	{
		uploadRouter.POST("/image", middleware.CheckLogin(true), imageHandler.UploadHandler)
		uploadRouter.POST("/image/url", middleware.CheckLogin(true), imageHandler.UploadByUrlHandler)
	}
	// 帖子模块
	posterSvc := do.MustInvoke[*service.PosterService](nil)
	posterHandler := handler.NewPosterHandler(posterSvc)
	posterRouter := router.Group("/posters")
	{
		posterRouter.GET("/:id", middleware.CheckLogin(true), posterHandler.GetHandler)
		posterRouter.GET("", middleware.CheckLogin(true), posterHandler.GetListHandler)
		posterRouter.POST("", middleware.CheckLogin(true), posterHandler.CreateHandler)
		posterRouter.PUT("/:id", middleware.CheckLogin(true), posterHandler.EditHandler)
		posterRouter.DELETE("/:id", middleware.CheckLogin(true), posterHandler.DeleteHandler)
		posterRouter.GET("/claims", posterHandler.GetClaimsHandler)
	}
	// 文章模块
	paperSvc := do.MustInvoke[*service.PaperService](nil)
	paperHandler := handler.NewPaperHandler(paperSvc)
	paperRouter := router.Group("/papers")
	{
		paperRouter.GET("/:id", middleware.CheckLogin(false), paperHandler.GetHandler)
		paperRouter.GET("", paperHandler.GetListHandler)
		paperRouter.POST("", middleware.CheckLogin(true), paperHandler.CreateHandler)
		paperRouter.PUT("/:id", middleware.CheckLogin(true), paperHandler.EditHandler)
		paperRouter.DELETE("/:id", middleware.CheckLogin(true), paperHandler.DeleteHandler)
	}
	// 交互反馈模块
	reactionSvc := do.MustInvoke[*service.ReactionService](nil)
	reactionHandler := handler.NewReactionHandler(reactionSvc)
	reactionRouter := router.Group("/reaction")
	{
		reactionRouter.POST("/like", middleware.CheckLogin(true), reactionHandler.LikeHandler)
		reactionRouter.POST("/comments", middleware.CheckLogin(true), reactionHandler.CommentHandler)
		reactionRouter.GET("/comments", middleware.CheckLogin(false), reactionHandler.GetCommentsHandler)
		reactionRouter.DELETE("/comments/:id", middleware.CheckLogin(true), reactionHandler.DeleteCommentHandler)
		reactionRouter.POST("/stay", middleware.CheckLogin(true), reactionHandler.StayHandler)
	}
	// 课程模块
	courseSvc := do.MustInvoke[*service.CourseService](nil)
	courseHandler := handler.NewCourseHandler(courseSvc)
	courseRouter := router.Group("/courses")
	{
		courseRouter.GET("", middleware.CheckLogin(true), courseHandler.GetCoursesHandler)
		courseRouter.GET("/:id", middleware.CheckLogin(true), courseHandler.GetCourseHandler)
		courseRouter.GET("/upload/url", middleware.CheckLogin(true), courseHandler.GetUploadUrlHandler)
		courseRouter.POST("/upload/log", middleware.CheckLogin(true), courseHandler.LogUploadHandler)
		courseRouter.GET("/schedule", courseHandler.CourseScheduleHandler)
		courseRouter.GET("/histories/:number", middleware.CheckLogin(true), courseHandler.GetCourseHistoryHandler)
	}
	// 变量模块
	variableSvc := do.MustInvoke[*service.VariableService](nil)
	variableHandler := handler.NewVariableHandler(variableSvc)
	variableRouter := router.Group("/variables")
	{
		variableRouter.GET("", variableHandler.GetHandler)
		variableRouter.POST("", middleware.CheckSuper(), variableHandler.SetHandler)
	}
	// 消息模块
	messageSvc := do.MustInvoke[*service.MessageService](nil)
	messageHandler := handler.NewMessageHandler(messageSvc)
	messageRouter := router.Group("/messages")
	{
		messageRouter.GET("", middleware.CheckLogin(true), messageHandler.GetListHandler)
		messageRouter.GET("/unread_num", middleware.CheckLogin(true), messageHandler.GetUnreadNumHandler)
		messageRouter.GET("/unread_nums", middleware.CheckLogin(true), messageHandler.GetUnreadNumsHandler)
		messageRouter.POST("/system", middleware.CheckAdmin(), messageHandler.SendSystemHandler)
		messageRouter.GET("/push", middleware.CheckLogin(true), messageHandler.WebpushRequestKeyHandler)
		messageRouter.POST("/push", middleware.CheckLogin(true), messageHandler.WebpushSubscribeHandler)
		messageRouter.DELETE("/push", middleware.CheckLogin(true), messageHandler.WebpushUnsubscribeHandler)
	}
	// 治理模块
	manageSvc := do.MustInvoke[*service.ManageService](nil)
	manageHandler := handler.NewManageHandler(manageSvc)
	manageRouter := router.Group("/manage")
	{
		manageRouter.GET("/report_types", manageHandler.GetReportTypesHandler)
		manageRouter.POST("/reports", middleware.CheckLogin(true), manageHandler.ReportHandler)
		manageRouter.GET("/reports", middleware.CheckAdmin(), manageHandler.GetReportsHandler)
		manageRouter.PUT("reports/:id", middleware.CheckAdmin(), manageHandler.UpdateReportStatusHandler)
		manageRouter.POST("/bans", middleware.CheckAdmin(), manageHandler.BanHandler)
		manageRouter.GET("/bans", middleware.CheckAdmin(), manageHandler.GetBansHandler)
	}
}
