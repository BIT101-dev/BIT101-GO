/*
 * @Author: flwfdd
 * @Date: 2023-03-13 10:20:13
 * @LastEditTime: 2023-09-23 23:36:27
 * @Description: _(:з」∠)_
 */
package main

import (
	"BIT101-GO/database"
	"BIT101-GO/router"
	"BIT101-GO/util/config"
	"BIT101-GO/util/other"
	"BIT101-GO/util/search"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
)

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
	config.Init()
	database.Init()

	//search
	search.Init()
	//search.Test()

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
	fmt.Println("BIT101-GO will run on port " + config.Config.Port)
	app.Run(":" + config.Config.Port)
}

func main() {
	flag.Usage = func() {
		fmt.Println(LOGO)
		fmt.Printf("Usage: %s [mode]\n", os.Args[0])
		fmt.Println("mode:")
		fmt.Println("\tserver\t\tRun server")
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
