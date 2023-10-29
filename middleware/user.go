/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:52:43
 * @LastEditTime: 2023-03-23 22:24:06
 * @Description: 用户模块中间件
 */
package middleware

import (
	"BIT101-GO/controller"
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/jwt"
	"github.com/gin-gonic/gin"
	"strconv"
)

// 验证用户是否登录
func CheckLogin(strict bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("fake-cookie")
		uid, ok, super, admin := jwt.VeirifyUserToken(token, config.Config.Key)
		if ok {
			uid_uint, err := strconv.ParseUint(uid, 10, 32)
			if err != nil {
				c.JSON(500, gin.H{"msg": "获取用户ID错误Orz"})
				c.Abort()
				return
			}
			if controller.CheckBan(uint(uid_uint)) {
				t, _ := controller.ParseTime(database.BanMap[uint(uid_uint)].Time)
				c.JSON(401, gin.H{"msg": "您已被关小黑屋Orz,解封时间：" + t.Format("2006-01-02 15:04:05")})
				c.Abort()
				return
			}
			c.Set("uid", uid)
			c.Set("uid_uint", uint(uid_uint))
			c.Set("super", super)
			c.Set("admin", admin)
		} else if strict {
			c.JSON(401, gin.H{"msg": "请先登录awa"})
			c.Abort()
		}
	}
}
