/*
 * @Author: flwfdd
 * @Date: 2023-03-21 17:34:55
 * @LastEditTime: 2023-03-21 22:56:57
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type PaperGetResponse struct {
	database.Paper
	UpdateUser gin.H `json:"update_user"`
	Like       bool  `json:"like"`
	Own        bool  `json:"own"`
}

// 获取文章
func PaperGet(c *gin.Context) {
	var paper database.Paper
	database.DB.Limit(1).Find(&paper, "id = ?", c.Param("id"))
	if paper.ID == 0 {
		c.JSON(500, gin.H{"msg": "文章不存在Orz"})
		return
	}
	var uid string
	if paper.Anonymous {
		uid = "-1"
	} else {
		uid = fmt.Sprint(paper.UpdateUid)
	}

	var res = PaperGetResponse{
		Paper:      paper,
		UpdateUser: GetUser(uid),
		Own:        paper.CreateUid == c.GetUint("uid"),
	}
	c.JSON(200, res)
}

// 新建文章请求接口
type PaperPostRequest struct {
	Title      string `json:"title" binding:"required"`
	Intro      string `json:"intro" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Anonymous  bool   `json:"anonymous" default:"false"`
	PublicEdit bool   `json:"public_edit" default:"true"`
}

// 新建文章
func PaperPost(c *gin.Context) {
	var query PaperPostRequest
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	var paper = database.Paper{
		Title:      query.Title,
		Intro:      query.Intro,
		Content:    query.Content,
		CreateUid:  c.GetUint("uid_uint"),
		UpdateUid:  c.GetUint("uid_uint"),
		Anonymous:  query.Anonymous,
		PublicEdit: query.PublicEdit,
		LikeNum:    0,
		CommentNum: 0,
	}
	database.DB.Create(&paper)
	pushHistory(&paper)
	c.JSON(200, gin.H{"msg": "发表成功OvO", "id": paper.ID})
}

// 修改文章请求接口
type PaperPutRequest struct {
	Title      string  `json:"title" binding:"required"`
	Intro      string  `json:"intro" binding:"required"`
	Content    string  `json:"content" binding:"required"`
	Anonymous  bool    `json:"anonymous" default:"false"`
	PublicEdit bool    `json:"public_edit" default:"true"`
	LastTime   float64 `json:"last_time" binding:"required"`
}

// 修改文章
func PaperPut(c *gin.Context) {
	var query PaperPutRequest
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	var paper database.Paper
	database.DB.Limit(1).Find(&paper, "id = ?", c.Param("id"))
	if paper.ID == 0 {
		c.JSON(500, gin.H{"msg": "文章不存在Orz"})
		return
	}
	if paper.CreateUid != c.GetUint("uid_uint") && !paper.PublicEdit && !c.GetBool("admin") {
		c.JSON(500, gin.H{"msg": "没有编辑权限Orz"})
		return
	}
	if paper.UpdatedAt.Unix() > int64(query.LastTime) {
		c.JSON(500, gin.H{"msg": "请基于最新版本编辑Orz"})
		return
	}
	paper.Title = query.Title
	paper.Intro = query.Intro
	paper.Content = query.Content
	paper.Anonymous = query.Anonymous
	paper.UpdateUid = c.GetUint("uid_uint")
	if paper.CreateUid == c.GetUint("uid_uint") || c.GetBool("admin") {
		paper.PublicEdit = query.PublicEdit
	}
	database.DB.Save(&paper)
	pushHistory(&paper)
	c.JSON(200, gin.H{"msg": "编辑成功OvO"})
}

// 添加到编辑记录
func pushHistory(paper *database.Paper) {
	history := database.PaperHistory{
		Pid:       paper.ID,
		Title:     paper.Title,
		Intro:     paper.Intro,
		Content:   paper.Content,
		Uid:       paper.UpdateUid,
		Anonymous: paper.Anonymous,
	}
	database.DB.Create(&history)
}

// 获取文章列表请求结构
type PaperListQuery struct {
	Search string `form:"search"`
	Order  string `form:"order" default:"id"` //rand | new | like
	Page   int    `form:"page" default:"0"`
}

// 获取文章列表返回结构
type PaperListResponseItem struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	Intro      string    `json:"intro"`
	LikeNum    uint      `json:"like_num"`
	CommentNum uint      `json:"comment_num"`
	UpdateTime time.Time `json:"update_time"`
}

// 获取文章列表
func PaperList(c *gin.Context) {
	var query PaperListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	var papers []database.Paper
	q := database.DB.Model(&database.Paper{}).Select("id,title,intro,updated_at,like_num,comment_num")
	// 搜索
	if query.Search != "" {
		q = q.Where("title LIKE ?", "%"+query.Search+"%").Or("intro LIKE ?", "%"+query.Search+"%").Or("content LIKE ?", "%"+query.Search+"%")
	}
	// 排序
	if query.Order == "rand" {
		q = q.Order("random()")
	} else if query.Order == "like" {
		q = q.Order("like_num DESC")
	} else { //默认new
		q = q.Order("updated_at DESC")
	}
	// 分页
	page_size := config.Config.PaperPageSize
	q = q.Offset(query.Page * page_size).Limit(page_size)
	q.Find(&papers)
	var res []PaperListResponseItem
	for _, paper := range papers {
		res = append(res, PaperListResponseItem{
			ID:         paper.ID,
			Title:      paper.Title,
			Intro:      paper.Intro,
			LikeNum:    paper.LikeNum,
			CommentNum: paper.CommentNum,
			UpdateTime: paper.UpdatedAt,
		})
	}

	c.JSON(200, res)
}
