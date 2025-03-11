/*
 * @Author: flwfdd
 * @Date: 2023-03-16 09:10:10
 * @LastEditTime: 2025-03-11 10:56:12
 * @Description: 成绩模块业务响应
 */
package handler

import (
	"BIT101-GO/api/common"
	"BIT101-GO/pkg/webvpn"

	"github.com/gin-gonic/gin"
)

// ScoreHandler 成绩模块响应
type ScoreHandler struct{}

// NewScoreHandler 创建成绩模块响应
func NewScoreHandler() *ScoreHandler {
	return &ScoreHandler{}
}

// GetScoreHandler 成绩查询
func (h *ScoreHandler) GetScoreHandler(c *gin.Context) {
	type Request struct {
		Detail bool `form:"detail"` // 学号
	}
	type Response struct {
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	cookie := c.Request.Header.Get("webvpn-cookie")
	if cookie == "" {
		common.HandleErrorWithCode(c, webvpn.ErrCookieInvalid, 400)
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
	c.JSON(200, Response{"查询成功OvO", table})
}

// GetReportHandler 获取可信成绩单
func (h *ScoreHandler) GetReportHandler(c *gin.Context) {
	cookie := c.Request.Header.Get("webvpn-cookie")
	if cookie == "" {
		common.HandleErrorWithCode(c, webvpn.ErrCookieInvalid, 400)
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
