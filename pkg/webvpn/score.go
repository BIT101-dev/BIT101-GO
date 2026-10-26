/*
 * @Author: flwfdd
 * @Date: 2023-03-15 18:24:54
 * @LastEditTime: 2025-05-12 21:37:25
 * @Description: webvpn成绩模块
 */
package webvpn

import (
	"BIT101-GO/pkg/request"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	ErrNeedReview = errors.New("need review") // 需要评教
)

func GetScore(cookie string, detail bool) ([][]string, error) {
	return GetScoreContext(context.Background(), cookie, detail)
}

func GetScoreContext(ctx context.Context, cookie string, detail bool) ([][]string, error) {
	return NewClient().GetScoreContext(ctx, cookie, detail)
}

func (c *Client) GetScore(cookie string, detail bool) ([][]string, error) {
	return c.GetScoreContext(context.Background(), cookie, detail)
}

func (c *Client) GetScoreContext(ctx context.Context, cookie string, detail bool) ([][]string, error) {
	// 预登录
	pre_url := "https://webvpn.bit.edu.cn/https/77726476706e69737468656265737421e3e44ed225397c1e7b0c9ce29b5b/cas/login?service=http%3A%2F%2Fjwms.bit.edu.cn%2F"
	_, err := c.preCheck(pre_url, cookie)
	if err != nil {
		return nil, err
	}
	// 获取成绩列表
	score_url := "https://webvpn.bit.edu.cn/http/77726476706e69737468656265737421fae04c8f69326144300d8db9d6562d/jsxsd/kscj/cjcx_list"
	res, err := request.GetWithClient(c, score_url, map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return nil, errors.New("webvpn get_score error")
	}
	if strings.Contains(res.Text, "评教") {
		return nil, ErrNeedReview
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
	if err != nil {
		return nil, err
	}
	rows := doc.Find("#dataList tr").Nodes
	if len(rows) == 0 {
		if strings.Contains(res.Text, "帐号登录或动态码登录") {
			return nil, ErrCookieInvalid
		}
		return nil, fmt.Errorf("webvpn score table not found: upstream=%s", res.URL)
	}
	head := make([]string, 0)
	head_map := make(map[string]int)
	data := make([]map[string]string, 0)
	type detailJob struct {
		index int
		url   string
	}
	jobs := make([]detailJob, 0)
	// 获取头
	goquery.NewDocumentFromNode(rows[0]).Children().Each(func(i int, s *goquery.Selection) {
		text := s.Get(0).FirstChild.Data
		head = append(head, text)
		head_map[text] = i
	})
	// 获取数据
	for row_index, tr := range rows[1:] {
		data = append(data, make(map[string]string))
		goquery.NewDocumentFromNode(tr).Children().Each(func(i int, s *goquery.Selection) {
			a := s.Find("a")
			if len(a.Nodes) == 1 { //包含<a>标签
				re := regexp.MustCompile("/jsxsd/kscj/cjfx.+cjfs")
				detail_url := re.FindString(a.AttrOr("onclick", ""))
				if detail_url != "" && detail { //查询详细信息
					detail_url = "https://webvpn.bit.edu.cn/http/77726476706e69737468656265737421fae04c8f69326144300d8db9d6562d" + strings.ReplaceAll(detail_url, "&amp;", "&")
					jobs = append(jobs, detailJob{row_index, detail_url})
				}
				data[row_index][head[i]] = strings.TrimSpace(a.Text())
			} else {
				x := s.Get(0).FirstChild
				if x != nil {
					data[row_index][head[i]] = strings.TrimSpace(x.Data)
				}
			}
		})
	}
	details, err := parallelMap(ctx, jobs, func(ctx context.Context, job detailJob) (map[string]string, error) {
		res, err := request.GetWithClientContext(ctx, c, job.url, map[string]string{"Cookie": cookie})
		if err != nil || res.Code != 200 {
			return nil, errors.New("webvpn get score detail error")
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
		if err != nil {
			return nil, err
		}
		result := make(map[string]string)
		doc.Find("td").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "：") {
				arr := strings.Split(s.Text(), "：")
				result[arr[0]] = arr[1]
			}
		})
		return result, nil
	})
	if err != nil {
		return nil, err
	}
	for i, detail := range details {
		for k, v := range detail {
			if _, ok := head_map[k]; !ok {
				head_map[k] = len(head)
				head = append(head, k)
			}
			data[jobs[i].index][k] = v
		}
	}
	// 将字典转为表格
	table := make([][]string, 0)
	table = append(table, make([]string, len(head)))
	table[0] = head
	for i, row := range data {
		table = append(table, make([]string, len(head)))
		for k, v := range row {
			table[i+1][head_map[k]] = v
		}
	}
	return table, nil
}

func GetReport(cookie string) ([]string, error) {
	return GetReportContext(context.Background(), cookie)
}

func GetReportContext(ctx context.Context, cookie string) ([]string, error) {
	return NewClient().GetReportContext(ctx, cookie)
}

func (c *Client) GetReport(cookie string) ([]string, error) {
	return c.GetReportContext(context.Background(), cookie)
}

func (c *Client) GetReportContext(ctx context.Context, cookie string) ([]string, error) {
	// 预登录
	_, err := c.preCheck("https://webvpn.bit.edu.cn/http/77726476706e69737468656265737421a1a70fcc69682601265c/cjd/Account/ExternalLogin", cookie)
	if err != nil {
		return nil, err
	}
	// 获取成绩单
	res, err := request.GetWithClient(c, "https://webvpn.bit.edu.cn/http/77726476706e69737468656265737421a1a70fcc69682601265c/cjd/ScoreReport2/Index?GPA=1", map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return nil, errors.New("webvpn get report error")
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0)
	doc.Find("main img").Each(func(i int, s *goquery.Selection) {
		urls = append(urls, s.AttrOr("src", ""))
	})
	if len(urls) == 0 {
		if strings.Contains(res.Text, "帐号登录或动态码登录") {
			return nil, ErrCookieInvalid
		}
		return nil, fmt.Errorf("webvpn report images not found: upstream=%s", res.URL)
	}

	return parallelMap(ctx, urls, func(ctx context.Context, url string) (string, error) {
		res, err := request.GetWithClientContext(ctx, c, "https://webvpn.bit.edu.cn"+url, map[string]string{"Cookie": cookie})
		if err != nil || res.Code != 200 {
			return "", errors.New("webvpn get report img error")
		}
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(res.Content), nil
	})
}

// 获取课程历史 接口同查询成绩详情
func GetCourseHistory(course_number string, term string, cookie string) (map[string]string, error) {
	course_history_url := fmt.Sprintf("https://webvpn.bit.edu.cn/http/77726476706e69737468656265737421fae04c8f69326144300d8db9d6562d/jsxsd/kscj/cjfx?kch=%s&xnxq01id=%s", course_number, term)
	mp := make(map[string]string)
	res, err := request.Get(course_history_url, map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
	if err != nil {
		return nil, err
	}
	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "：") {
			arr := strings.Split(s.Text(), "：")
			mp[arr[0]] = arr[1]
		}
	})
	return mp, nil
}
