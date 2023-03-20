/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:52:43
 * @LastEditTime: 2023-03-20 15:04:24
 * @Description: 用户模块中间件
 */
package middleware

import (
	"BIT101-GO/util/config"
	"BIT101-GO/util/jwt"

	"github.com/gin-gonic/gin"
)

// 验证用户是否登录
func CheckLogin(strict bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("fake-cookie")
		uid, ok := jwt.VeirifyUserToken(token, config.Config.Key)
		if ok {
			c.Set("uid", uid)
		} else if strict {
			c.JSON(401, gin.H{"msg": "请先登录awa"})
			c.Abort()
		}
	}
}
