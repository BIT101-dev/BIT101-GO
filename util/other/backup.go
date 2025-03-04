/*
 * @Author: flwfdd
 * @Date: 2023-11-19 12:14:30
 * @LastEditTime: 2023-11-19 14:08:50
 * @Description: _(:з」∠)_
 */
package other

import (
	"BIT101-GO/util/config"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func BackupDatabase() error {
	// 解析DSN
	// 使用空格分割字符串
	pairs := strings.Fields(config.GetConfig().Dsn)

	// 创建一个映射用于存储键值对
	params := make(map[string]string)

	// 遍历键值对并解析
	for _, pair := range pairs {
		// 使用等号分割键值对
		splits := strings.Split(pair, "=")
		if len(splits) == 2 {
			key := splits[0]
			value := splits[1]
			params[key] = value
		}
	}

	// 提取连接参数
	user := params["user"]
	password := params["password"]
	host := params["host"]
	port := params["port"]
	dbName := params["dbname"]

	// 定义备份文件的绝对路径
	backupPath := "./data/backup/" + time.Now().Format("20060102_150405") + ".sql"

	// 构建pg_dump命令
	cmdArg := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", host, port, user, password, dbName)
	// 执行pg_dump命令
	cmd := exec.Command("pg_dump", "-f", backupPath, cmdArg)
	err := cmd.Run()
	if err != nil {
		println(err)
		return err
	}

	println("数据库备份至" + backupPath)
	return nil
}

func Backup() {
	// 初始化备份文件夹
	if err := os.MkdirAll("./data/backup/", 0750); err != nil {
		println(err)
		return
	}

	if BackupDatabase() != nil {
		println("备份失败Orz")
	}
}
