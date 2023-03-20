/*
 * @Author: flwfdd
 * @Date: 2023-03-18 09:43:50
 * @LastEditTime: 2023-03-21 01:24:04
 * @Description: _(:з」∠)_
 */
package config

import (
	"fmt"

	"github.com/jinzhu/configor"
)

var Config = struct {
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
		Url   string
		Local struct {
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
	}
	DefaultAvatar string `yaml:"default_avatar"`
}{}

func Init() {
	configor.Load(&Config, "config.yml")
	fmt.Printf("config: %#v", Config)
}
