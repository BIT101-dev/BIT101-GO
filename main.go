/*
 * @Author: flwfdd
 * @Date: 2023-03-13 10:20:13
 * @LastEditTime: 2024-02-21 00:39:46
 * @Description: _(:з」∠)_
 */
package main

import (
	"BIT101-GO/database"
	"BIT101-GO/router"
	"BIT101-GO/util/config"
	"BIT101-GO/util/gorse"
	"BIT101-GO/util/other"
	"BIT101-GO/util/search"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var VERSION = "v1.0.2"

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

/*
日志记录改进
问题: 当前的日志直接使用 fmt.Println，难以追踪和管理。
改进: 使用标准的日志包zap支持日志级别和输出格式化。
*/
var logger, _ = zap.NewProduction()

// 服务，启动！
func runServer() {
	logger.Info("Initializing server...")
	config.Init()
	database.Init()
	search.Init()
	gorse.Init()
	go sync()

	if config.Config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	app := gin.Default()
	app.Use(limits.RequestSizeLimiter(config.Config.Saver.MaxSize << 20))
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Content-Type", "fake-cookie", "webvpn-cookie"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		// ExposeHeaders:    []string{"Content-Length"},
		// AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// 	return true
		// },
		MaxAge: 12 * time.Hour,
	}))
	router.SetRouter(app)
	logger.Info("Server started", zap.String("port", config.Config.Port))
	app.Run(":" + config.Config.Port)
}

/*
问题: 当前的 sync 使用了无限循环和 time.Sleep，可能导致 Goroutine 堆积或服务资源浪费。
改进建议: 使用 time.Ticker 替代 time.Sleep，并确保任务完成后再开始下一轮。
*/
func sync() {
	ticker := time.NewTicker(time.Duration(config.Config.SyncInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		timeAfter := time.Now()
		fmt.Println("Syncing... ", timeAfter.Format("2006-01-02 15:04:05"))
		go gorse.Sync(timeAfter)
		go search.Sync(timeAfter)
	}
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
