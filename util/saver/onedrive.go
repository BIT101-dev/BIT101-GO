/*
 * @Author: flwfdd
 * @Date: 2023-03-23 17:28:04
 * @LastEditTime: 2023-06-20 16:21:05
 * @Description: _(:з」∠)_
 */
package saver

import (
	"BIT101-GO/util/config"
	"BIT101-GO/util/request"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var onedrive_refresh_time int64 = 0
var onedrive_token = ""
var mux sync.Mutex

// 获取OneDrive的token
func OneDriveGetHead() (map[string]string, error) {
	if time.Now().Unix() > onedrive_refresh_time {
		mux.Lock()
		defer mux.Unlock()
		if time.Now().Unix() > onedrive_refresh_time {
			data := map[string]string{
				"client_id":     config.Config.Saver.OneDrive.ClientId,
				"client_secret": config.Config.Saver.OneDrive.ClientSecret,
				"redirect_uri":  "http://localhost",
				"refresh_token": config.Config.Saver.OneDrive.RefreshToken,
				"grant_type":    "refresh_token",
			}
			res, err := request.PostForm(config.Config.Saver.OneDrive.AuthApi, data, map[string]string{})
			if err != nil {
				return nil, err
			}
			if res.Code != 200 {
				return nil, errors.New(res.Text)
			}
			var token struct {
				ExpiresIn   uint   `json:"expires_in"`
				AccessToken string `json:"access_token"`
			}
			if err := json.Unmarshal(res.Content, &token); err != nil {
				return nil, err
			}
			onedrive_refresh_time = time.Now().Unix() + int64(token.ExpiresIn)
			onedrive_token = "bearer " + token.AccessToken
		}
	}
	return map[string]string{
		"Authorization": onedrive_token,
	}, nil
}

// 获取OneDrive操作路径
func OneDriveGetPath(path string, op string) string {
	if len(path) == 0 || len(op) == 0 {
		return ""
	}
	if path[0] == '/' {
		path = path[1:]
	}
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	if op[0] == '/' {
		op = op[1:]
	}
	return fmt.Sprintf("%v/root:/BIT101/%v:/%v", config.Config.Saver.OneDrive.Api, path, op)
}

// 获取OneDrive上传链接
func OneDriveGetUploadUrl(path string) (string, error) {
	conflict := "fail"
	data := fmt.Sprintf(`{"item": {"@microsoft.graph.conflictBehavior": "%v"}}`, conflict)
	head, err := OneDriveGetHead()
	if err != nil {
		return "", err
	}
	res, err := request.PostJSON(OneDriveGetPath(path, "createUploadSession"), data, head)
	if err != nil {
		return "", err
	}
	if res.Code != 200 {
		return "", errors.New(res.Text)
	}
	var dic struct {
		UploadUrl string `json:"uploadUrl"`
	}
	if err := json.Unmarshal(res.Content, &dic); err != nil {
		return "", err
	}
	return dic.UploadUrl, nil
}

// 上传文件
func OneDriveUploadFile(path string, data []byte) error {
	var res request.Response
	size := len(data)
	if size > 4000000 {
		url, err := OneDriveGetUploadUrl(path)
		if err != nil {
			return err
		}
		chunk_size := 3276800
		for i := 0; i < size; i += chunk_size {
			chunk_data := data[i:min(i+chunk_size, size)]
			head, err := OneDriveGetHead()
			if err != nil {
				return err
			}
			head["Content-Length"] = fmt.Sprintf("%v", len(chunk_data))
			head["Content-Range"] = fmt.Sprintf("bytes %v-%v/%v", i, i+len(chunk_data)-1, size)
			res, err = request.Put(url, chunk_data, head)
			if err != nil {
				return err
			}
			if res.Code != 202 {
				break
			}
		}
	} else {
		head, err := OneDriveGetHead()
		if err != nil {
			return err
		}
		res, err = request.Put(OneDriveGetPath(path, "content"), data, head)
		if err != nil {
			return err
		}
	}
	if res.Code != 200 && res.Code != 201 {
		return errors.New(res.Text)
	}
	return nil
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
