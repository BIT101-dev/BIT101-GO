/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:11:38
 * @LastEditTime: 2025-05-12 20:59:35
 * @Description: 用户模块业务响应
 */
package handler

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/middleware"
	"BIT101-GO/api/service"
	"BIT101-GO/api/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户模块接口
type UserHandler struct {
	UserSvc *service.UserService
}

// NewUserHandler 创建用户模块接口
func NewUserHandler(s *service.UserService) UserHandler {
	return UserHandler{s}
}

// LoginHandler 登录接口
func (h *UserHandler) LoginHandler(c *gin.Context) {

	type Request struct {
		Sid      string `json:"sid" binding:"required"`      // 学号
		Password string `json:"password" binding:"required"` // MD5密码
	}

	type Response struct {
		FakeCookie string `json:"fake_cookie"`
		Msg        string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindJSON(&query), 400) {
		return
	}

	fakeCookie, err := h.UserSvc.Login(query.Sid, query.Password)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{
		FakeCookie: fakeCookie,
		Msg:        "登录成功OvO",
	})
}

// WebvpnVerifyInit webvpn验证初始化接口
func (h *UserHandler) WebvpnVerifyInitHandler(c *gin.Context) {
	type Request struct {
		Sid string `json:"sid"` // 学号
	}

	type Response struct {
		Salt      string `json:"salt"`      // 照搬webvpn的salt
		Execution string `json:"execution"` // 照搬webvpn的execution
		Cookie    string `json:"cookie"`    // 照搬webvpn的cookie
		Captcha   string `json:"captcha"`   // webpn验证码图片base64
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindJSON(&query), 400) {
		return
	}

	data, captcha, err := h.UserSvc.WebvpnVerifyInit(query.Sid)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{
		Salt:      data.Salt,
		Execution: data.Execution,
		Cookie:    data.Cookie,
		Captcha:   captcha,
	})
}

// WebvpnVerifyHandler webvpn验证接口
func (h *UserHandler) WebvpnVerifyHandler(c *gin.Context) {
	type Request struct {
		Sid       string `json:"sid" binding:"required"`       // 学号
		Salt      string `json:"salt" binding:"required"`      // 照搬webvpn的salt
		Password  string `json:"password" binding:"required"`  // EncryptedPassword.js加密后的密码
		Execution string `json:"execution" binding:"required"` // 照搬webvpn的execution
		Cookie    string `json:"cookie" binding:"required"`    // 照搬webvpn的cookie
		Captcha   string `json:"captcha"`                      // 验证码结果
	}

	type Response struct {
		Token string `json:"token"` // webvpn验证成功后的token
		Code  string `json:"code"`  // 用于注册的验证码
		Msg   string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindJSON(&query), 400) {
		return
	}

	token, code, err := h.UserSvc.WebvpnVerify(query.Sid, query.Salt, query.Password, query.Execution, query.Cookie, query.Captcha)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{
		Token: token,
		Code:  code,
		Msg:   "统一身份认证成功OvO",
	})
}

// MailVerifyHandler 发送邮件验证码接口
func (h *UserHandler) MailVerifyHandler(c *gin.Context) {
	type Request struct {
		Sid string `form:"sid" binding:"required"` // 学号
	}
	type Response struct {
		Token string `json:"token"`
		Msg   string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindJSON(&query), 400) {
		return
	}

	token, err := h.UserSvc.MailVerify(query.Sid)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{
		Token: token,
		Msg:   "发送成功OvO",
	})
}

// RegisterHandler 注册接口
func (h *UserHandler) RegisterHandler(c *gin.Context) {
	type Request struct {
		Password  string `json:"password" binding:"required"` // MD5密码
		Token     string `json:"token" binding:"required"`    // token
		Code      string `json:"code" binding:"required"`     // 验证码
		LoginMode bool   `json:"login_mode"`                  // 是否使用不强制修改密码的登录模式
	}

	type Response struct {
		FakeCookie string `json:"fake_cookie"`
		Msg        string `json:"msg"`
	}

	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindJSON(&query), 400) {
		return
	}

	fakeCookie, err := h.UserSvc.Register(query.Token, query.Code, query.Password, query.LoginMode)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{
		FakeCookie: fakeCookie,
		Msg:        "注册成功OvO",
	})
}

