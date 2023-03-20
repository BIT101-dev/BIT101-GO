/*
 * @Author: flwfdd
 * @Date: 2023-03-21 00:36:32
 * @LastEditTime: 2023-03-21 00:51:54
 * @Description: _(:з」∠)_
 */
package saver

import (
	"BIT101-GO/util/config"
	"errors"
	"path/filepath"
)

// 保存文件 返回url
func Save(path string, content []byte) error {
	err1 := SaveLocal(path, content)
	err2 := SaveCOS(path, content)
	if err1 != nil || err2 != nil {
		return errors.New("save failed")
	}
	return nil
}

// 通过文件路径获取url
func GetUrl(path string) string {
	return filepath.Join(config.Config.Saver.Url, path)
}
