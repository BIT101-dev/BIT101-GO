/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:11:38
 * @LastEditTime: 2023-03-21 20:43:23
 * @Description: 用户模块业务响应
 */
package controller

import (
	"BIT101-GO/controller/webvpn"
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/jwt"
	"BIT101-GO/util/mail"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 清除敏感信息
func CleanUser(user *database.User) {
	user.Password = ""
	user.Sid = ""
	// 转换头像链接
	user.Avatar = GetImageUrl(user.Avatar)
}

func GetUser(uid string) gin.H {
	if uid == "-1" {
		return gin.H{
			"id":          -1,
			"create_time": time.Now(),
			"nickname":    "匿名者",
			"avatar":      GetImageUrl(""),
			"motto":       "面对愚昧，匿名者自己也缄口不言。",
			"level":       1,
		}
	}
	var user database.User
	database.DB.Limit(1).Find(&user, "id = ?", uid)
	return gin.H{
		"id":          user.ID,
		"create_time": user.CreatedAt,
		"nickname":    user.Nickname,
		"avatar":      GetImageUrl(user.Avatar),
		"motto":       user.Motto,
		"level":       user.Level,
	}
}

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
	var user database.User
	database.DB.Limit(1).Find(&user, "sid = ?", query.Sid)
	if user.ID == 0 || user.Password != query.Password {
		c.JSON(500, gin.H{"msg": "登录失败Orz"})
		return
	}
	token := jwt.GetUserToken(fmt.Sprint(user.ID), config.Config.LoginExpire, config.Config.Key)
	c.JSON(200, gin.H{"msg": "登录成功OvO", "fake_cookie": token})
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
		captcha = "data:image/png;base64," + base64.StdEncoding.EncodeToString(img)
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
		c.JSON(500, gin.H{"msg": "统一身份认证失败Orz"})
		return
	}
	//生成验证码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	token := jwt.GetUserToken(query.Sid, config.Config.VerifyCodeExpire, config.Config.Key+code)
	c.JSON(200, gin.H{"msg": "统一身份认证成功OvO", "token": token, "code": code})
}

// 发送邮件验证码请求结构
type UserMailVerifyQuery struct {
	Sid string `form:"sid" binding:"required"` // 学号
}

// 发送邮件验证码
func UserMailVerify(c *gin.Context) {
	var query UserMailVerifyQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	//生成验证码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	token := jwt.GetUserToken(query.Sid, config.Config.VerifyCodeExpire, config.Config.Key+code)
	//发送邮件
	err := mail.Send(query.Sid+"@bit.edu.cn", "[BIT101]验证码", fmt.Sprintf("【%v】 是你的验证码ヾ(^▽^*)))", code))
	if err != nil {
		c.JSON(500, gin.H{"msg": "发送邮件失败Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "发送成功OvO", "token": token})
}

// 注册请求结构
type UserRegisterQuery struct {
	Password string `form:"password" binding:"required"` // MD5密码
	Token    string `form:"token" binding:"required"`    // token
	Code     string `form:"code" binding:"required"`     // 验证码
}

// 注册
func UserRegister(c *gin.Context) {
	var query UserRegisterQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 验证token
	sid, ok := jwt.VeirifyUserToken(query.Token, config.Config.Key+query.Code)
	if !ok {
		c.JSON(500, gin.H{"msg": "验证码无效Orz"})
		return
	}

	// 查询用户是否已经注册过
	var user database.User
	database.DB.Limit(1).Find(&user, "sid = ?", sid)
	if user.ID == 0 {
		// 未注册过
		user.Sid = sid
		user.Password = query.Password
		user.Motto = "I offer you the BITterness of a man who has looked long and long at the lonely moon." // 默认格言

		// 生成昵称
		user_ := database.User{}
		for {
			nickname := "BIT101-" + uuid.New().String()[:8]
			database.DB.Limit(1).Find(&user_, "nickname = ?", nickname)
			if user_.ID == 0 {
				user.Nickname = nickname
				break
			}
		}

		database.DB.Create(&user)
	} else {
		// 已经注册过 修改密码
		database.DB.Model(&user).Update("password", query.Password)
	}
	token := jwt.GetUserToken(fmt.Sprint(user.ID), config.Config.LoginExpire, config.Config.Key)
	c.JSON(200, gin.H{"msg": "注册成功OvO", "fake_cookie": token})
}

// 获取用户信息请求结构
type UserGetInfoQuery struct {
	Id string `form:"id"` // uid
}

// 获取用户信息
func UserGetInfo(c *gin.Context) {
	var query UserGetInfoQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 匿名用户
	if query.Id == "-1" {
		c.JSON(200, GetUser("-1"))
		return
	}

	var uid string
	if query.Id == "" || query.Id == "0" {
		// 获取自己的信息
		uid = c.GetString("uid")
		if uid == "" {
			c.JSON(401, gin.H{"msg": "请先登录awa"})
			return
		}
	} else {
		uid = query.Id
	}

	var user database.User
	database.DB.Limit(1).Find(&user, uid)
	if user.ID == 0 {
		c.JSON(500, gin.H{"msg": "用户不存在Orz"})
		return
	}

	CleanUser(&user)

	c.JSON(200, user)
}

// 修改用户信息请求结构
type UserSetInfoQuery struct {
	Nickname string `form:"nickname"` // 昵称
	Avatar   string `form:"avatar"`   // 头像
	Motto    string `form:"motto"`    // 格言 简介
}

// 修改用户信息
func UserSetInfo(c *gin.Context) {
	var query UserSetInfoQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	uid := c.GetString("uid")
	var user database.User
	database.DB.Limit(1).Find(&user, uid)
	if user.ID == 0 {
		c.JSON(500, gin.H{"msg": "用户不存在Orz"})
		return
	}

	if query.Nickname != "" {
		user_ := database.User{}
		database.DB.Limit(1).Find(&user_, "nickname = ?", query.Nickname)
		if user_.ID != 0 && user_.ID != user.ID {
			c.JSON(500, gin.H{"msg": "昵称冲突Orz"})
			return
		}
		user.Nickname = query.Nickname
	}
	if query.Avatar != "" {
		// 验证图片是否存在
		avatar := database.Image{}
		database.DB.Limit(1).Find(&avatar, "mid = ?", query.Avatar)
		if avatar.ID == 0 {
			c.JSON(500, gin.H{"msg": "头像图片无效Orz"})
			return
		}
		user.Avatar = query.Avatar
	}
	if query.Motto != "" {
		user.Motto = query.Motto
	}
	database.DB.Save(&user)
	c.JSON(200, gin.H{"msg": "修改成功OvO"})
}
