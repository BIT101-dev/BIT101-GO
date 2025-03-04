/*
 * @Author: flwfdd
 * @Date: 2023-03-16 10:40:12
 * @LastEditTime: 2023-05-16 11:29:35
 * @Description: 代理中间件
 */
package middleware

import (
	"BIT101-GO/config"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// 使用代理请求
func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.GetConfig().Proxy.Enable {
			proxyUrl, _ := url.Parse(config.GetConfig().Proxy.Url)

			proxy := httputil.ReverseProxy{
				Director: func(req *http.Request) {
					req.Host = proxyUrl.Host
					req.URL.Scheme = proxyUrl.Scheme
					req.URL.Host = proxyUrl.Host
				},
				ModifyResponse: func(resp *http.Response) error {
					// 防止返回重复的CORS头
					resp.Header.Del("Access-Control-Allow-Origin")
					return nil
				},
			}

			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}

	}
}
