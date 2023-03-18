/*
 * @Author: flwfdd
 * @Date: 2023-03-16 10:40:12
 * @LastEditTime: 2023-03-18 09:56:21
 * @Description: 代理中间件
 */
package middleware

import (
	"BIT101-GO/util/config"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// 使用代理请求
func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.Config.Proxy.Enable {
			proxyUrl, _ := url.Parse(config.Config.Proxy.Url)

			proxy := httputil.ReverseProxy{Director: func(req *http.Request) {
				req.Host = proxyUrl.Host
				req.URL.Scheme = proxyUrl.Scheme
				req.URL.Host = proxyUrl.Host
			}}

			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}

	}
}
