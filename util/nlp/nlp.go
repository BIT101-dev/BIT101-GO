/*
 * @Author: flwfdd
 * @Date: 2023-03-22 23:35:32
 * @LastEditTime: 2023-03-23 00:38:05
 * @Description: _(:з」∠)_
 */
package nlp

import (
	"strings"

	"github.com/yanyiwu/gojieba"
)

var jieba = gojieba.NewJieba()

// 去除空白字符串
func RemoveBlank(s []string) []string {
	out := make([]string, 0)
	for _, word := range s {
		if strings.TrimSpace(word) != "" {
			out = append(out, word)
		}
	}
	return out
}

// 精确分词
func Cut(s string) []string {
	words := jieba.Cut(strings.ReplaceAll(s, "|", " "), true)
	return RemoveBlank(words)
}

// 搜索分词
func CutForSearch(s string) []string {
	words := jieba.CutForSearch(s, true)
	return RemoveBlank(words)
}
