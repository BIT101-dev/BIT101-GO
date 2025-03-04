/*
 * @Author: flwfdd
 * @Date: 2023-03-13 14:38:28
 * @LastEditTime: 2023-03-18 10:47:35
 * @Description: webvpn学校统一身份验证
 */
package webvpn

import (
	"BIT101-GO/pkg/request"
	"encoding/json"
	"errors"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var ErrCookieInvalid = errors.New("webvpn cookie invalid")

// 登录初始化返回结构
type InitLoginReturn struct {
	Salt      string
	Execution string
	Cookie    string
}

// 登录初始化
func InitLogin() (InitLoginReturn, error) {
	res, err := request.Post("https://webvpn.bit.edu.cn/http/77726476706e69737468656265737421fcf84695297e6a596a468ca88d1b203b/authserver/login?service=https%3A%2F%2Fwebvpn.bit.edu.cn%2Flogin%3Fcas_login%3Dtrue", nil)
	if err != nil || res.Code != 200 {
		return InitLoginReturn{}, errors.New("webvpn init login error")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
	if err != nil {
		return InitLoginReturn{}, err
	}
	form := goquery.NewDocumentFromNode(doc.Find("#pwdFromId").Nodes[0])
	return InitLoginReturn{
		Salt:      form.Find("#pwdEncryptSalt").AttrOr("value", ""),
		Execution: form.Find("#execution").AttrOr("value", ""),
		Cookie:    res.Header.Get("Set-Cookie"),
	}, nil
}

// 登录
func Login(username string, password string, execution string, cookie string, captcha string) error {
	res, err := request.PostForm("https://webvpn.bit.edu.cn/https/77726476706e69737468656265737421fcf84695297e6a596a468ca88d1b203b/authserver/login?service=https%3A%2F%2Fwebvpn.bit.edu.cn%2Flogin%3Fcas_login%3Dtrue", map[string]string{
		"username":   username,
		"password":   password,
		"execution":  execution,
		"captcha":    captcha,
		"_eventId":   "submit",
		"rememberMe": "true",
		"cllt":       "userNameLogin",
		"dllt":       "generalLogin",
		"lt":         "",
	}, map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 || strings.Contains(res.Text, "帐号登录或动态码登录") {
		return errors.New("webvpn login error")
	}
	return nil
}

// 是否需要验证码
func NeedCaptcha(username string) (bool, error) {
	res, err := request.Get("https://webvpn.bit.edu.cn/https/77726476706e69737468656265737421fcf84695297e6a596a468ca88d1b203b/authserver/checkNeedCaptcha.htl?username="+username, nil)
	if err != nil || res.Code != 200 {
		return false, errors.New("webvpn need captcha error")
	}
	var data struct {
		IsNeed bool `json:"isNeed"`
	}
	err = json.Unmarshal([]byte(res.Text), &data)
	if err != nil {
		return false, err
	}
	return data.IsNeed, nil
}

// 获取验证码图片
func CaptchaImage(cookie string) ([]byte, error) {
	res, err := request.Get("https://webvpn.bit.edu.cn/https/77726476706e69737468656265737421fcf84695297e6a596a468ca88d1b203b/authserver/getCaptcha.htl", map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return nil, errors.New("webvpn captcha image error")
	}
	return res.Content, nil
}

// 前序验证并检查cookie是否有效
func preCheck(url string, cookie string) (request.Response, error) {
	res, err := request.Get(url, map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return res, errors.New("webvpn precheck error")
	}
	if strings.Contains(res.Text, "帐号登录或动态码登录") {
		return res, ErrCookieInvalid
	}
	return res, nil
}
