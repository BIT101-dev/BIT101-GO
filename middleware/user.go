/*
 * @Author: flwfdd
 * @Date: 2023-03-13 11:52:43
 * @LastEditTime: 2023-03-15 18:13:59
 * @Description: 用户模块中间件
 */
package middleware

import (
	"github.com/gin-gonic/gin"
)

// 验证用户是否登录
func CheckLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		fake_cookie := c.GetHeader("fake_cookie")
		if fake_cookie != "2333" {
			c.JSON(401, gin.H{"msg": "未登录"})
			c.Abort()
		}
	}
}
