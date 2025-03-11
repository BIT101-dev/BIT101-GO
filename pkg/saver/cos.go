/*
 * @Author: flwfdd
 * @Date: 2023-03-20 16:34:11
 * @LastEditTime: 2023-03-23 23:09:03
 * @Description: _(:з」∠)_
 */
package saver

import (
	"BIT101-GO/config"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/tencentyun/cos-go-sdk-v5"
)

var client *cos.Client
var once sync.Once

func InitCOS() {
	once.Do(func() {
		u, _ := url.Parse(fmt.Sprintf("https://%v.cos.%v.myqcloud.com", config.Get().Saver.Cos.Bucket, config.Get().Saver.Cos.Region))
		b := &cos.BaseURL{BucketURL: u}

		client = cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  config.Get().Saver.Cos.SecretId,
				SecretKey: config.Get().Saver.Cos.SecretKey,
			},
		})
	})
}

// 保存文件到腾讯云COS
func SaveCOS(path string, data []byte) error {
	if !config.Get().Saver.Cos.Enable {
		return nil
	}
	InitCOS()
	_, err := client.Object.Put(context.Background(), path, bytes.NewReader(data), nil)
	if err != nil {
		return err
	}
	return nil
}
