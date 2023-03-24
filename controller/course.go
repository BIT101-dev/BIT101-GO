/*
 * @Author: flwfdd
 * @Date: 2023-03-23 16:07:43
 * @LastEditTime: 2023-03-25 01:51:00
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/nlp"
	"BIT101-GO/util/saver"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

// 获取课程列表请求结构
type CourseListQuery struct {
	Search string `form:"search"`
	Order  string `form:"order" default:"new"` //like | comment | rate | new
	Page   int    `form:"page" default:"0"`
}

// 获取课程列表
func CourseList(c *gin.Context) {
	var query CourseListQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	var courses []database.Course
	q := database.DB.Model(&database.Course{})
	// 搜索
	if query.Search != "" {
		query_l := nlp.CutAll(query.Search)
		query_l = append(query_l, nlp.CutForSearch(query.Search)...)
		if query.Order == "search" {
			q = q.Scopes(database.SearchText(query_l))
		} else {
			q = q.Scopes(database.FilterText(query_l))
		}
	}
	// 排序
	if query.Order != "search" {
		if query.Order == "comment" {
			q = q.Order("comment_num DESC")
		} else if query.Order == "like" {
			q = q.Order("like_num DESC")
		} else if query.Order == "rate" {
			q = q.Order("rate DESC")
		} else { //默认new
			q = q.Order("updated_at DESC")
		}
	}
	// 分页
	page_size := int(config.Config.CoursePageSize)
	q = q.Offset(query.Page * page_size).Limit(page_size)
	if err := q.Find(&courses).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	c.JSON(200, courses)
}

// 获取课程详情返回结构
type CourseInfoResponse struct {
	database.Course
	Like bool `json:"like"`
}

// 获取课程详情
func CourseInfo(c *gin.Context) {
	var course database.Course
	if err := database.DB.Where("id = ?", c.Param("id")).Limit(1).Find(&course).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if course.ID == 0 {
		c.JSON(500, gin.H{"msg": "课程不存在Orz"})
		return
	}
	c.JSON(200, CourseInfoResponse{
		Course: course,
		Like:   CheckLike(fmt.Sprintf("course%v", course.ID), c.GetUint("uid_uint")),
	})
}

// 点赞
func CourseOnLike(id string, delta int) (uint, error) {
	var course database.Course
	if err := database.DB.Limit(1).Find(&course, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if course.ID == 0 {
		return 0, errors.New("课程不存在Orz")
	}
	course.LikeNum = uint(int(course.LikeNum) + delta)
	if err := database.DB.Save(&course).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	return course.LikeNum, nil
}

// 评论
func CourseOnComment(id string, delta_num int, delta_rate int) (uint, error) {
	var course database.Course
	if err := database.DB.Limit(1).Find(&course, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if course.ID == 0 {
		return 0, errors.New("课程不存在Orz")
	}
	course.CommentNum = uint(int(course.CommentNum) + delta_num)
	course.RateSum = uint(int(course.RateSum) + delta_rate)
	if course.RateSum == 0 {
		course.Rate = 0
	} else {
		course.Rate = float64(course.RateSum) / float64(course.CommentNum)
	}
	if err := database.DB.Save(&course).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	return course.CommentNum, nil
}

// 获取课程资料上传链接请求结构
type CourseUploadUrlQuery struct {
	Number string `form:"number" binding:"required"` //课程编号
	Name   string `form:"name" binding:"required"`   //文件名
	Type   string `form:"type" default:"other"`      //资料类型
}

// 获取课程资料上传链接
func CourseUploadUrl(c *gin.Context) {
	var query CourseUploadUrlQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	var course database.Course
	if err := database.DB.Where("number = ?", query.Number).Limit(1).Find(&course).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if course.ID == 0 {
		c.JSON(500, gin.H{"msg": "课程不存在Orz"})
		return
	}
	tp := query.Type
	if tp != "book" && tp != "ppt" && tp != "exam" {
		tp = "other"
	}
	url, err := saver.OneDriveGetUploadUrl(fmt.Sprintf("course/%v-%v/%v/%v", course.Name, course.Number, tp, query.Name))
	if err != nil {
		c.JSON(500, gin.H{"msg": "获取上传链接失败Orz"})
		return
	}

	var log database.CourseUploadLog
	if err := database.DB.Where("course_number = ? AND uid = ? AND type = ? AND name = ?", course.Number, c.GetString("uid"), tp, query.Name).Limit(1).Find(&log).Error; err != nil {
		c.JSON(500, gin.H{"msg": "数据库错误Orz"})
		return
	}
	if log.ID == 0 {
		log = database.CourseUploadLog{
			Uid:          c.GetUint("uid_uint"),
			CourseNumber: course.Number,
			CourseName:   course.Name,
			Type:         tp,
			Name:         query.Name,
		}
		database.DB.Create(&log)
	}
	c.JSON(200, gin.H{"url": url, "id": log.ID})
}

// 上报课程资料上传记录请求结构
type CourseUploadLogQuery struct {
	ID  uint   `json:"id" binding:"required"` //上传记录ID
	Msg string `json:"msg"`                   //上传备注
}

// 上报课程资料上传记录
func CourseUploadLog(c *gin.Context) {
	var query CourseUploadLogQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 启用事务以应对高并发
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			c.JSON(500, gin.H{"msg": "数据库错误Orz"})
			tx.Rollback()
		}
	}()
	var log database.CourseUploadLog
	if err := tx.Where("id = ?", query.ID).Limit(1).Find(&log).Error; err != nil {
		panic(err)
	}
	if log.ID == 0 {
		c.JSON(500, gin.H{"msg": "上传记录不存在Orz"})
		return
	}
	if log.Uid != c.GetUint("uid_uint") {
		c.JSON(500, gin.H{"msg": "上传记录不属于你Orz"})
		return
	}
	log.Msg = query.Msg
	log.Finish = true
	if err := tx.Save(&log).Error; err != nil {
		panic(err)
	}

	// 更新README
	var readme database.CourseUploadReadme
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("course_number = ?", log.CourseNumber).Limit(1).Find(&readme).Error; err != nil {
		panic(err)
	}
	if readme.ID == 0 {
		readme = database.CourseUploadReadme{
			CourseNumber: log.CourseNumber,
		}
	}
	if readme.Text == "" {
		readme.Text = fmt.Sprintf(readme_template, log.CourseName, log.CourseNumber, config.Config.MainUrl, log.CourseName, config.Config.MainUrl, log.CourseNumber)
	}
	readme.Text += fmt.Sprintf(log_template, log.Type, log.Name, time.Now().Format("2006-01-02 15:04:05"), log.Msg)
	if err := tx.Save(&readme).Error; err != nil {
		panic(err)
	}
	if err := tx.Commit().Error; err != nil {
		panic(err)
	}
	saver.OneDriveUploadFile(fmt.Sprintf("course/%v-%v/README.md", log.CourseName, log.CourseNumber), []byte(readme.Text))
	c.JSON(200, gin.H{"msg": "上传成功OvO"})
}

// course_name number main_url course_name main_url number
var readme_template = `
# BIT101 %v 课程共享资料
> 课程编号：%v
本页面由[BIT101](%v)维护，[点击查找 %v 课程及评价](%v/#/course/?search=%v)
## 类别说明
* 书籍(book)：教科书、课程相关电子书等
* 课件(ppt)：PPT什么的
* 考试(exam)：考试相关的往年题、复习资料等
* 其他(other)：兜底条款，比如作业资料....？
## 文件上传记录和说明
`

// /{type}/{name}@{time} {msg}
var log_template = "\n* `/%v/%v`@`%v` %v"
