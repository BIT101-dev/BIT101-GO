/*
 * @Author: flwfdd
 * @Date: 2023-03-13 14:38:28
 * @LastEditTime: 2025-05-12 21:38:29
 * @Description: webvpn学校统一身份验证
 */
package webvpn

import (
	"BIT101-GO/pkg/request"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var ErrCookieInvalid = errors.New("未通过身份认证Orz")

// 登录初始化返回结构
type InitLoginReturn struct {
	Salt        string
	Execution   string
	Cookie      string
	Fingerprint string
}

const casLoginURL = "https://sso.bit.edu.cn/cas/login?service=https%3A%2F%2Fwebvpn.bit.edu.cn%2Flogin%3Fcas_login%3Dtrue"
const webvpnLoginURL = "https://webvpn.bit.edu.cn/https/77726476706e69737468656265737421e3e44ed225397c1e7b0c9ce29b5b/cas/login?service=https%3A%2F%2Fwebvpn.bit.edu.cn%2Flogin%3Fcas_login%3Dtrue"

// 登录初始化
func InitLogin() (InitLoginReturn, error) {
	return defaultClient.InitLogin()
}

func (c *Client) InitLogin() (InitLoginReturn, error) {
	return c.initLogin(casLoginURL)
}

func (c *Client) initLogin(loginURL string) (InitLoginReturn, error) {
	res, err := request.GetWithClient(c, loginURL, nil)
	if err != nil || res.Code != 200 {
		return InitLoginReturn{}, errors.New("webvpn init login error")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
	if err != nil {
		return InitLoginReturn{}, err
	}
	cookie := res.Header.Get("Set-Cookie")
	if sessionCookie := c.Cookie(); sessionCookie != "" {
		cookie = sessionCookie
	}
	result := InitLoginReturn{
		Salt:        doc.Find("#login-croypto").Text(),
		Execution:   doc.Find("#login-page-flowkey").Text(),
		Cookie:      cookie,
		Fingerprint: c.fingerprint,
	}
	if result.Salt == "" || result.Execution == "" || result.Cookie == "" {
		return InitLoginReturn{}, errors.New("webvpn init page invalid")
	}
	return result, nil
}

// 登录
func Login(username string, salt string, password string, execution string, cookie string, captcha string) error {
	return defaultClient.Login(username, salt, password, execution, cookie, captcha)
}

func (c *Client) Login(username string, salt string, password string, execution string, cookie string, captcha string) error {
	form := map[string]string{
		"username":        username,
		"password":        password,
		"execution":       execution,
		"captcha_payload": captcha,
		"croypto":         salt,
		"captcha_code":    "",
		"type":            "UsernamePassword",
		"_eventId":        "submit",
		"geolocation":     "",
	}
	risk, err := c.riskFields(salt, cookie)
	if err != nil {
		return err
	}
	for name, value := range risk {
		form[name] = value
	}
	if !c.hasWebVPNTicket() {
		if err := c.establishWebVPN(); err != nil {
			return err
		}
	}
	res, err := request.PostFormWithClient(noRedirectClient{client: c}, casLoginURL, form, map[string]string{
		"Cookie":  cookie,
		"Origin":  "https://sso.bit.edu.cn",
		"Referer": casLoginURL,
	})
	if err != nil {
		return fmt.Errorf("webvpn login request: %w", err)
	}
	if isSecondFactorPage(res) {
		return c.prepareSecondFactor(username, res)
	}
	if res.Code < 300 || res.Code >= 400 {
		if isCASLoginPage(res) {
			return loginRejection(res, c.riskMode, execution)
		}
		return fmt.Errorf("webvpn login rejected: status=%d risk=%s", res.Code, c.riskMode)
	}
	return c.finishLogin(res)
}

func (c *Client) finishLogin(res request.Response) error {
	callback, err := loginCallback(res)
	if err != nil {
		return err
	}
	if !c.hasWebVPNTicket() {
		if err := c.establishWebVPN(); err != nil {
			return err
		}
	}
	res, err = request.GetWithClient(c, callback, nil)
	if err != nil {
		return fmt.Errorf("webvpn callback request: %w", err)
	}
	if res.Code != 200 || isCASLoginPage(res) || !c.hasWebVPNTicket() {
		return fmt.Errorf("webvpn callback rejected: status=%d risk=%s", res.Code, c.riskMode)
	}
	log.Printf("webvpn login callback succeeded: upstream=%s", res.URL)
	return nil
}

func loginCallback(res request.Response) (string, error) {
	location := res.Header.Get("Location")
	if location == "" {
		return "", errors.New("webvpn login callback missing")
	}
	base, _ := url.Parse(casLoginURL)
	callback, err := url.Parse(location)
	if err != nil {
		return "", errors.New("webvpn login callback invalid")
	}
	callback = base.ResolveReference(callback)
	if callback.Scheme != "https" || callback.User != nil ||
		!strings.HasSuffix(callback.Hostname(), ".bit.edu.cn") ||
		(callback.Port() != "" && callback.Port() != "443") || callback.Query().Get("ticket") == "" {
		return "", errors.New("webvpn login callback invalid")
	}
	return callback.String(), nil
}

func loginRejection(res request.Response, riskMode string, execution string) error {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
	code := strings.TrimSpace(doc.Find("#login-error-code").Text())
	messagePresent := strings.TrimSpace(doc.Find("#login-error-msg").Text()) != ""
	flowReplaced := doc.Find("#login-page-flowkey").Text() != execution
	return fmt.Errorf("webvpn login rejected: status=%d code=%s risk=%s message=%t flow_replaced=%t", res.Code, code, riskMode, messagePresent, flowReplaced)
}

func isCASLoginPage(res request.Response) bool {
	return strings.Contains(res.Text, "帐号登录或动态码登录") ||
		strings.Contains(res.Text, "login-page-flowkey") ||
		strings.HasSuffix(strings.TrimSuffix(res.URL, "/"), "/cas/login")
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
func (c *Client) preCheck(url string, cookie string) (request.Response, error) {
	res, err := request.GetWithClient(c, url, map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return res, errors.New("webvpn precheck error")
	}
	if isCASLoginPage(res) {
		log.Printf("webvpn precheck rejected: upstream=%s", res.URL)
		return res, ErrCookieInvalid
	}
	return res, nil
}
