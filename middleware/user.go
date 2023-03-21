/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:52:43
 * @LastEditTime: 2023-03-21 17:50:35
 * @Description: 用户模块中间件
 */
package middleware

import (
	"BIT101-GO/util/config"
	"BIT101-GO/util/jwt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 验证用户是否登录
func CheckLogin(strict bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("fake-cookie")
		uid, ok := jwt.VeirifyUserToken(token, config.Config.Key)
		if ok {
			c.Set("uid", uid)
			uid_uint, err := strconv.ParseUint(uid, 10, 32)
			if err != nil {
				c.JSON(500, gin.H{"msg": "获取用户ID错误Orz"})
				c.Abort()
				return
			}
			c.Set("uid_uint", uint(uid_uint))
		} else if strict {
			c.JSON(401, gin.H{"msg": "请先登录awa"})
			c.Abort()
		}
	}
}
