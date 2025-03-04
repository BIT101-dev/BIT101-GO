/*
 * @Author: flwfdd
 * @Date: 2023-03-21 00:36:32
 * @LastEditTime: 2023-03-30 19:08:13
 * @Description: _(:з」∠)_
 */
package saver

import (
	"BIT101-GO/util/config"
	"errors"
	"path/filepath"
	"strings"
)

// 保存文件 返回url
func Save(path string, content []byte) (string, error) {
	println(path)
	err1 := SaveLocal(path, content)
	path = strings.ReplaceAll(path, "\\", "/")
	err2 := SaveCOS(path, content)
	if err1 != nil || err2 != nil {
		return "", errors.New("save failed")
	}
	return GetUrl(path), nil
}

// 通过文件路径获取url
func GetUrl(path string) string {
	return config.GetConfig().Saver.Url + filepath.Join("/", path)
}
