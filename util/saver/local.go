/*
 * @Author: flwfdd
 * @Date: 2023-03-20 19:36:00
 * @LastEditTime: 2023-03-20 20:46:29
 * @Description: _(:з」∠)_
 */
package saver

import (
	"BIT101-GO/util/config"
	"os"
	"path/filepath"
)

// 保存文件到本地 path为子路径
func SaveLocal(path string, content []byte) error {
	// 检查配置开关
	if !config.GetConfig().Saver.Local.Enable {
		return nil
	}

	// 创建路径
	dst := filepath.Join(config.GetConfig().Saver.Local.Path, path)
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	// 写入文件
	os.WriteFile(dst, content, 0666)
	return nil
}
