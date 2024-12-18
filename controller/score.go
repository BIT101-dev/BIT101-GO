/*
 * @Author: flwfdd
 * @Date: 2023-03-16 09:10:10
 * @LastEditTime: 2023-03-29 15:41:51
 * @Description: 成绩模块业务响应
 */

/*
	优化

统一错误处理函数 (HandleError)：
简化 c.JSON 错误响应的逻辑，减少重复代码。
提高代码一致性。

独立错误分类处理函数 (handleWebvpnError)：
将 webvpn 的错误处理提取为独立函数，便于复用和后续扩展。
统一处理逻辑，减少重复代码。

模块化代码结构：
把独立的逻辑（如错误处理）封装成函数，让主逻辑更加清晰。
*/
package controller

import (
	"BIT101-GO/controller/webvpn"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 成绩查询请求结构
type ScoreQuery struct {
	Detail bool `form:"detail"` // 学号
}

// HandleError 统一错误处理函数
func HandleError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"msg": msg})
}

// 成绩查询
func Score(c *gin.Context) {
	var query ScoreQuery
	if err := c.ShouldBind(&query); err != nil {
		HandleError(c, http.StatusBadRequest, "参数错误awa")
		return
	}
	cookie := c.Request.Header.Get("webvpn-cookie")
	if cookie == "" {
		HandleError(c, http.StatusBadRequest, "参数错误awa")
		return
	}
	table, err := webvpn.GetScore(cookie, query.Detail)
	if err != nil {
		handleWebvpnError(c, err, "查询成绩出错Orz")
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "查询成功OvO", "data": table})
}

// 获取可信成绩单
func Report(c *gin.Context) {
	cookie := c.Request.Header.Get("webvpn-cookie")
	if cookie == "" {
		HandleError(c, http.StatusBadRequest, "参数错误awa")
		return
	}
	imgs, err := webvpn.GetReport(cookie)
	if err != nil {
		handleWebvpnError(c, err, "获取可信成绩单出错Orz")
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "获取成功OvO", "data": imgs})
}

// handleWebvpnError 处理 WebVPN 错误
func handleWebvpnError(c *gin.Context, err error, defaultMsg string) {
	if err == webvpn.ErrCookieInvalid {
		HandleError(c, http.StatusUnauthorized, "未通过身份认证Orz")
	} else {
		HandleError(c, http.StatusInternalServerError, defaultMsg)
	}
}
