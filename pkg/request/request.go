/*
 * @Author: flwfdd
 * @Date: 2023-03-13 14:32:05
 * @LastEditTime: 2023-03-23 19:16:41
 * @Description: 封装的网络请求工具包
 */
package request

import (
	"bytes"
	"context"
	"io"
	"net/http"
	net_url "net/url"
	"strings"
)

type Response struct {
	Code    int
	Text    string
	Content []byte
	Header  http.Header
	URL     string
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func request(client HTTPClient, request_type string, url string, headers map[string]string, request_body io.Reader) (Response, error) {
	return requestWithContext(context.Background(), client, request_type, url, headers, request_body)
}

func requestWithContext(ctx context.Context, client HTTPClient, request_type string, url string, headers map[string]string, request_body io.Reader) (Response, error) {
	req, err := http.NewRequestWithContext(ctx, request_type, url, request_body)
	if err != nil {
		return Response{}, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Response{}, err
	}
	finalURL := *res.Request.URL
	finalURL.RawQuery = ""
	finalURL.ForceQuery = false
	finalURL.Fragment = ""
	finalURL.User = nil
	return Response{res.StatusCode, string(body), body, res.Header, finalURL.String()}, nil
}

func Get(url string, headers map[string]string) (Response, error) {
	return GetWithClient(http.DefaultClient, url, headers)
}

func GetWithClient(client HTTPClient, url string, headers map[string]string) (Response, error) {
	return request(client, "GET", url, headers, nil)
}

func GetWithClientContext(ctx context.Context, client HTTPClient, url string, headers map[string]string) (Response, error) {
	return requestWithContext(ctx, client, "GET", url, headers, nil)
}

func Post(url string, headers map[string]string) (Response, error) {
	return PostWithClient(http.DefaultClient, url, headers)
}

func PostWithClient(client HTTPClient, url string, headers map[string]string) (Response, error) {
	return request(client, "POST", url, headers, nil)
}

func PostForm(url string, form map[string]string, headers map[string]string) (Response, error) {
	return PostFormWithClient(http.DefaultClient, url, form, headers)
}

func PostFormWithClient(client HTTPClient, url string, form map[string]string, headers map[string]string) (Response, error) {
	formValues := net_url.Values{}
	for k, v := range form {
		formValues.Set(k, v)
	}
	formDataStr := formValues.Encode()
	formDataBytes := []byte(formDataStr)
	formBytesReader := bytes.NewReader(formDataBytes)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	return request(client, "POST", url, headers, formBytesReader)
}

func PostJSON(url string, data string, headers map[string]string) (Response, error) {
	return PostJSONWithClient(http.DefaultClient, url, data, headers)
}

func PostJSONWithClient(client HTTPClient, url string, data string, headers map[string]string) (Response, error) {
	headers["Content-Type"] = "application/json; charset=UTF-8"
	return request(client, "POST", url, headers, strings.NewReader(data))
}

func Put(url string, data []byte, headers map[string]string) (Response, error) {
	return request(http.DefaultClient, "PUT", url, headers, bytes.NewBuffer(data))
}
