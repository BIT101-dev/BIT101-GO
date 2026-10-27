package webvpn

import (
	"BIT101-GO/config"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const ProxyTokenHeader = "X-BIT101-Proxy-Token"

var proxyHTTPClient = &http.Client{
	Transport: http.DefaultTransport,
	Timeout:   45 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) == 0 || (req.URL.Scheme == via[0].URL.Scheme && req.URL.Host == via[0].URL.Host) {
			return nil
		}
		return http.ErrUseLastResponse
	},
}

type proxyVerifyResponse struct {
	Success     bool   `json:"success"`
	Status      string `json:"status"`
	Cookie      string `json:"webvpn_cookie"`
	Challenge   string `json:"challenge"`
	MaskedPhone string `json:"masked_phone"`
	Sid         string `json:"sid"`
}

type SMSRequiredError struct {
	Challenge   string
	MaskedPhone string
}

func (e *SMSRequiredError) Error() string {
	return "学校要求短信二次验证"
}

func proxyEndpoint(path string) (string, error) {
	base := strings.TrimSpace(config.Get().Proxy.Url)
	if base == "" {
		return "", errors.New("proxy url is empty")
	}
	return strings.TrimRight(base, "/") + path, nil
}

func proxyDo(req *http.Request) (*http.Response, error) {
	if config.Get().Proxy.Token == "" {
		return nil, errors.New("proxy token is empty")
	}
	req.Header.Set(ProxyTokenHeader, config.Get().Proxy.Token)
	return proxyHTTPClient.Do(req)
}

func ProxyVerifyInit(sid string) (InitLoginReturn, string, error) {
	if sid == "" {
		return InitLoginReturn{}, "", errors.New("sid is required")
	}
	endpoint, err := proxyEndpoint("/internal/webvpn/init")
	if err != nil {
		return InitLoginReturn{}, "", err
	}
	body, _ := json.Marshal(map[string]string{"sid": sid})
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return InitLoginReturn{}, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := proxyDo(req)
	if err != nil {
		return InitLoginReturn{}, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return InitLoginReturn{}, "", fmt.Errorf("proxy init status %d", resp.StatusCode)
	}
	var result struct {
		Salt        string `json:"salt"`
		Execution   string `json:"execution"`
		Cookie      string `json:"cookie"`
		Fingerprint string `json:"fingerprint"`
		Captcha     string `json:"captcha"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return InitLoginReturn{}, "", err
	}
	return InitLoginReturn{
		Salt:        result.Salt,
		Execution:   result.Execution,
		Cookie:      result.Cookie,
		Fingerprint: result.Fingerprint,
	}, result.Captcha, nil
}

func ProxyVerify(sid string, salt string, pwd string, execution string, cookie string, captcha string, fingerprint string) (string, error) {
	if sid == "" || salt == "" || pwd == "" || execution == "" || cookie == "" {
		return "", errors.New("missing required proxy verify field")
	}
	endpoint, err := proxyEndpoint("/internal/webvpn/verify")
	if err != nil {
		return "", err
	}
	payload := map[string]string{
		"sid":         sid,
		"salt":        salt,
		"password":    pwd,
		"execution":   execution,
		"cookie":      cookie,
		"captcha":     captcha,
		"fingerprint": fingerprint,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := proxyDo(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("proxy verify status %d", resp.StatusCode)
	}
	var result proxyVerifyResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return "", err
	}
	if !result.Success {
		if result.Status == "sms_required" && result.Challenge != "" {
			return "", &SMSRequiredError{Challenge: result.Challenge, MaskedPhone: result.MaskedPhone}
		}
		return "", fmt.Errorf("proxy verify failed: %s", result.Status)
	}
	return OpenSessionCookie(result.Cookie, config.Get().Proxy.Token)
}

func ProxyVerifySMS(challenge string, code string) (string, string, error) {
	if challenge == "" || code == "" {
		return "", "", errors.New("missing SMS challenge or code")
	}
	endpoint, err := proxyEndpoint("/internal/webvpn/verify_sms")
	if err != nil {
		return "", "", err
	}
	body, _ := json.Marshal(map[string]string{"challenge": challenge, "code": code})
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := proxyDo(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("proxy SMS verify status %d", resp.StatusCode)
	}
	var result proxyVerifyResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return "", "", err
	}
	if !result.Success {
		return "", "", errors.New(result.Status)
	}
	if result.Sid == "" {
		return "", "", errors.New("proxy SMS verify returned empty sid")
	}
	cookie, err := OpenSessionCookie(result.Cookie, config.Get().Proxy.Token)
	return result.Sid, cookie, err
}

func ProxySchedule(cookie string, term string) (GetScheduleResponse, error) {
	if cookie == "" {
		return GetScheduleResponse{}, errors.New("webvpn cookie is empty")
	}
	endpoint, err := proxyEndpoint("/internal/schedule")
	if err != nil {
		return GetScheduleResponse{}, err
	}
	body, _ := json.Marshal(map[string]string{"cookie": cookie, "term": term})
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return GetScheduleResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := proxyDo(req)
	if err != nil {
		return GetScheduleResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var result struct {
			Msg string `json:"msg"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result)
		if result.Msg != "" {
			return GetScheduleResponse{}, errors.New(result.Msg)
		}
		return GetScheduleResponse{}, fmt.Errorf("proxy schedule status %d", resp.StatusCode)
	}
	var result GetScheduleResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&result); err != nil {
		return GetScheduleResponse{}, err
	}
	return result, nil
}
