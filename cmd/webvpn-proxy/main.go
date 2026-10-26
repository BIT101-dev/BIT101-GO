package main

import (
	"BIT101-GO/pkg/webvpn"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	token := strings.TrimSpace(os.Getenv("BIT101_PROXY_TOKEN"))
	if token == "" {
		log.Fatal("BIT101_PROXY_TOKEN is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(recoverRequests, logRequests)
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	authorized := router.Group("/", proxyAuth(token), schoolConcurrencyLimit(4))
	authorized.POST("/internal/webvpn/init", handleInit)
	authorized.POST("/internal/webvpn/verify", func(c *gin.Context) {
		handleVerify(c, token)
	})
	authorized.POST("/internal/schedule", handleSchedule)
	authorized.GET("/score", handleScore)
	authorized.GET("/score/report", handleReport)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      75 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("webvpn proxy listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func recoverRequests(c *gin.Context) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("request panic recovered: %T", recovered)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "代理处理出错"})
		}
	}()
	c.Next()
}

func logRequests(c *gin.Context) {
	started := time.Now()
	c.Next()
	log.Printf("%s %s %d %dms", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(started).Milliseconds())
}

func proxyAuth(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if subtle.ConstantTimeCompare([]byte(c.GetHeader(webvpn.ProxyTokenHeader)), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
			return
		}
		c.Next()
	}
}

func schoolConcurrencyLimit(max int) gin.HandlerFunc {
	slots := make(chan struct{}, max)
	return func(c *gin.Context) {
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
			c.Next()
		default:
			c.Header("Retry-After", "5")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "学校请求并发已满，请稍后重试"})
		}
	}
}

func handleInit(c *gin.Context) {
	var input struct {
		Sid string `json:"sid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "status": "invalid_request"})
		return
	}
	started := time.Now()
	result, err := webvpn.NewClient().InitLogin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "status": "init_failed", "duration_ms": time.Since(started).Milliseconds()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"salt":        result.Salt,
		"execution":   result.Execution,
		"cookie":      result.Cookie,
		"fingerprint": result.Fingerprint,
		"captcha":     "",
		"duration_ms": time.Since(started).Milliseconds(),
	})
}

func handleVerify(c *gin.Context, proxyToken string) {
	var input struct {
		Sid         string `json:"sid" binding:"required"`
		Salt        string `json:"salt" binding:"required"`
		Password    string `json:"password" binding:"required"`
		Execution   string `json:"execution" binding:"required"`
		Cookie      string `json:"cookie" binding:"required"`
		Captcha     string `json:"captcha"`
		Fingerprint string `json:"fingerprint"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "status": "invalid_request"})
		return
	}
	started := time.Now()
	client := webvpn.NewClient()
	client.SetFingerprint(input.Fingerprint)
	if err := client.Login(input.Sid, input.Salt, input.Password, input.Execution, input.Cookie, input.Captcha); err != nil {
		log.Printf("verify request failed: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"success":     false,
			"status":      "login_rejected",
			"duration_ms": time.Since(started).Milliseconds(),
		})
		return
	}
	cookie, err := webvpn.SealSessionCookie(client.Cookie(), proxyToken)
	if err != nil {
		log.Printf("verify session encryption failed")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "status": "session_encoding_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"status":        "success",
		"webvpn_cookie": cookie,
		"duration_ms":   time.Since(started).Milliseconds(),
	})
}

func handleSchedule(c *gin.Context) {
	var input struct {
		Cookie string `json:"cookie" binding:"required"`
		Term   string `json:"term"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}
	schedule, err := webvpn.NewClient().GetSchedule(input.Cookie, input.Term)
	if err != nil {
		log.Printf("schedule request failed: %v", err)
		if err == webvpn.ErrCookieInvalid {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "未通过身份认证Orz"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "获取课表出错Orz"})
		}
		return
	}
	c.JSON(http.StatusOK, schedule)
}

func handleScore(c *gin.Context) {
	cookie := c.GetHeader("webvpn-cookie")
	if cookie == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "未通过身份认证Orz"})
		return
	}
	detail := c.DefaultQuery("detail", "false") == "true"
	table, err := webvpn.NewClient().GetScoreContext(c.Request.Context(), cookie, detail)
	if err != nil {
		switch err {
		case webvpn.ErrCookieInvalid:
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "未通过身份认证Orz"})
		case webvpn.ErrNeedReview:
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "请先完成评教Orz"})
		default:
			log.Printf("score request failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "查询成绩出错Orz"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "查询成功OvO", "data": table})
}

func handleReport(c *gin.Context) {
	cookie := c.GetHeader("webvpn-cookie")
	if cookie == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "未通过身份认证Orz"})
		return
	}
	images, err := webvpn.NewClient().GetReportContext(c.Request.Context(), cookie)
	if err != nil {
		if err == webvpn.ErrCookieInvalid {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "未通过身份认证Orz"})
		} else {
			log.Printf("report request failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "获取可信成绩单出错Orz"})
		}
		return
	}
	payload := gin.H{"msg": "获取成功OvO", "data": images}
	encoded, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "编码可信成绩单出错Orz"})
		return
	}
	if len(encoded) > 5*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"msg": "可信成绩单响应过大"})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", encoded)
}
