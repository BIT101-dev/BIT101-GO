/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:11:38
 * @LastEditTime: 2023-03-15 18:14:18
 * @Description: 用户模块业务响应
 */
package controller

import (
	"BIT101-GO/util/cache"
	"BIT101-GO/util/webvpn"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

// 登录请求结构
type UserLoginQuery struct {
	Sid      string `form:"sid" binding:"required"`      // 学号
	Password string `form:"password" binding:"required"` // MD5密码
}

// 登录
func UserLogin(c *gin.Context) {
	var query UserLoginQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	c.JSON(200, gin.H{"fake_cookie": "2333"})
}

// webvpn验证初始化请求结构
type UserWebvpnVerifyInitQuery struct {
	Sid string `form:"sid"` // 学号
}

// webvpn验证初始化
func UserWebvpnVerifyInit(c *gin.Context) {
	var query UserWebvpnVerifyInitQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	data, err := webvpn.InitLogin()
	if err != nil {
		c.JSON(500, gin.H{"msg": "初始化登陆失败Orz"})
		return
	}
	needCaptcha, err := webvpn.NeedCaptcha(query.Sid)
	if err != nil {
		c.JSON(500, gin.H{"msg": "检查是否需要验证失败Orz"})
		return
	}
	captcha := ""
	if needCaptcha {
		img, err := webvpn.CaptchaImage(data.Cookie)
		if err != nil {
			c.JSON(500, gin.H{"msg": "获取验证码图片失败Orz"})
			return
		}
		captcha = base64.StdEncoding.EncodeToString(img)
	}
	c.JSON(200, gin.H{
		"salt":      data.Salt,
		"execution": data.Execution,
		"cookie":    data.Cookie,
		"captcha":   captcha,
	})
}

// webvpn验证请求结构
type UserWebvpnVerifyQuery struct {
	Sid       string `form:"sid" binding:"required"`      // 学号
	Password  string `form:"password" binding:"required"` // EncryptedPassword.js加密后的密码
	Execution string `form:"execution" binding:"required"`
	Cookie    string `form:"cookie" binding:"required"`
	Captcha   string `form:"captcha"`
}

// webvpn验证
func UserWebvpnVerify(c *gin.Context) {
	var query UserWebvpnVerifyQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	err := webvpn.Login(query.Sid, query.Password, query.Execution, query.Cookie, query.Captcha)
	if err != nil {
		c.JSON(500, gin.H{"msg": "登陆失败Orz"})
		return
	}
	//生成验证码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	cache.Instance.Set("verify"+query.Sid, code, 60)
	c.JSON(200, gin.H{"verify_code": code})
}
