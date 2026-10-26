/*
 * @Author: flwfdd
 * @Date: 2023-03-16 10:40:12
 * @LastEditTime: 2023-05-16 11:29:35
 * @Description: 代理中间件
 */
package middleware

import (
	"BIT101-GO/config"
	"BIT101-GO/pkg/webvpn"
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 使用代理请求
func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Get().Proxy.Enable {
			return
		}
		proxyURL, err := url.Parse(config.Get().Proxy.Url)
		if err != nil || proxyURL.Scheme == "" || proxyURL.Host == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "代理配置错误"})
			return
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.ResponseHeaderTimeout = 45 * time.Second
		proxy := httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.Host = proxyURL.Host
				req.URL.Scheme = proxyURL.Scheme
				req.URL.Host = proxyURL.Host
				req.URL.Path = joinProxyPath(proxyURL.Path, req.URL.Path)
				req.URL.RawPath = ""
				req.Header.Del("fake-cookie")
				req.Header.Del("Authorization")
				req.Header.Del("Cookie")
				req.Header.Set(webvpn.ProxyTokenHeader, config.Get().Proxy.Token)
			},
			ModifyResponse: func(resp *http.Response) error {
				if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
					body := []byte(`{"msg":"学校代理鉴权失败，请检查代理配置"}`)
					_ = resp.Body.Close()
					resp.Body = io.NopCloser(bytes.NewReader(body))
					resp.StatusCode = http.StatusBadGateway
					resp.Status = "502 Bad Gateway"
					resp.ContentLength = int64(len(body))
					resp.Header.Set("Content-Length", strconv.Itoa(len(body)))
					resp.Header.Set("Content-Type", "application/json; charset=utf-8")
					resp.TransferEncoding = nil
				}
				resp.Header.Del("Access-Control-Allow-Origin")
				return nil
			},
			Transport: transport,
		}

		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

func joinProxyPath(base string, requestPath string) string {
	if base == "" || base == "/" {
		return requestPath
	}
	return strings.TrimRight(base, "/") + requestPath
}
