/*
 * @Author: flwfdd
 * @Date: 2023-03-15 18:24:54
 * @LastEditTime: 2023-09-23 23:37:15
 * @Description: webvpn成绩模块
 */
package webvpn

import (
	"BIT101-GO/util/request"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetScore(cookie string, detail bool) ([][]string, error) {
	score_url := "https://webvpn.bit.edu.cn/http/77726476706e69737468656265737421fae04c8f69326144300d8db9d6562d/jsxsd/kscj/cjcx_list"
	// 预登录
	_, err := preCheck(score_url, cookie)
	if err != nil {
		return nil, err
	}
	// 获取成绩列表
	res, err := request.Get(score_url, map[string]string{"Cookie": cookie})
	if err != nil || res.Code != 200 {
		return nil, errors.New("webvpn get_score error")
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
	if err != nil {
		return nil, err
	}
	rows := doc.Find("#dataList tr").Nodes
	head := make([]string, 0)
	head_map := make(map[string]int)
	data := make([]map[string]string, 0)
	type chData struct {
		index int
		data  map[string]string
	}
	ch := make(chan *chData, 24)
	ch_count := 0
	ctx, cancel := context.WithCancel(context.Background())
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
					ch_count++
					// 开启协程查询详细信息 传回一个map[string]string 出错则传回nil
					go func(ctx context.Context, index int, url string) {
						mp := make(map[string]string)
						res, err := request.Get(url, map[string]string{"Cookie": cookie})
						if err != nil || res.Code != 200 {
							select {
							case <-ctx.Done():
							default:
								ch <- nil
							}
							return
						}
						doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Text))
						if err != nil {
							select {
							case <-ctx.Done():
							default:
								ch <- nil
							}
							return
						}
						doc.Find("td").Each(func(i int, s *goquery.Selection) {
							if strings.Contains(s.Text(), "：") {
								arr := strings.Split(s.Text(), "：")
								mp[arr[0]] = arr[1]
							}
						})
						select {
						case <-ctx.Done():
						default:
							ch <- &chData{index, mp}
						}
					}(ctx, row_index, detail_url)
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
	// 等待详细信息
	for i := 0; i < ch_count; i++ {
		detail_data := <-ch
		if detail_data == nil {
			cancel()
			return nil, errors.New("webvpn get score detail error")
		}
		for k, v := range detail_data.data {
			if _, ok := head_map[k]; !ok {
				head_map[k] = len(head)
				head = append(head, k)
			}
			data[detail_data.index][k] = v
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
	cancel()
	return table, nil
}

func GetReport(cookie string) ([]string, error) {
	// 预登录
	_, err := preCheck("https://webvpn.bit.edu.cn/http/77726476706e69737468656265737421a1a70fcc69682601265c/cjd/Account/ExternalLogin", cookie)
	if err != nil {
		return nil, err
	}
	// 获取成绩单
	res, err := request.Get("https://webvpn.bit.edu.cn/http/77726476706e69737468656265737421a1a70fcc69682601265c/cjd/ScoreReport2/Index?GPA=1", map[string]string{"Cookie": cookie})
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

	// 并发获取成绩单图片
	type chData struct {
		index int
		data  string
	}
	ch := make(chan *chData, 24)
	ctx, cancel := context.WithCancel(context.Background())
	imgs := make([]string, len(urls))
	for i, url := range urls {
		go func(ctx context.Context, index int, url string) {
			res, err := request.Get(url, map[string]string{"Cookie": cookie})
			if err != nil || res.Code != 200 {
				select {
				case <-ctx.Done():
				default:
					ch <- nil
				}
				return
			}
			select {
			case <-ctx.Done():
			default:
				ch <- &chData{index, base64.StdEncoding.EncodeToString(res.Content)}
			}
		}(ctx, i, "https://webvpn.bit.edu.cn"+url)
	}

	// 接收图片
	for i := 0; i < len(urls); i++ {
		data := <-ch
		if data == nil {
			cancel()
			return nil, errors.New("webvpn get report img error")
		}
		imgs[data.index] = "data:image/png;base64," + data.data
	}
	cancel()
	return imgs, nil
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
