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
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 枚举用户类型(需要与数据库中定义一致)
const (
	Identity_Super = iota
	Identity_Normal
	Identity_Admin
	Identity_Organization
	Identity_Robot
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
	Avatar     ImageAPI  `json:"avatar"`
	Motto      string    `json:"motto"`
	Type       Type      `json:"type"`
}

type Type struct {
	Text  string `json:"text"`
	Color string `json:"color"`
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
				Avatar:     GetImageAPI(""),
				Motto:      "面对愚昧，匿名者自己也缄口不言。",
				Type:       Type{database.IdentityMap[Identity_Normal].Text, database.IdentityMap[Identity_Normal].Color},
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
			Avatar:     GetImageAPI(user.Avatar),
			Motto:      user.Motto,
			Type:       Type{database.IdentityMap[user.Identity].Text, database.IdentityMap[user.Identity].Color},
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
	token := jwt.GetUserToken(fmt.Sprint(user.ID), config.Config.LoginExpire, config.Config.Key, int(user.Identity))
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
	token := jwt.GetUserToken(query.Sid, config.Config.VerifyCodeExpire, config.Config.Key+code, -1)
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
	token := jwt.GetUserToken(query.Sid, config.Config.VerifyCodeExpire, config.Config.Key+code, -1)
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
	sid, ok, _, _ := jwt.VeirifyUserToken(query.Token, config.Config.Key+query.Code)
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
	token := jwt.GetUserToken(fmt.Sprint(user.ID), config.Config.LoginExpire, config.Config.Key, int(user.Identity))
	c.JSON(200, gin.H{"msg": "注册成功OvO", "fake_cookie": token})
}

// UserGetInfoResponse 获取用户信息返回结构
type UserGetInfoResponse struct {
	UserAPI      UserAPI `json:"user"`
	FollowingNum int64   `json:"following_num"`
	FollowerNum  int64   `json:"follower_num"`
	Following    bool    `json:"following"`
	Follower     bool    `json:"follower"`
	Own          bool    `json:"own"`
}

// UserGetInfo 获取用户信息
func UserGetInfo(c *gin.Context) {
	id_str := c.Param("id")
	if id_str == "-1" {
		c.JSON(200, UserGetInfoResponse{
			GetUserAPI(-1),
			0,
			0,
			false,
			false,
			false,
		})
		return
	}
	var uid uint
	if id_str == "" || id_str == "0" {
		// 获取自己的信息
		uid = c.GetUint("uid_uint")
		if uid == 0 {
			c.JSON(401, gin.H{"msg": "请先登录awa"})
			return
		}
	} else {
		uid_, err := strconv.ParseUint(id_str, 10, 32)
		if err != nil {
			c.JSON(400, gin.H{"msg": "参数错误awa"})
			return
		}
		var user database.User
		database.DB.Limit(1).Find(&user, "id = ?", uid_)
		if user.ID == 0 {
			c.JSON(404, gin.H{"msg": "用户不存在Orz"})
			return
		}
		uid = uint(uid_)
	}
	followPostResponse := GetFollowPostResponse(uid, c.GetUint("uid_uint"))
	c.JSON(200, UserGetInfoResponse{
		GetUserAPI(int(uid)),
		followPostResponse.FollowingNum,
		followPostResponse.FollowerNum,
		followPostResponse.Following,
		followPostResponse.Follower,
		uid == c.GetUint("uid_uint"),
	})
}

// OldUserGetInfoQuery 获取用户信息请求结构
type OldUserGetInfoQuery struct {
	Id string `form:"id"` // uid
}

