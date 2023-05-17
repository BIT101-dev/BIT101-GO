/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:11:38
 * @LastEditTime: 2023-05-17 16:50:12
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
func CleanUser(old_user database.User) database.User {
	user := old_user
	user.Password = ""
	user.Sid = ""
	// 转换头像链接
	user.Avatar = GetImageUrl(user.Avatar)
	return user
}

type UserAPI struct {
	ID         int       `json:"id"`
	CreateTime time.Time `json:"create_time"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	Motto      string    `json:"motto"`
	Level      int       `json:"level"`
}

// 获取用户信息
func GetUserAPI(uid int) UserAPI {
	return GetUserAPIMap(map[int]bool{uid: true})[uid]
}

// 批量获取用户信息
func GetUserAPIMap(uid_map map[int]bool) map[int]UserAPI {
	out := make(map[int]UserAPI)
	uid_list := make([]int, 0)
	for uid := range uid_map {
		if uid == -1 {
			out[-1] = UserAPI{
				ID:         -1,
				CreateTime: time.Now(),
				Nickname:   "匿名者",
				Avatar:     GetImageUrl(""),
				Motto:      "面对愚昧，匿名者自己也缄口不言。",
				Level:      1,
			}
		} else {
			uid_list = append(uid_list, uid)
		}
	}

	var users []database.User
	database.DB.Where("id IN ?", uid_list).Find(&users)
	for _, user := range users {
		out[int(user.ID)] = UserAPI{
			ID:         int(user.ID),
			CreateTime: user.CreatedAt,
			Nickname:   user.Nickname,
			Avatar:     GetImageUrl(user.Avatar),
			Motto:      user.Motto,
			Level:      user.Level,
		}
	}
	return out
}

// 登录请求结构
type UserLoginQuery struct {
	Sid      string `json:"sid" binding:"required"`      // 学号
	Password string `json:"password" binding:"required"` // MD5密码
}

// 登录
func UserLogin(c *gin.Context) {
	var query UserLoginQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	var user database.User
	if err := database.DB.Limit(1).Find(&user, "sid = ?", query.Sid).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if user.ID == 0 || user.Password != query.Password {
		c.JSON(500, gin.H{"msg": "登录失败Orz"})
		return
	}
	token := jwt.GetUserToken(fmt.Sprint(user.ID), config.Config.LoginExpire, config.Config.Key, user.Level == 0)
	c.JSON(200, gin.H{"msg": "登录成功OvO", "fake_cookie": token})
}

// webvpn验证初始化请求结构
type UserWebvpnVerifyInitQuery struct {
	Sid string `json:"sid"` // 学号
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
	Sid       string `json:"sid" binding:"required"`      // 学号
	Password  string `json:"password" binding:"required"` // EncryptedPassword.js加密后的密码
	Execution string `json:"execution" binding:"required"`
	Cookie    string `json:"cookie" binding:"required"`
	Captcha   string `json:"captcha"`
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
	token := jwt.GetUserToken(query.Sid, config.Config.VerifyCodeExpire, config.Config.Key+code, false)
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
	token := jwt.GetUserToken(query.Sid, config.Config.VerifyCodeExpire, config.Config.Key+code, false)
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
	Password  string `json:"password" binding:"required"` // MD5密码
	Token     string `json:"token" binding:"required"`    // token
	Code      string `json:"code" binding:"required"`     // 验证码
	LoginMode bool   `json:"login_mode"`                  // 是否使用不强制修改密码的登录模式
}

// 注册
func UserRegister(c *gin.Context) {
	var query UserRegisterQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 验证token
	sid, ok, _ := jwt.VeirifyUserToken(query.Token, config.Config.Key+query.Code)
	if !ok {
		c.JSON(500, gin.H{"msg": "验证码无效Orz"})
		return
	}

	// 查询用户是否已经注册过
	var user database.User
	if err := database.DB.Limit(1).Find(&user, "sid = ?", sid).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if user.ID == 0 {
		// 未注册过
		user.Sid = sid
		user.Password = query.Password
		user.Motto = "I offer you the BITterness of a man who has looked long and long at the lonely moon." // 默认格言

		// 生成昵称
		user_ := database.User{}
		for {
			nickname := "BIT101-" + uuid.New().String()[:8]
			if err := database.DB.Limit(1).Find(&user_, "nickname = ?", nickname).Error; err != nil {
				c.JSON(500, gin.H{"msg": "数据库错误Orz"})
				return
			}
			if user_.ID == 0 {
				user.Nickname = nickname
				break
			}
		}

		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
			return
		}
	} else {
		// 已经注册过且不处于登录模式则修改密码
		if !query.LoginMode {
			if err := database.DB.Model(&user).Update("password", query.Password).Error; err != nil {
				c.JSON(500, gin.H{"msg": "数据库错误Orz"})
				return
			}
		}
	}
	token := jwt.GetUserToken(fmt.Sprint(user.ID), config.Config.LoginExpire, config.Config.Key, user.Level == 0)
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
		c.JSON(200, GetUserAPI(-1))
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
	if err := database.DB.Limit(1).Find(&user, "id = ?", uid).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if user.ID == 0 {
		c.JSON(500, gin.H{"msg": "用户不存在Orz"})
		return
	}

	user = CleanUser(user)

	c.JSON(200, user)
}

// 修改用户信息请求结构
type UserSetInfoQuery struct {
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像
	Motto    string `json:"motto"`    // 格言 简介
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
	if err := database.DB.Limit(1).Find(&user, "id = ?", uid).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if user.ID == 0 {
		c.JSON(500, gin.H{"msg": "用户不存在Orz"})
		return
	}

	if query.Nickname != "" {
		user_ := database.User{}
		if err := database.DB.Limit(1).Find(&user_, "nickname = ?", query.Nickname).Error; err != nil {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
			return
		}
		if user_.ID != 0 && user_.ID != user.ID {
			c.JSON(500, gin.H{"msg": "昵称冲突Orz"})
			return
		}
		user.Nickname = query.Nickname
	}
	if query.Avatar != "" {
		// 验证图片是否存在
		avatar := database.Image{}
		if err := database.DB.Limit(1).Find(&avatar, "mid = ?", query.Avatar).Error; err != nil {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
			return
		}
		if avatar.ID == 0 {
			c.JSON(500, gin.H{"msg": "头像图片无效Orz"})
			return
		}
		user.Avatar = query.Avatar
	}
	if query.Motto != "" {
		user.Motto = query.Motto
	}
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "修改成功OvO"})
}
