/*
 * @Author: flwfdd
 * @Date: 2023-03-18 09:43:50
 * @LastEditTime: 2023-03-20 14:27:08
 * @Description: _(:з」∠)_
 */
package config

import (
	"fmt"

	"github.com/jinzhu/configor"
)

var Config = struct {
	Proxy struct {
		Enable bool   `default:"false"`
		Url    string `default:""`
	}
	Key              string `default:"BIT101"`
	VerifyCodeExpire int64  `default:"300"`
	LoginExpire      int64  `default:"3600"`
	Mail             struct {
		Host     string
		User     string
		Password string
	}
}{}

func Init() {
	configor.Load(&Config, "config.yml")
	fmt.Printf("config: %#v", Config)
}
