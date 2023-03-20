/*
 * @Author: flwfdd
 * @Date: 2023-03-18 09:43:50
 * @LastEditTime: 2023-03-20 02:15:47
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
	VerifyCodeExpire int64  `default:"600"`
}{}

func Init() {
	configor.Load(&Config, "config.yml")
	fmt.Printf("config: %#v", Config)
}
