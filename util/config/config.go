/*
 * @Author: flwfdd
 * @Date: 2023-03-18 09:43:50
 * @LastEditTime: 2023-09-21 16:46:08
 * @Description: _(:з」∠)_
 */
package config

import (
	"fmt"
	"os"

	"github.com/jinzhu/configor"
)

var Config = struct {
	Port  string
	Proxy struct {
		Enable bool
		Url    string
	}
	Key              string
	VerifyCodeExpire int64 `yaml:"verify_code_expire"`
	LoginExpire      int64 `yaml:"login_expire"`
	Mail             struct {
		Host     string
		User     string
		Password string
	}
	Dsn   string
	Saver struct {
		MaxSize int64 `yaml:"max_size"`
		Url     string
		Local   struct {
			Enable bool
			Path   string
		}
		Cos struct {
			Enable    bool
			SecretId  string `yaml:"secret_id"`
			SecretKey string `yaml:"secret_key"`
			Bucket    string
			Region    string
			Path      string
		}
		OneDrive struct {
			Enable       bool
			Api          string `yaml:"api"`
			AuthApi      string `yaml:"auth_api"`
			ClientId     string `yaml:"client_id"`
			ClientSecret string `yaml:"client_secret"`
			RefreshToken string `yaml:"refresh_token"`
		}
	}
	DefaultAvatar      string `yaml:"default_avatar"`
	PaperPageSize      uint   `yaml:"paper_page_size"`
	CommentPageSize    uint   `yaml:"comment_page_size"`
	CommentPreviewSize uint   `yaml:"comment_preview_size"`
	CoursePageSize     uint   `yaml:"course_page_size"`
	MessagePageSize    uint   `yaml:"message_page_size"`
	MainUrl            string `yaml:"main_url"`
	ReleaseMode        bool   `yaml:"release_mode"`
}{}

func Init() {
	path := "config.yml"
	_, err := os.Stat(path)
	if err != nil {
		fmt.Println("config.yml not found, please copy config_example.yml to config.yml and edit it")
		os.Exit(1)
	}
	configor.Load(&Config, path)
	// fmt.Printf("config: %#v", Config)
}
