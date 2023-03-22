/*
 * @Author: flwfdd
 * @Date: 2023-03-16 09:10:10
 * @LastEditTime: 2023-03-22 00:48:38
 * @Description: 成绩模块业务响应
 */
package controller

import (
	"BIT101-GO/controller/webvpn"

	"github.com/gin-gonic/gin"
)

// 成绩查询请求结构
type ScoreQuery struct {
	Detail bool `form:"detail"` // 学号
}

// 成绩查询
func Score(c *gin.Context) {
	var query ScoreQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	cookie := c.Request.Header.Get("webvpn-cookie")
	if cookie == "" {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	table, err := webvpn.GetScore(cookie, query.Detail)
	if err != nil {
		if err == webvpn.ErrCookieInvalid {
			c.JSON(500, gin.H{"msg": "未通过身份认证Orz"})
		} else {
			c.JSON(500, gin.H{"msg": "查询成绩出错Orz"})
		}
		return
	}
	c.JSON(200, gin.H{"msg": "查询成功OvO", "data": table})
}

// 获取可信成绩单
func Report(c *gin.Context) {
	var query ScoreQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	cookie := c.Request.Header.Get("webvpn-cookie")
	if cookie == "" {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	imgs, err := webvpn.GetReport(cookie)
	if err != nil {
		if err == webvpn.ErrCookieInvalid {
			c.JSON(500, gin.H{"msg": "未通过身份认证Orz"})
		} else {
			c.JSON(500, gin.H{"msg": "获取可信成绩单出错Orz"})
		}
		return
	}
	c.JSON(200, gin.H{"msg": "获取成功OvO", "data": imgs})
}
