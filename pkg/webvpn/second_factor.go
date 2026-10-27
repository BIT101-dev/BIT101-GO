package webvpn

import (
	"BIT101-GO/pkg/request"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const secondFactorPhoneURL = schoolWebVPNCASBase + "/cas/api/protected/sms/getPhoneNumberByUserId"
const secondFactorSendURL = schoolWebVPNCASBase + "/cas/api/protected/sms/publicNoToken/sendSmsCode"
const secondFactorCheckURL = schoolWebVPNCASBase + "/cas/api/protected/sms/checkToken"

const secondFactorPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAjVr1zKwohU3xA0afprWLSQvIymaSH/V27MedFc+CecXSnORIFMAp4uEIb4taDq/2X4eMeTI66Mu/rB5GKSFDbExF2Gu4NaO/CNDpf1gHMScUrIFCh4CDqzBnx17kclvezLkIK0T8FVa4cRsINvzjbnA6jUSMaf6Fm1n9wTAtW6QYBjssGOEtCj+c38PTBdFMmJbXp3brt1tEBesz6lb3Fjp76FGvDZ08xtYG8fxYPuiMwKU04eS+mcX/BunwgpU3zwekHYB+PWRIvq0lBry9Wms25sJE5T/RAv5fEuMLbBkfcZK3+7ivSZthTmPpr2Ap/ji70ZZ6u2jvR5VJq+LJHQIDAQAB
-----END PUBLIC KEY-----`

type SecondFactorChallenge struct {
	Username    string `json:"username"`
	Cookie      string `json:"cookie"`
	Fingerprint string `json:"fingerprint"`
	Execution   string `json:"execution"`
	FormAction  string `json:"form_action"`
	Phone       string `json:"phone"`
	MaskedPhone string `json:"masked_phone"`
	ExpiresAt   int64  `json:"expires_at"`
}

type SecondFactorError struct {
	Challenge SecondFactorChallenge
}

func (e *SecondFactorError) Error() string {
	return "webvpn SMS verification required"
}

func SealSecondFactorChallenge(challenge SecondFactorChallenge, secret string) (string, error) {
	encoded, err := json.Marshal(challenge)
	if err != nil {
		return "", err
	}
	return SealSessionCookie(string(encoded), secret)
}

func OpenSecondFactorChallenge(encoded string, secret string) (SecondFactorChallenge, error) {
	plaintext, err := OpenSessionCookie(encoded, secret)
	if err != nil {
		return SecondFactorChallenge{}, err
	}
	var challenge SecondFactorChallenge
	if json.Unmarshal([]byte(plaintext), &challenge) != nil {
		return SecondFactorChallenge{}, errors.New("invalid SMS challenge")
	}
	if challenge.ExpiresAt <= time.Now().Unix() {
		return SecondFactorChallenge{}, errors.New("webvpn SMS challenge expired")
	}
	return challenge, nil
}

func isSecondFactorPage(res request.Response) bool {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
	if err != nil {
		return false
	}
	return doc.Find("#login-page-flowkey").Text() != "" && doc.Find("#user-object-id").Text() != "" &&
		(strings.Contains(res.Text, "secondSmsLoginForm") || strings.Contains(res.Text, "second-auth-tip") || strings.Contains(res.Text, "cas-gateway"))
}

func (c *Client) prepareSecondFactor(username string, res request.Response) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
	if err != nil {
		return err
	}
	base, err := url.Parse(res.URL)
	if err != nil {
		return err
	}
	action := doc.Find("form").First().AttrOr("action", "login")
	reference, err := url.Parse(action)
	if err != nil {
		return errors.New("invalid second factor form action")
	}
	formAction := base.ResolveReference(reference)
	if formAction.Scheme != "https" || formAction.User != nil || !strings.HasSuffix(formAction.Hostname(), ".bit.edu.cn") {
		return errors.New("invalid second factor form action")
	}
	phone, maskedPhone, err := c.secondFactorPhone(doc.Find("#user-object-id").Text())
	if err != nil {
		return err
	}
	if phone == "" {
		phone = doc.Find("#phone-number").Text()
	}
	if phone == "" {
		return errors.New("second factor phone missing")
	}
	if maskedPhone == "" {
		maskedPhone = "绑定手机"
	}
	challenge := SecondFactorChallenge{
		Username:    username,
		Cookie:      c.Cookie(),
		Fingerprint: c.fingerprint,
		Execution:   doc.Find("#login-page-flowkey").Text(),
		FormAction:  formAction.String(),
		Phone:       phone,
		MaskedPhone: maskedPhone,
		ExpiresAt:   time.Now().Add(5 * time.Minute).Unix(),
	}
	if err := c.sendSecondFactorSMS(challenge.Cookie, challenge.Phone); err != nil {
		return err
	}
	challenge.Cookie = c.Cookie()
	return &SecondFactorError{Challenge: challenge}
}

var ErrSMSCodeRejected = errors.New("webvpn SMS code rejected")

func (c *Client) CompleteSecondFactor(challenge SecondFactorChallenge, code string) error {
	if challenge.ExpiresAt <= time.Now().Unix() {
		return errors.New("webvpn SMS challenge expired")
	}
	if len(code) < 4 || len(code) > 8 || strings.ContainsFunc(code, func(r rune) bool { return r < '0' || r > '9' }) {
		return errors.New("invalid SMS code")
	}
	if challenge.Username == "" || challenge.Cookie == "" || challenge.Execution == "" || challenge.FormAction == "" || challenge.Phone == "" {
		return errors.New("invalid SMS challenge")
	}
	payload, _ := json.Marshal(map[string]any{
		"phone":       challenge.Phone,
		"token":       code,
		"delete":      false,
		"trustDevice": false,
	})
	res, err := request.PostJSONWithClient(c, secondFactorCheckURL, string(payload), protectedSecondFactorHeaders(challenge.Cookie, schoolWebVPNOrigin+"/cas/"))
	if err != nil {
		return err
	}
	if res.Code != http.StatusOK {
		return errors.New("webvpn SMS check failed")
	}
	if status, _ := secondFactorResponseStatus(res); status != 200 {
		return ErrSMSCodeRejected
	}
	res, err = request.PostFormWithClient(noRedirectClient{client: c}, challenge.FormAction, map[string]string{
		"username":     challenge.Username,
		"password":     code,
		"type":         "smsLogin",
		"_eventId":     "submit",
		"geolocation":  "",
		"execution":    challenge.Execution,
		"captcha_code": "",
		"trustDevice":  "false",
	}, secondFactorHeaders(challenge.Cookie, schoolWebVPNOrigin+"/cas/"))
	if err != nil {
		return err
	}
	if res.Code < 300 || res.Code >= 400 {
		return errors.New("webvpn SMS login rejected")
	}
	return c.finishLogin(res)
}

func (c *Client) secondFactorPhone(userObjectID string) (string, string, error) {
	aesKey := make([]byte, 16)
	if _, err := rand.Read(aesKey); err != nil {
		return "", "", err
	}
	plaintext, _ := json.Marshal(map[string]string{"userId": userObjectID})
	encrypted, err := EncryptPassword(string(plaintext), base64.StdEncoding.EncodeToString(aesKey))
	if err != nil {
		return "", "", err
	}
	block, _ := pem.Decode([]byte(secondFactorPublicKey))
	if block == nil {
		return "", "", errors.New("invalid second factor public key")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", "", err
	}
	publicKey, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return "", "", errors.New("invalid second factor public key")
	}
	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(base64.StdEncoding.EncodeToString(aesKey)))
	if err != nil {
		return "", "", err
	}
	headers := protectedSecondFactorHeaders(c.Cookie(), casLoginURL)
	headers["hasCrypto"] = "true"
	headers["privateKey"] = base64.StdEncoding.EncodeToString(encryptedKey)
	res, err := request.PostJSONWithClient(c, secondFactorPhoneURL, encrypted, headers)
	if err != nil {
		return "", "", err
	}
	if res.Code != http.StatusOK {
		return "", "", errors.New("second factor phone lookup failed")
	}
	payload, err := decryptSecondFactorResponse(res.Text, aesKey)
	if err != nil {
		return "", "", err
	}
	data, _ := payload["data"].(map[string]any)
	phone, _ := data["tel"].(string)
	maskedPhone, _ := data["maskTel"].(string)
	return phone, maskedPhone, nil
}

func decryptSecondFactorResponse(text string, key []byte) (map[string]any, error) {
	current := text
	for round := 0; round < 4; round++ {
		var value any
		if json.Unmarshal([]byte(current), &value) == nil {
			switch typed := value.(type) {
			case map[string]any:
				return typed, nil
			case string:
				current = typed
				continue
			}
		}
		encrypted, err := base64.StdEncoding.DecodeString(current)
		if err != nil {
			break
		}
		plaintext, err := decryptSchoolPassword(encrypted, key)
		if err != nil {
			break
		}
		current = string(plaintext)
	}
	return nil, errors.New("invalid second factor phone response")
}

func (c *Client) sendSecondFactorSMS(cookie string, phone string) error {
	payload, _ := json.Marshal(map[string]string{"phone": phone, "businessNo": "0008"})
	res, err := request.PostJSONWithClient(c, secondFactorSendURL, string(payload), protectedSecondFactorHeaders(cookie, casLoginURL))
	if err != nil {
		return err
	}
	status, message := secondFactorResponseStatus(res)
	if status != 200 && !(strings.Contains(message, "验证码") && strings.Contains(message, "有效期内") && strings.Contains(message, "重复发送")) {
		return errors.New("webvpn SMS request rejected")
	}
	return nil
}

func secondFactorResponseStatus(res request.Response) (int, string) {
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Msg     string `json:"msg"`
	}
	if json.Unmarshal(res.Content, &result) != nil {
		return 0, ""
	}
	if result.Message != "" {
		return result.Code, result.Message
	}
	return result.Code, result.Msg
}

func secondFactorHeaders(cookie string, referer string) map[string]string {
	return map[string]string{
		"Cookie":  cookie,
		"Origin":  schoolWebVPNOrigin,
		"Referer": referer,
	}
}

func protectedSecondFactorHeaders(cookie string, referer string) map[string]string {
	headers := secondFactorHeaders(cookie, referer)
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for index := range key {
		key[index] = alphabet[int(key[index])%len(alphabet)]
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	midpoint := len(encoded) / 2
	mixed := encoded[:midpoint] + encoded + encoded[midpoint:]
	digest := md5.Sum([]byte(mixed))
	headers["Csrf-Key"] = string(key)
	headers["Csrf-Value"] = hex.EncodeToString(digest[:])
	headers["Sid-Language"] = "zh_CN"
	return headers
}
