/*
 * @Author: flwfdd
 * @Date: 2023-03-22 16:35:17
 * @LastEditTime: 2023-03-23 01:20:50
 * @Description: _(:з」∠)_
 */
package database

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// postgresql 中的 tsvector 类型
type Tsvector struct {
	A []string `json:"a"`
	B []string `json:"b"`
	C []string `json:"c"`
	D []string `json:"d"`
}

func (t *Tsvector) Scan(value interface{}) error {
	return nil
}

func (t Tsvector) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL:  `setweight(to_tsvector('simple',?),'A') || setweight(to_tsvector('simple',?),'B')  || setweight(to_tsvector('simple',?),'C') || setweight(to_tsvector('simple',?),'D')`,
		Vars: []interface{}{strings.Join(t.A, " "), strings.Join(t.B, " "), strings.Join(t.C, " "), strings.Join(t.D, " ")},
	}
}

func (Tsvector) GormDataType() string {
	return "tsvector"
}

// 简便地筛选并搜索
func SearchText(query []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		query_s := ""
		for _, word := range query {
			word = strings.ReplaceAll(word, "'", "")
			word = strings.ReplaceAll(word, "|", "")
			word = strings.ReplaceAll(word, "\\", "")
			if word == "" {
				continue
			}
			if query_s != "" {
				query_s += "|"
			}
			query_s += word
		}
		return db.Where("tsv @@ ?::tsquery", query_s).Clauses(clause.OrderBy{
			Expression: clause.Expr{SQL: "ts_rank(tsv,?::tsquery) DESC", Vars: []interface{}{query_s}},
		})
	}
}