// Deprecated: OldUserGetInfo 获取用户信息(旧)
func OldUserGetInfo(c *gin.Context) {
	var query OldUserGetInfoQuery
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
	Nickname  string `json:"nickname"`   // 昵称
	AvatarMid string `json:"avatar_mid"` // 头像
	Motto     string `json:"motto"`      // 格言 简介
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
	if query.AvatarMid != "" {
		// 验证图片是否存在
		avatar := database.Image{}
		if err := database.DB.Limit(1).Find(&avatar, "mid = ?", query.AvatarMid).Error; err != nil {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
			return
		}
		if avatar.ID == 0 {
			c.JSON(500, gin.H{"msg": "头像图片无效Orz"})
			return
		}
		user.Avatar = query.AvatarMid
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

// FollowPostResponse 关注请求返回结构
type FollowPostResponse struct {
	FollowingNum int64 `json:"following_num"` // 关注数
	FollowerNum  int64 `json:"follower_num"`  // 粉丝数
	Following    bool  `json:"following"`     // 是否被我关注
	Follower     bool  `json:"follower"`      // 是否关注我
}

func GetFollowPostResponse(targetUid uint, myUid uint) FollowPostResponse {
	var following_num int64
	var follower_num int64
	following := false
	follower := false
	database.DB.Model(&database.Follow{}).Where("uid = ?", targetUid).Count(&following_num)
	database.DB.Model(&database.Follow{}).Where("follow_uid = ?", targetUid).Count(&follower_num)
	var follow database.Follow
	var follow_ database.Follow
	database.DB.Limit(1).Find(&follow, "uid = ? AND follow_uid = ?", myUid, targetUid)
	if follow.ID != 0 {
		following = true
	}
	database.DB.Limit(1).Find(&follow_, "uid = ? AND follow_uid = ?", targetUid, myUid)
	if follow_.ID != 0 {
		follower = true
	}
	return FollowPostResponse{
		Follower:     follower,
		FollowerNum:  follower_num,
		Following:    following,
		FollowingNum: following_num,
	}
}

// FollowPost 关注
func FollowPost(c *gin.Context) {
	if c.Param("id") == "-1" {
		c.JSON(400, gin.H{"msg": "不能关注匿名者Orz"})
		return
	}
	if c.Param("id") == c.GetString("uid") || c.Param("id") == "0" {
		c.JSON(400, gin.H{"msg": "不能关注自己Orz"})
		return
	}
	// 将字符串转换为uint
	follow_uid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	var user database.User
	database.DB.Unscoped().Where("id = ?", follow_uid).Limit(1).Find(&user)
	if user.ID == 0 {
		c.JSON(404, gin.H{"msg": "不存在此对象Orz"})
		return
	}

	var follow database.Follow
	database.DB.Unscoped().Where("uid = ?", c.GetString("uid")).
		Where("follow_uid = ?", c.Param("id")).Limit(1).Find(&follow)
	if follow.ID == 0 { //新建
		follow = database.Follow{
			Uid:       c.GetUint("uid_uint"),
			FollowUid: uint(follow_uid),
		}
		database.DB.Create(&follow)
		MessageSend(int(follow.Uid), follow.FollowUid, "", "follow", "", "")
	} else if follow.DeletedAt.Valid { //取消删除
		follow.DeletedAt = gorm.DeletedAt{}
		database.DB.Unscoped().Save(&follow)
		MessageSend(int(follow.Uid), follow.FollowUid, "", "follow", "", "")
	} else { //删除
		database.DB.Delete(&follow)
	}
	c.JSON(200, GetFollowPostResponse(uint(follow_uid), c.GetUint("uid_uint")))
}

// FollowListQuery 获取关注列表请求参数
type FollowListQuery struct {
	Page uint `form:"page"` // 页数
}

// FollowListGet 获取关注列表
func FollowListGet(c *gin.Context) {
	var query FollowListQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	var follow_list []database.Follow
	database.DB.Where("uid = ?", c.GetString("uid")).Order("updated_at DESC").Offset(int(query.Page * config.Config.FollowPageSize)).Limit(int(config.Config.FollowPageSize)).Find(&follow_list)
	users := make([]UserAPI, 0, len(follow_list))
	for _, follow := range follow_list {
		users = append(users, GetUserAPI(int(follow.FollowUid)))
	}
	c.JSON(200, users)
}

// FansListGet 获取粉丝列表
func FansListGet(c *gin.Context) {
	var query FollowListQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	var follow_list []database.Follow
	database.DB.Where("follow_uid = ?", c.GetString("uid")).Order("updated_at DESC").Offset(int(query.Page * config.Config.FollowPageSize)).Limit(int(config.Config.FollowPageSize)).Find(&follow_list)
	users := make([]UserAPI, 0, len(follow_list))
	for _, follow := range follow_list {
		users = append(users, GetUserAPI(int(follow.Uid)))
	}
	c.JSON(200, users)
}
