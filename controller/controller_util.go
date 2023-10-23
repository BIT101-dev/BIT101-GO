/**
* @author:YHCnb
* Package:
* @date:2023/10/23 19:38
* Description:
 */
package controller

import "strings"

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
