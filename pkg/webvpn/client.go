package webvpn

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	httpClient  *http.Client
	fingerprint string
	riskMode    string
	restoreOnce sync.Once
}

var defaultClient = &Client{fingerprint: newFingerprint()}

func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
			Jar:     jar,
		},
		fingerprint: newFingerprint(),
	}
}

func (c *Client) SetFingerprint(fingerprint string) {
	if fingerprint != "" {
		c.fingerprint = fingerprint
	}
}

func (c *Client) http() *http.Client {
	if c.httpClient != nil {
		return c.httpClient
	}
	return http.DefaultClient
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.do(req, true)
}

type noRedirectClient struct {
	client *Client
}

func (c noRedirectClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.do(req, false)
}

func (c *Client) do(req *http.Request, followRedirects bool) (*http.Response, error) {
	req = req.Clone(req.Context())
	for name, value := range browserHeaders {
		if req.Header.Get(name) == "" {
			req.Header.Set(name, value)
		}
	}
	if c.http().Jar != nil && req.Header.Get("Cookie") != "" {
		c.restoreOnce.Do(func() {
			cookies := req.Cookies()
			for _, cookie := range cookies {
				cookie.Path = "/"
			}
			c.http().Jar.SetCookies(req.URL, cookies)
		})
		req.Header.Del("Cookie")
	}
	client := *c.http()
	if !followRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	fingerprintHandled := false
	for {
		original := req.Clone(req.Context())
		response, err := client.Do(req)
		if err != nil {
			return nil, requestError(err)
		}
		if !isFingerprintURL(response.Request.URL) {
			return response, nil
		}
		response.Body.Close()
		if fingerprintHandled {
			return nil, errors.New("webvpn fingerprint was not accepted")
		}
		fingerprintHandled = true
		fingerprintResponse, err := client.Do(fingerprintRequest(req, response.Request.URL, c.fingerprint))
		if err != nil {
			return nil, requestError(err)
		}
		fingerprintResponse.Body.Close()
		if fingerprintResponse.StatusCode >= 400 {
			return nil, errors.New("webvpn fingerprint setup failed")
		}
		req, err = replayRequest(original)
		if err != nil {
			return nil, err
		}
	}
}

func replayRequest(req *http.Request) (*http.Request, error) {
	next := req.Clone(req.Context())
	if req.Body != nil {
		if req.GetBody == nil {
			return nil, errors.New("webvpn request body cannot be replayed")
		}
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		next.Body = body
	}
	return next, nil
}

func isFingerprintURL(url *url.URL) bool {
	return url != nil && url.Scheme == "https" && url.Hostname() == "webvpn.bit.edu.cn" && url.Path == "/fingerprint"
}

func fingerprintRequest(req *http.Request, current *url.URL, fingerprint string) *http.Request {
	next := req.Clone(req.Context())
	next.Method = http.MethodGet
	next.URL = new(url.URL)
	*next.URL = *current
	next.URL.Path = "/set-fingerprint"
	next.URL.RawPath = ""
	query := next.URL.Query()
	query.Set("fingerprint", fingerprint)
	next.URL.RawQuery = query.Encode()
	next.Body = nil
	next.GetBody = nil
	next.ContentLength = 0
	next.Header = req.Header.Clone()
	next.Header.Del("Content-Length")
	next.Header.Del("Content-Type")
	return next
}

func requestError(err error) error {
	var urlError *url.Error
	if errors.As(err, &urlError) {
		return requestError(urlError.Err)
	}
	return err
}

func newFingerprint() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func (c *Client) establishWebVPN() error {
	request, err := http.NewRequest(http.MethodGet, "https://webvpn.bit.edu.cn/set-fingerprint?fingerprint="+c.fingerprint, nil)
	if err != nil {
		return err
	}
	response, err := c.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return errors.New("webvpn fingerprint setup failed")
	}
	if c.http().Jar != nil && !c.hasWebVPNTicket() {
		return errors.New("webvpn ticket missing after fingerprint setup")
	}
	return nil
}

func (c *Client) hasWebVPNTicket() bool {
	if c.http().Jar == nil {
		return false
	}
	parsed, _ := url.Parse("https://webvpn.bit.edu.cn")
	for _, cookie := range c.http().Jar.Cookies(parsed) {
		if cookie.Value == "" {
			continue
		}
		if cookie.Name == "wengine_vpn_ticket" || cookie.Name == "wengine_vpn_ticketwebvpn_bit_edu_cn" {
			return true
		}
	}
	return false
}

func (c *Client) Cookie() string {
	if c.http().Jar == nil {
		return ""
	}
	cookies := make(map[string]string)
	for _, rawURL := range []string{"https://sso.bit.edu.cn", "https://webvpn.bit.edu.cn", casLoginURL, webvpnLoginURL} {
		parsed, _ := url.Parse(rawURL)
		for _, cookie := range c.http().Jar.Cookies(parsed) {
			if cookie.Value != "" {
				cookies[cookie.Name] = cookie.Value
			}
		}
	}
	names := make([]string, 0, len(cookies))
	for name := range cookies {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, name+"="+cookies[name])
	}
	return strings.Join(parts, "; ")
}
