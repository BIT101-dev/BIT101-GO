package webvpn

import (
	"BIT101-GO/pkg/request"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

const schoolUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36"
const schoolRiskURL = schoolWebVPNCASBase + "/ustc-rba-front/fp"

var browserHeaders = map[string]string{
	"User-Agent":         schoolUserAgent,
	"Accept-Language":    "zh-CN,zh;q=0.9,en;q=0.7",
	"sec-ch-ua":          `"Not;A=Brand";v="8", "Chromium";v="150", "Google Chrome";v="150"`,
	"sec-ch-ua-mobile":   "?0",
	"sec-ch-ua-platform": `"macOS"`,
}

func (c *Client) riskFields(salt string, cookie string) (map[string]string, error) {
	groupID := schoolCookieValue(cookie, "riskSystemGroupId")
	device := schoolCookieValue(cookie, "device")
	if device == "" {
		device = schoolSHA256(strconv.FormatInt(time.Now().UnixMilli(), 10))
		cookie = cookie + "; device=" + device
	}
	payload := []byte(`{"error":true}`)
	fingerprint, _ := json.Marshal(schoolRiskFingerprint(device, groupID))
	res, err := request.PostJSONWithClient(c, schoolRiskURL, string(fingerprint), map[string]string{
		"Cookie":  cookie,
		"Origin":  schoolWebVPNOrigin,
		"Referer": casLoginURL,
	})
	if err == nil && res.Code == 200 {
		var result struct {
			Token string `json:"responsetoken"`
			Data  struct {
				Token string `json:"responsetoken"`
			} `json:"data"`
		}
		if json.Unmarshal(res.Content, &result) == nil {
			if result.Token == "" {
				result.Token = result.Data.Token
			}
			if result.Token != "" {
				payload, _ = json.Marshal(map[string]string{"token": result.Token, "groupId": groupID})
				c.riskMode = "ustc-token"
			}
		}
	}
	if c.riskMode != "ustc-token" {
		c.riskMode = "error-fallback"
	}
	encrypted, err := EncryptPassword(string(payload), salt)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"risk_payload": encrypted,
		"targetSystem": "sso",
		"siteId":       "sourceId",
		"riskEngine":   "true",
	}, nil
}

func schoolCookieValue(header string, name string) string {
	for _, part := range strings.Split(header, ";") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok && key == name {
			return value
		}
	}
	return ""
}

func schoolRiskFingerprint(device string, groupID string) map[string]string {
	values := []string{
		`["Arial","Helvetica Neue","PingFang SC","Times New Roman"]`,
		"16",
		"10",
		`"Asia/Shanghai"`,
		`"not available"`,
		`"MacIntel"`,
		`"zh-CN"`,
		"[956,1470]",
	}
	return map[string]string{
		"fonts":                 schoolSHA256(values[0]),
		"deviceMemory":          schoolSHA256(values[1]),
		"hardwareConcurrency":   schoolSHA256(values[2]),
		"localgroupId":          groupID,
		"timezone":              values[3],
		"cpuClass":              schoolSHA256(values[4]),
		"platform":              values[5],
		"language":              values[6],
		"screenResolution":      values[7],
		"fingerprint":           schoolSHA256(strings.Join(values, "")),
		"cookieValue":           device,
		"userAgent":             schoolUserAgent,
		"platformAuthenticator": "support",
	}
}

func schoolSHA256(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
