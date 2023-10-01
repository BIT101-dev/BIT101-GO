/*
 * @Author: flwfdd
 * @Date: 2023-03-23 16:07:43
 * @LastEditTime: 2023-05-17 16:54:23
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/controller/webvpn"
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/saver"
	"BIT101-GO/util/search"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

// CourseListQuery 获取课程列表请求结构
type CourseListQuery struct {
	Search string `form:"search"`
	Order  string `form:"order"` //like | comment | rate | new
	Page   int    `form:"page"`
}

// CourseList 获取课程列表
func CourseList(c *gin.Context) {
	var query CourseListQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	// 排序
	var order []string
	if query.Order != "search" {
		if query.Order == "comment" {
			order = append(order, "comment_num:desc")
		} else if query.Order == "like" {
			order = append(order, "like_num:desc")
		} else if query.Order == "rate" {
			order = append(order, "rate:desc")
		} else { //默认new
			order = append(order, "update_time:desc")
		}
	}

	fmt.Println("CourseList search:", query.Search, "order:", order, "page:", query.Page)
	var courses []database.Course
	err := search.Search(&courses, "course", query.Search, order, int64(query.Page))
	if err != nil {
		c.JSON(500, gin.H{"msg": "搜索失败Orz"})
		return
	}
	c.JSON(200, courses)
}

// CourseInfoResponse 获取课程详情返回结构
type CourseInfoResponse struct {
	database.Course
	Like bool `json:"like"`
}

// CourseInfo 获取课程详情
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

// CourseOnLike 点赞
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
	if err := search.Update("course", []map[string]interface{}{database.StructToMap(course)}); err != nil {
		return 0, errors.New("search同步失败Orz")
	}
	return course.LikeNum, nil
}

// CourseOnComment 评论
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
	if err := search.Update("course", []map[string]interface{}{database.StructToMap(course)}); err != nil {
		return 0, errors.New("search同步失败Orz")
	}
	return course.CommentNum, nil
}

// CourseUploadUrlQuery 获取课程资料上传链接请求结构
type CourseUploadUrlQuery struct {
	Number string `form:"number" binding:"required"` //课程编号
	Name   string `form:"name" binding:"required"`   //文件名
	Type   string `form:"type"`                      //资料类型 默认other
}

// CourseUploadUrl 获取课程资料上传链接
func CourseUploadUrl(c *gin.Context) {
	var query CourseUploadUrlQuery
	query.Type = "other"
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

// CourseUploadLogQuery 上报课程资料上传记录请求结构
type CourseUploadLogQuery struct {
	ID  uint   `json:"id" binding:"required"` //上传记录ID
	Msg string `json:"msg"`                   //上传备注
}

// CourseUploadLog 上报课程资料上传记录
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

// 上课时间表 节次 上课/下课时间 时,分
var time_table = [][][]int{
	{{8, 0}, {8, 45}},
	{{8, 50}, {9, 35}},
	{{9, 55}, {10, 40}},
	{{10, 45}, {11, 30}},
	{{11, 35}, {12, 20}},
	{{13, 20}, {14, 5}},
	{{14, 10}, {14, 55}},
	{{15, 15}, {16, 0}},
	{{16, 5}, {16, 50}},
	{{16, 55}, {17, 40}},
	{{18, 30}, {19, 15}},
	{{19, 20}, {20, 5}},
	{{20, 10}, {20, 55}},
}

// CourseScheduleQuery 获取课程表请求结构
type CourseScheduleQuery struct {
	Term string `form:"term"` // 学期
}

// CourseSchedule 获取课程表
func CourseSchedule(c *gin.Context) {
	var query CourseScheduleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	cookie := c.Request.Header.Get("webvpn-cookie")
	if cookie == "" {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}

	schedule, err := webvpn.GetSchedule(cookie, query.Term)
	if err != nil {
		c.JSON(500, gin.H{"msg": "获取课程表失败Orz"})
		return
	}
	first_day, err := time.Parse("2006-01-02", schedule.FirstDay)
	if err != nil {
		c.JSON(500, gin.H{"msg": "解析时间失败Orz"})
		return
	}

	class_ct := 0
	time_ct := 0

	calendar := ""
	calendar += "BEGIN:VCALENDAR\n"
	calendar += "VERSION:2.0\n"
	calendar += fmt.Sprintf("PRODID:BIT101 %v\n", time.Now())
	calendar += "TZID:Asia/Shanghai\n"
	calendar += "X-WR-CALNAME:BIT101课程表\n"

	for _, course := range schedule.Data {
		for week, ok := range course.SKZC {
			if ok == '1' {
				calendar += "BEGIN:VEVENT\n"
				calendar += fmt.Sprintf("SUMMARY:%v\n", course.KCM)
				calendar += fmt.Sprintf("LOCATION:%v\n", course.JASMC)
				calendar += fmt.Sprintf("DESCRIPTION:%v | %v\n", course.SKJS, course.YPSJDD)
				start_time_table := time_table[course.KSJC-1][0]
				start_time := first_day.AddDate(0, 0, week*7+course.SKXQ-1).Add(time.Minute * time.Duration(60*start_time_table[0]+start_time_table[1])).Format("20060102T150405")
				calendar += fmt.Sprintf("DTSTART:%v\n", start_time)
				end_time_table := time_table[course.JSJC-1][1]
				end_time := first_day.AddDate(0, 0, week*7+course.SKXQ-1).Add(time.Minute * time.Duration(60*end_time_table[0]+end_time_table[1])).Format("20060102T150405")
				calendar += fmt.Sprintf("DTEND:%v\n", end_time)
				calendar += fmt.Sprintf("UID:%v\n", uuid.New())
				calendar += "END:VEVENT\n"

				class_ct++
				time_ct += (course.JSJC - course.KSJC + 1) * 45
			}
		}
	}

	calendar += "END:VCALENDAR\n"

	url, err := saver.Save(fmt.Sprintf("/tmp/%v.ics", uuid.New()), []byte(calendar))
	if err != nil {
		c.JSON(500, gin.H{"msg": "保存课程表失败Orz"})
		return
	}
	c.JSON(200, gin.H{"url": url, "note": fmt.Sprintf("一共添加了%v学期的%v节课，合计坐牢时间%v小时（雾", schedule.Term, class_ct, float64(time_ct)/60), "msg": "获取成功OvO"})
}
