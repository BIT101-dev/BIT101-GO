/*
 * @Author: flwfdd
 * @Date: 2023-03-13 14:32:05
 * @LastEditTime: 2023-03-15 18:13:47
 * @Description: 封装的网络请求工具包
 */
package request

import (
	"bytes"
	"io"
	"net/http"
	net_url "net/url"
)

type Response struct {
	Code    int
	Text    string
	Content []byte
	Header  http.Header
}

func request(request_type string, url string, headers map[string]string, request_body io.Reader) (Response, error) {
	req, _ := http.NewRequest(request_type, url, request_body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, err
	}
	return Response{res.StatusCode, string(body), body, res.Header}, nil
}

func Get(url string, headers map[string]string) (Response, error) {
	return request("GET", url, headers, nil)
}

func Post(url string, headers map[string]string) (Response, error) {
	return request("POST", url, headers, nil)
}

func PostForm(url string, form map[string]string, headers map[string]string) (Response, error) {
	formValues := net_url.Values{}
	for k, v := range form {
		formValues.Set(k, v)
	}
	formDataStr := formValues.Encode()
	formDataBytes := []byte(formDataStr)
	formBytesReader := bytes.NewReader(formDataBytes)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	return request("POST", url, headers, formBytesReader)
}