// GetInfo 获取用户信息接口
func (h *UserHandler) GetInfoHandler(c *gin.Context) {
	type Request struct {
		ID string `uri:"id"` // 用户ID
	}
	type Response types.UserInfo
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindUri(&query), 400) {
		return
	}
	var uid int
	// 获取当前用户
	myUserCtx, err := middleware.GetUserContext(c)
	if query.ID == "" || query.ID == "0" {
		if common.HandleErrorWithCode(c, err, 401) {
			return
		}
		uid = myUserCtx.UIDInt
	} else {
		uid, err = strconv.Atoi(query.ID)
		if common.HandleErrorWithCode(c, err, 400) {
			return
		}
	}

	user, err := h.UserSvc.GetInfo(uid, uint(myUserCtx.UIDInt))
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(user))
}

// Deprecated: OldUserGetInfo 获取用户信息(旧)
func (h *UserHandler) OldUserGetInfo(c *gin.Context) {
	type Request struct {
		ID string `form:"id"` // uid
	}
	type Response types.UserAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	var uid int
	// 获取当前用户
	myUserCtx, err := middleware.GetUserContext(c)
	if query.ID == "" || query.ID == "0" {
		if common.HandleErrorWithCode(c, err, 401) {
			return
		}
		uid = myUserCtx.UIDInt
	} else {
		uid, err = strconv.Atoi(query.ID)
		if common.HandleErrorWithCode(c, err, 400) {
			return
		}
	}

	user, err := h.UserSvc.OldGetInfo(uid)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(user))
}

// SetInfoHandler 修改用户信息接口
func (h *UserHandler) SetInfoHandler(c *gin.Context) {
	type Request struct {
		Nickname  string `json:"nickname"`   // 昵称
		AvatarMid string `json:"avatar_mid"` // 头像
		Motto     string `json:"motto"`      // 格言 简介
	}
	type Response struct {
		Msg string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}

	uid := middleware.MustGetUserContext(c).UIDInt
	if common.HandleError(c, h.UserSvc.SetInfo(uid, query.Nickname, query.AvatarMid, query.Motto)) {
		return
	}

	c.JSON(200, Response{Msg: "修改成功OvO"})
}

// FollowHandler 关注/取消关注接口
func (h *UserHandler) FollowHandler(c *gin.Context) {
	type Request struct {
		ID string `uri:"id"` // 用户ID
	}

	type Response types.FollowAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindUri(&query), 400) {
		return
	}
	followUid, err := strconv.Atoi(query.ID)
	if common.HandleErrorWithCode(c, err, 400) {
		return
	}
	myUid := middleware.MustGetUserContext(c).UIDUint

	if common.HandleError(c, h.UserSvc.Follow(followUid, myUid)) {
		return
	}
	followAPI, err := h.UserSvc.GetFollowAPI(uint(followUid), myUid)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(followAPI))
}

// GetFollowListHandler 获取关注列表接口
func (h *UserHandler) GetFollowListHandler(c *gin.Context) {
	type Request struct {
		Page uint `form:"page"` // 页数
	}

	type Response []types.UserAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}

	followList, err := h.UserSvc.GetFollowList(middleware.MustGetUserContext(c).UIDUint, query.Page)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(followList))
}

// GetFansListHandler 获取粉丝列表接口
func (h *UserHandler) GetFansListHandler(c *gin.Context) {
	type Request struct {
		Page uint `form:"page"` // 页数
	}

	type Response []types.UserAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}

	fansList, err := h.UserSvc.GetFansList(middleware.MustGetUserContext(c).UIDUint, query.Page)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(fansList))
}
