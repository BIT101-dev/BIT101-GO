/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:52:43
 * @LastEditTime: 2025-03-09 23:08:59
 * @Description: 用户模块中间件
 */
package middleware

import (
	"BIT101-GO/api/common"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserContext 用户上下文
type UserContext struct {
	UIDStr  string
	UIDUint uint
	UIDInt  int
	IsSuper bool
	IsAdmin bool
}

// UserContextKey 用户上下文键
const UserContextKey = "user_ctx"

// 验证用户是否登录
func CheckLogin(strict bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("fake-cookie")
		uid, ok, super, admin := common.VeirifyUserToken(token, config.Get().Key)
		if ok {
			uid_uint, err := strconv.ParseUint(uid, 10, 32)
			if err != nil {
				c.JSON(500, gin.H{"msg": "获取用户ID错误Orz"})
				c.Abort()
				return
			}
			if ban, ok := database.BanMap[uint(uid_uint)]; ok {
				if common.GetNowTime().Before(ban.Time) {
					c.JSON(403, gin.H{"msg": "您已被关小黑屋Orz,解封时间：" + ban.Time.Format("2006-01-02 15:04:05")})
					c.Abort()
					return
				}
			}

			// 写入上下文
			c.Set(UserContextKey, &UserContext{
				UIDStr:  uid,
				UIDUint: uint(uid_uint),
				UIDInt:  int(uid_uint),
				IsSuper: super,
				IsAdmin: admin || super,
			})
		} else if strict {
			c.JSON(401, gin.H{"msg": "请先登录awa"})
			c.Abort()
		}
	}
}

// GetUserContext 获取用户上下文
func GetUserContext(c *gin.Context) (UserContext, error) {
	user, exist := c.Get(UserContextKey)
	if !exist {
		return UserContext{}, errors.New("获取用户上下文错误Orz")
	}
	userCtx, ok := user.(*UserContext)
	if !ok {
		return UserContext{}, errors.New("获取用户上下文错误Orz")
	}
	return *userCtx, nil
}

// MustGetUserContext 获取用户上下文
func MustGetUserContext(c *gin.Context) UserContext {
	userCtx, err := GetUserContext(c)
	if err != nil {
		c.JSON(500, gin.H{"msg": "获取用户错误Orz"})
	}
	return userCtx
}

// CheckAdmin 验证用户是否为admin/super
func CheckAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		CheckLogin(true)(c)
		if !MustGetUserContext(c).IsAdmin {
			c.JSON(403, gin.H{"msg": "权限不足awa"})
			c.Abort()
		}
	}
}

// CheckSuper 验证用户是否为super
func CheckSuper() gin.HandlerFunc {
	return func(c *gin.Context) {
		CheckLogin(true)(c)
		if !MustGetUserContext(c).IsSuper {
			c.JSON(403, gin.H{"msg": "权限不足awa"})
			c.Abort()
		}
	}
}
