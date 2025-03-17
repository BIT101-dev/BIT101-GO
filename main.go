/*
 * @Author: flwfdd
 * @Date: 2023-03-13 10:20:13
 * @LastEditTime: 2025-03-17 22:29:22
 * @Description: _(:з」∠)_
 */
package main

import (
	"BIT101-GO/api"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"BIT101-GO/pkg/cache"
	"BIT101-GO/pkg/other"
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

var VERSION = "v1.1.0"

var LOGO = `
                                                              
 ________   ___   _________    _____   ________     _____     
|\   __  \ |\  \ |\___   ___\ / __  \ |\   __  \   / __  \    
\ \  \|\ /_\ \  \\|___ \  \_||\/_|\  \\ \  \|\  \ |\/_|\  \   
 \ \   __  \\ \  \    \ \  \ \|/ \ \  \\ \  \\\  \\|/ \ \  \  
  \ \  \|\  \\ \  \    \ \  \     \ \  \\ \  \\\  \    \ \  \ 
   \ \_______\\ \__\    \ \__\     \ \__\\ \_______\    \ \__\
    \|_______| \|__|     \|__|      \|__| \|_______|     \|__|
 ________   ________                                          
|\   ____\ |\   __  \                                         
\ \  \___| \ \  \|\  \                                        
 \ \  \  ___\ \  \\\  \                                       
  \ \  \|\  \\ \  \\\  \                                      
   \ \_______\\ \_______\                                     
    \|_______| \|_______|                                     
                                                              
`

// 服务，启动！
func runServer() {
	database.Init()
	cache.Init()

	// 获取配置
	cfg := config.Get()

	// 初始化Gin
	if cfg.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	app := gin.Default()

	// 进行依赖注入并注册路由
	api.SetupContainer()
	api.RegisterRouter(app)

	// 启动服务
	fmt.Println("BIT101-GO will run on port " + cfg.Port)
	app.Run(":" + cfg.Port)
}

func main() {
	flag.Usage = func() {
		fmt.Println(LOGO)
		fmt.Printf("Usage: %s [mode]\n", os.Args[0])
		fmt.Println("mode:")
		fmt.Println("\tserver\t\tRun server (default)")
		fmt.Println("\tversion\t\tShow version")
		fmt.Println("\tbackup\t\tBackup database")
		fmt.Println("\timport_course [path]\t\tImport course data from path/*.csv (default path: ./data/course/)")
		fmt.Println("\thistory_score [start_year] [end_year] [webvpn_cookie]\t\tGet history score from term start_year-start_year+1 to end_year-1-end_year")
	}

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = append(args, "server")
	}

	switch args[0] {
	case "server": // 启动服务
		runServer()
	case "version": // 显示版本
		fmt.Println("BIT101-GO " + VERSION)
	case "backup": // 备份数据库
		other.Backup()
	case "import_course": // 导入课程
		if len(args) <= 1 {
			args = append(args, "./data/course/")
		}
		other.ImportCourse(args[1])
	case "history_score": // 获取历史均分
		if len(args) <= 3 {
			flag.Usage()
			return
		}
		other.GetCourseHistory(args[1], args[2], args[3])
	default:
		flag.Usage()
	}
}
