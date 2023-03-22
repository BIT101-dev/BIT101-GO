package nlp

import (
	"encoding/json"
	"errors"
	"html"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type block struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type paragraph struct {
	Text string `json:"text"`
}

type header struct {
	Text string `json:"text"`
}

type list struct {
	Items []string `json:"items"`
}

type quote struct {
	Text    string `json:"text"`
	Caption string `json:"caption"`
}

type code struct {
	Code string `json:"code"`
}

type table struct {
	Content [][]string `json:"content"`
}

// 提取HTML中的文本
func cleanText(s string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html.UnescapeString(s)))
	if err != nil {
		return s
	}
	return doc.Text()
}

// 解析EditorJS的文本
func ParseEditorJS(raw string) (string, error) {
	out := ""
	var dic struct {
		Blocks []block `json:"blocks"`
	}
	if json.Unmarshal([]byte(raw), &dic) != nil {
		return "", errors.New("parse json error")
	}
	for _, block := range dic.Blocks {
		switch block.Type {
		case "paragraph":
			var p paragraph
			if json.Unmarshal([]byte(block.Data), &p) != nil {
				return "", errors.New("parse paragraph error")
			}
			s := cleanText(p.Text)
			out += s + " "
		case "header":
			var h header
			if json.Unmarshal([]byte(block.Data), &h) != nil {
				return "", errors.New("parse header error")
			}
			s := cleanText(h.Text)
			out += s + " "
		case "list":
			var l list
			if json.Unmarshal([]byte(block.Data), &l) != nil {
				return "", errors.New("parse list error")
			}
			for _, item := range l.Items {
				s := cleanText(item)
				out += s + " "
			}
		case "quote":
			var q quote
			if json.Unmarshal([]byte(block.Data), &q) != nil {
				return "", errors.New("parse quote error")
			}
			s := cleanText(q.Text + " " + q.Caption)
			out += s + " "
		case "code":
			var c code
			if json.Unmarshal([]byte(block.Data), &c) != nil {
				return "", errors.New("parse code error")
			}
			s := cleanText(c.Code)
			out += s + " "
		case "table":
			var t table
			if json.Unmarshal([]byte(block.Data), &t) != nil {
				return "", errors.New("parse table error")
			}
			for _, row := range t.Content {
				for _, col := range row {
					s := cleanText(col)
					out += s + " "
				}
			}
		}
	}
	return out, nil
}
