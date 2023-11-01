/**
* @author:YHCnb
* Package:
* @date:2023/10/23 19:38
* Description:
 */
package controller

import (
	"strings"
	"time"
)

// 用于分割字符串（处理空元素的情况）
func spilt(str string) []string {
	l := strings.Split(str, " ")
	out := make([]string, 0)
	for i := range l {
		if l[i] != "" {
			out = append(out, l[i])
		}
	}
	return out
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
