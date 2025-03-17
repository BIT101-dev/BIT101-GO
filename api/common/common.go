/*
 * @Author: flwfdd
 * @Date: 2025-03-06 12:00:39
 * @LastEditTime: 2025-03-12 10:12:04
 * @Description: _(:з」∠)_
 */
package common

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// HandleError 处理错误 提示信息为错误 返回是否有错误
func HandleError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	slog.Error("HandleError", "url", c.Request.URL.Path, "err", err.Error())
	c.JSON(500, gin.H{"msg": err.Error()})
	return true
}

// HandleErrorWithMessage 处理错误 自定义提示信息 返回是否有错误
func HandleErrorWithMessage(c *gin.Context, err error, message string) bool {
	if err == nil {
		return false
	}
	slog.Error("HandleErrorWithMessage", "url", c.Request.URL.Path, "err", err.Error())
	c.JSON(500, gin.H{"msg": message})
	return true
}

// HandleErrorWithCode 处理错误 使用返回码预设提示信息 返回是否有错误
func HandleErrorWithCode(c *gin.Context, err error, code int) bool {
	if err == nil {
		return false
	}
	switch code {
	case 400:
		c.JSON(code, gin.H{"msg": "参数错误awa"})
	case 401:
		c.JSON(code, gin.H{"msg": "请先登录awa"})
	default:
		c.JSON(code, gin.H{"msg": err.Error()})
	}
	slog.Error("HandleErrorWithCode", "code", code, "url", c.Request.URL.Path, "err", err.Error())
	return true
}

// HandleErrorWithCodeAndMessage 处理错误 自定义返回码和提示信息 返回是否有错误
func HandleErrorWithCodeAndMessage(c *gin.Context, err error, code int, message string) bool {
	if err == nil {
		return false
	}
	c.JSON(code, gin.H{"msg": message})
	return true
}

// GetNowTime 当前时间
func GetNowTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc)
}

// ParseTime 解析时间
func ParseTime(t string) (time.Time, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	_time, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return time.Time{}, err
	}
	return _time.In(loc), nil
}

// 用于分割字符串（处理空元素的情况）
func Spilt(str string) []string {
	l := strings.Split(str, " ")
	out := make([]string, 0)
	for i := range l {
		if l[i] != "" {
			out = append(out, l[i])
		}
	}
	return out
}
