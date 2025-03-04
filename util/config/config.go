/*
 * @Author: flwfdd
 * @Date: 2023-03-18 09:43:50
 * @LastEditTime: 2025-02-08 15:42:02
 * @Description: _(:з」∠)_
 */
package config

import (
	"fmt"
	"os"

	"github.com/jinzhu/configor"
)

type Config struct {
	Port  string
	Proxy struct {
		Enable bool
		Url    string
	}
	Key              string
	VerifyCodeExpire int64 `yaml:"verify_code_expire"`
	LoginExpire      int64 `yaml:"login_expire"`
	SyncInterval     int64 `yaml:"sync_interval"`
	Mail             struct {
		Host     string
		User     string
		Password string
	}
	Dsn   string
	Redis struct {
		Addr string
	}
	Saver struct {
		MaxSize        int64 `yaml:"max_size"`
		Url            string
		ImageUrlSuffix string `yaml:"image_url_suffix"`
		Local          struct {
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
	PostPageSize       uint   `yaml:"post_page_size"`
	FollowPageSize     uint   `yaml:"follow_page_size"`
	ReportPageSize     uint   `yaml:"report_page_size"`
	BanPageSize        uint   `yaml:"ban_page_size"`
	RecommendPageSize  uint   `yaml:"recommend_page_size"`
	MainUrl            string `yaml:"main_url"`
	ReleaseMode        bool   `yaml:"release_mode"`
	Meilisearch        struct {
		Url       string
		MasterKey string `yaml:"master_key"`
	}
	WebPushKeys struct {
		Public  string `yaml:"vapid_public"`
		Private string `yaml:"vapid_private"`
	} `yaml:"web_push_keys"`
}

var config Config

func GetConfig() Config {
	return config
}

func init() {
	path := "config.yml"
	_, err := os.Stat(path)
	if err != nil {
		fmt.Println("config.yml not found, please copy config_example.yml to config.yml and edit it")
		os.Exit(1)
	}
	configor.Load(&config, path)
}
