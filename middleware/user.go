/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:52:43
 * @LastEditTime: 2023-03-23 22:24:06
 * @Description: 用户模块中间件
 */

/*
优化点说明
常量定义
把错误信息抽取为常量，避免硬编码，提高代码的可维护性。

抽取公共逻辑
提取了 parseUID 方法，用于统一处理 uid 的解析逻辑。
提取了 isUserBanned 方法，用于判断用户是否被禁用，减少重复代码。

简化条件判断
使用逻辑短路和更清晰的条件表达式来减少嵌套。
时间格式标准化

使用 time.RFC3339 代替手动格式化日期字符串，便于解析和对比。
*/
package middleware

import (
	"BIT101-GO/controller"
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/jwt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 常量定义
const (
	errInvalidToken     = "获取用户ID错误Orz"
	errNotLoggedIn      = "请先登录awa"
	errInsufficientPerm = "权限不足awa"
	errUserBanned       = "您已被关小黑屋Orz,解封时间："
)

// CheckLogin 验证用户是否登录
func CheckLogin(strict bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("fake-cookie")
		uid, isValid, isSuper, isAdmin := jwt.VeirifyUserToken(token, config.Config.Key)

		if !isValid {
			if strict {
				c.JSON(401, gin.H{"msg": errNotLoggedIn})
				c.Abort()
			}
			return
		}

		uidUint, err := parseUID(uid)
		if err != nil {
			c.JSON(500, gin.H{"msg": errInvalidToken})
			c.Abort()
			return
		}

		if isUserBanned(c, uidUint) {
			return
		}

		// 将用户信息存入上下文
		c.Set("uid", uid)
		c.Set("uid_uint", uidUint)
		c.Set("super", isSuper)
		c.Set("admin", isAdmin)
	}
}

// CheckAdmin 验证用户是否为管理员或超级管理员
func CheckAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("admin") && !c.GetBool("super") {
			c.JSON(403, gin.H{"msg": errInsufficientPerm})
			c.Abort()
		}
	}
}

// CheckSuper 验证用户是否为超级管理员
func CheckSuper() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("super") {
			c.JSON(403, gin.H{"msg": errInsufficientPerm})
			c.Abort()
		}
	}
}

// parseUID 将字符串 UID 转换为 uint 类型
func parseUID(uid string) (uint, error) {
	uidUint64, err := strconv.ParseUint(uid, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(uidUint64), nil
}

// isUserBanned 检查用户是否被禁用
func isUserBanned(c *gin.Context, uid uint) bool {
	if controller.CheckBan(uid) {
		banInfo := database.BanMap[uid]
		unbanTime, _ := controller.ParseTime(banInfo.Time)
		c.JSON(403, gin.H{"msg": errUserBanned + unbanTime.Format(time.RFC3339)})
		c.Abort()
		return true
	}
	return false
}
