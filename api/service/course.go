/*
 * @Author: flwfdd
 * @Date: 2025-03-09 23:38:54
 * @LastEditTime: 2025-03-19 02:08:57
 * @Description: _(:з」∠)_
 */
package service

import (
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"BIT101-GO/pkg/saver"
	"BIT101-GO/pkg/webvpn"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CourseService 课程服务
type CourseService struct {
	reactionSvc     *ReactionService
	meiliSearchSvc  *MeilisearchService
	subscriptionSvc *SubscriptionService
}

// NewCourseService 创建课程服务
func NewCourseService(reactionSvc *ReactionService, meiliSearchSvc *MeilisearchService, subscriptionSvc *SubscriptionService) *CourseService {
	s := CourseService{
		reactionSvc:     reactionSvc,
		meiliSearchSvc:  meiliSearchSvc,
		subscriptionSvc: subscriptionSvc,
	}
	types.RegisterObjHandler(&s)
	return &s
}

/* ObjHandler */

// IsExist 判断课程是否存在
func (s *CourseService) IsExist(id uint) bool {
	_, err := s.getCourse(id)
	return err == nil
}

// GetObjType 获取对象类型
func (s *CourseService) GetObjType() string {
	return "course"
}

// GetUid 获取课程作者
func (s *CourseService) GetUid(id uint) (uint, error) {
	return 0, errors.New("课程没有作者Orz")
}

// GetText 获取课程简介
func (s *CourseService) GetText(id uint) (string, error) {
	course, err := s.getCourse(id)
	if err != nil {
		return "", err
	}
	return course.Name + " " + course.TeachersName, nil
}

// LikeHandler 点赞
func (s *CourseService) LikeHandler(tx *gorm.DB, id uint, delta int, uid uint) (likeNum uint, err error) {
	var course database.Course
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&course, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if course.ID == 0 {
		return 0, errors.New("课程不存在Orz")
	}
	course.LikeNum = uint(int(course.LikeNum) + delta)
	if err := tx.Save(&course).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	go func() {
		s.meiliSearchSvc.Add(s.GetObjType(), course)
	}()
	return course.LikeNum, nil
}

// CommentHandler 评论
func (s *CourseService) CommentHandler(tx *gorm.DB, id uint, comment database.Comment, delta int, uid uint) (commentNum uint, err error) {
	var course database.Course
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&course, "id = ?", id).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if course.ID == 0 {
		return 0, errors.New("课程不存在Orz")
	}
	deltaRate := int(comment.Rate) * delta
	if deltaRate == 0 || deltaRate < -10 || deltaRate > 10 {
		return 0, errors.New("评分错误Orz")
	}
	course.CommentNum = uint(int(course.CommentNum) + delta)
	course.RateSum = uint(int(course.RateSum) + deltaRate)
	if course.RateSum == 0 {
		course.Rate = 0
	} else {
		course.Rate = float64(course.RateSum) / float64(course.CommentNum)
	}
	if err := tx.Save(&course).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	go func() {
		s.meiliSearchSvc.Add(s.GetObjType(), course)
		if delta > 0 {
			s.subscriptionSvc.NotifySubscription(fmt.Sprintf("%s%v", s.GetObjType(), course.ID), database.SubscriptionLevelComment, course.Name+" 有新评价："+comment.Text)
		}
	}()
	return course.CommentNum, nil
}

// getCourse 获取课程
func (s *CourseService) getCourse(id uint) (database.Course, error) {
	var course database.Course
	if err := database.DB.Where("id = ?", id).Limit(1).Find(&course).Error; err != nil {
		return database.Course{}, errors.New("数据库错误Orz")
	}
	if course.ID == 0 {
		return database.Course{}, errors.New("课程不存在Orz")
	}
	return course, nil
}

// GetCourses 获取课程列表
func (s *CourseService) GetCourses(keyword string, order string, page uint) ([]database.Course, error) {
	// 排序
	var orders []string
	switch order {
	case "comment":
		orders = append(orders, "comment_num:desc")
	case "like":
		orders = append(orders, "like_num:desc")
	case "rate":
		orders = append(orders, "rate:desc")
		orders = append(orders, "comment_num:desc")
	default:
		orders = append(orders, "update_time:desc")
	}

	courses := make([]database.Course, 0)
	err := s.meiliSearchSvc.Search(&courses, s.GetObjType(), keyword, page, config.Get().CoursePageSize, orders, nil)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

// GetCourseAPI 获取课程详情
func (s *CourseService) GetCourseAPI(id uint, uid uint) (types.CourseAPI, error) {
	var course database.Course
	if err := database.DB.Where("id = ?", id).Limit(1).Find(&course).Error; err != nil {
		return types.CourseAPI{}, errors.New("数据库错误Orz")
	}
	if course.ID == 0 {
		return types.CourseAPI{}, errors.New("课程不存在Orz")
	}
	subscription, err := s.subscriptionSvc.GetSubscriptionLevel(uid, fmt.Sprintf("%s%v", s.GetObjType(), course.ID))
	if err != nil {
		slog.Error("course: get subscription level failed", "error", err.Error())
		subscription = 0
	}

	return types.CourseAPI{
		Course:       course,
		Like:         s.reactionSvc.CheckLike(fmt.Sprintf("course%d", course.ID), uid),
		Subscription: subscription,
	}, nil
}

// GetUploadUrl 获取课程资料上传链接
func (s *CourseService) GetUploadUrl(courseNumber, fileName, typ string, uid uint) (url string, id uint, err error) {
	var course database.Course
	if err = database.DB.Where("number = ?", courseNumber).Limit(1).Find(&course).Error; err != nil {
		return "", 0, errors.New("数据库错误Orz")
	}
	if course.ID == 0 {
		return "", 0, errors.New("课程不存在Orz")
	}
	if typ != "book" && typ != "ppt" && typ != "exam" {
		typ = "other"
	}
	url, err = saver.OneDriveGetUploadUrl(fmt.Sprintf("course/%v-%v/%v/%v", course.Name, course.Number, typ, fileName))
	if err != nil {
		return "", 0, errors.New("获取上传链接失败Orz")
	}

	var log database.CourseUploadLog
	if err := database.DB.Where("course_number = ? AND uid = ? AND type = ? AND name = ?", course.Number, uid, typ, fileName).Limit(1).Find(&log).Error; err != nil {
		return "", 0, errors.New("数据库错误Orz")
	}
	if log.ID == 0 {
		log = database.CourseUploadLog{
			Uid:          uid,
			CourseNumber: course.Number,
			CourseName:   course.Name,
			Type:         typ,
			Name:         fileName,
		}
		if err := database.DB.Create(&log).Error; err != nil {
			return "", 0, errors.New("数据库错误Orz")
		}
	}
	return url, log.ID, nil
}

// course_name number main_url course_name main_url number
var readmeTemplate = `
# BIT101 %v 课程共享资料
> 课程编号：%v
本页面由[BIT101](%v)维护，[点击查找 %v 课程及评价](%v/course/?search=%v)
## 类别说明
* 书籍(book)：教科书、课程相关电子书等
* 课件(ppt)：PPT什么的
* 考试(exam)：考试相关的往年题、复习资料等
* 其他(other)：兜底条款，比如作业资料....？
## 文件上传记录和说明
`

// /{type}/{name}@{time} {msg}
var logTemplate = "\n* `/%v/%v`@`%v` %v"

// LogUpload 记录上传
func (s *CourseService) LogUpload(id uint, msg string, uid uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var log database.CourseUploadLog
		if err := tx.Where("id = ?", id).Limit(1).Find(&log).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		if log.ID == 0 {
			return errors.New("上传记录不存在Orz")
		}
		if log.Uid != uid {
			return errors.New("上传记录不属于你Orz")
		}
		log.Msg = msg
		log.Finish = true
		if err := tx.Save(&log).Error; err != nil {
			return errors.New("数据库错误Orz")
		}

		// 更新README
		var readme database.CourseUploadReadme
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("course_number = ?", log.CourseNumber).Limit(1).Find(&readme).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		if readme.ID == 0 {
			readme = database.CourseUploadReadme{
				CourseNumber: log.CourseNumber,
			}
		}
		if readme.Text == "" {
			readme.Text = fmt.Sprintf(readmeTemplate, log.CourseName, log.CourseNumber, config.Get().MainUrl, log.CourseName, config.Get().MainUrl, log.CourseNumber)
		}
		readme.Text += fmt.Sprintf(logTemplate, log.Type, log.Name, time.Now().Format("2006-01-02 15:04:05"), log.Msg)
		if err := tx.Save(&readme).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		if err := tx.Commit().Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		saver.OneDriveUploadFile(fmt.Sprintf("course/%v-%v/README.md", log.CourseName, log.CourseNumber), []byte(readme.Text))
		return nil
	})
}

// GetCourseHistory 获取课程历史
func (s *CourseService) GetCourseHistory(courseNumber string) ([]types.CourseHistoryAPI, error) {
	var histories []database.CourseHistory
	if err := database.DB.Where("number = ?", courseNumber).Find(&histories).Error; err != nil {
		return nil, errors.New("数据库错误Orz")
	}
	historyAPI := make([]types.CourseHistoryAPI, 0)
	for _, history := range histories {
		historyAPI = append(historyAPI, types.CourseHistoryAPI{
			Term:       history.Term,
			AvgScore:   history.AvgScore,
			MaxScore:   history.MaxScore,
			StudentNum: history.StudentNum,
		})
	}
	return historyAPI, nil
}

// 上课时间表 节次 上课/下课时间 时,分
var timeTable = [][][]int{
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

// 建筑名称关键字与坐标的映射表
var buildingMap = map[string][2]float64{
	// 基础教室
	"综教A": {39.733193, 116.170654},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E7%BB%BC%E5%90%88%E6%95%99%E5%AD%A6%E5%A4%A7%E6%A5%BCA%E5%8F%B7%E6%A5%BC&auid=1118574501998747&ll=39.733193,116.170654&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E7%BB%BC%E5%90%88%E6%95%99%E5%AD%A6%E5%A4%A7%E6%A5%BCA%E5%8F%B7%E6%A5%BC&t=r

	"综教B": {39.733184, 116.171878},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E7%BB%BC%E5%90%88%E6%95%99%E5%AD%A6%E5%A4%A7%E6%A5%BCB%E5%8F%B7%E6%A5%BC&auid=1118574501801355&ll=39.733184,116.171878&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E7%BB%BC%E5%90%88%E6%95%99%E5%AD%A6%E5%A4%A7%E6%A5%BCB%E5%8F%B7%E6%A5%BC&t=r

	"理教楼": {39.730116, 116.171359},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF8%E5%8F%B7%E9%99%A2%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA&auid=1117160144852604&ll=39.730116,116.171359&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E7%90%86%E7%A7%91%E6%95%99%E5%AD%A6%E5%A4%A7%E6%A5%BC&t=r

	"理学A": {39.728886, 116.171800},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E9%AB%98%E6%95%99%E5%9B%AD%E5%8C%BA%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA&auid=1117160147079854&ll=39.728886,116.171800&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%8D%97%E6%A0%A1%E5%8C%BA%E7%90%86%E5%AD%A6%E6%A5%BCA&t=r

	"理学B": {39.729267, 116.171739},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E9%AB%98%E6%95%99%E5%9B%AD%E5%8C%BA%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA&auid=1118368566951329&ll=39.729267,116.171739&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%8D%97%E6%A0%A1%E5%8C%BA%E7%90%86%E5%AD%A6%E6%A5%BCB&t=r

	"理学C": {39.729633, 116.171778},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E7%90%86%E5%AD%A6%E6%A5%BCC%E5%8F%B7%E6%A5%BC&auid=1118574501764258&ll=39.729633,116.171778&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E7%90%86%E5%AD%A6%E6%A5%BCC%E5%8F%B7%E6%A5%BC&t=r

	"文萃楼A": {39.732606, 116.174479},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCA&auid=1118574501792819&ll=39.732606,116.174479&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCA&t=r

	"文萃楼B": {39.732217, 116.174489},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCB&auid=1118670288433274&ll=39.732217,116.174489&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCB&t=r

	"文萃楼C": {39.731655, 116.174267},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA-%E6%96%87%E8%90%83%E6%A5%BCC&auid=1118574501754538&ll=39.731655,116.174267&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA-%E6%96%87%E8%90%83%E6%A5%BCC&t=r

	"文萃楼D": {39.731670, 116.173885},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BAD%E5%8F%B7%E6%A5%BC&auid=1118574501744728&ll=39.731670,116.173885&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BAD%E5%8F%B7%E6%A5%BC&t=r

	"文萃楼E": {39.731669, 116.173429},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BAE%E5%8F%B7%E6%A5%BC&auid=1118574501764261&ll=39.731669,116.173429&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BAE%E5%8F%B7%E6%A5%BC&t=r

	"文萃楼F": {39.732060, 116.173821},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCF&auid=1118574501969227&ll=39.732060,116.173821&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCF&t=r

	"文萃楼G": {39.732216, 116.173101},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCG&auid=1118574501764260&ll=39.732216,116.173101&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCG&t=r

	"文萃楼H": {39.732995, 116.173098},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCH&auid=1118574501774062&ll=39.732995,116.173098&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCH&t=r

	"文萃楼I": {39.733083, 116.173866},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCI&auid=1118574501754537&ll=39.733083,116.173866&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCI&t=r

	"文萃楼J": {39.733518, 116.173408},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BAJ%E5%8F%B7%E6%A5%BC&auid=1118574501792821&ll=39.733518,116.173408&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BAJ%E5%8F%B7%E6%A5%BC&t=r

	"文萃楼K": {39.733440, 116.173841},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BAK%E5%8F%B7%E6%A5%BC&auid=1118574501900655&ll=39.733440,116.173841&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BAK%E5%8F%B7%E6%A5%BC&t=r

	"文萃楼L": {39.733488, 116.174220},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BAL%E5%8F%B7%E6%A5%BC&auid=1118574501998751&ll=39.733488,116.174220&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BAL%E5%8F%B7%E6%A5%BC&t=r

	"文萃楼M": {39.733058, 116.174525},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCM&auid=1118670288679343&ll=39.733058,116.174525&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E8%90%83%E6%A5%BCM&t=r

	// 体育课
	"良乡体育馆": {39.731844, 116.176544},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6(%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA)&auid=1118368915071108&ll=39.731844,116.176544&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%96%87%E4%BD%93%E4%B8%AD%E5%BF%83&t=r

	"北校区篮球场": {39.736357, 116.170721},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E9%AB%98%E6%95%99%E5%9B%AD%E5%8C%BA%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%8C%97%E6%A0%A1%E5%8C%BA&auid=1118368566943099&ll=39.736357,116.170721&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%8C%97%E6%A0%A1%E5%8C%BA%E7%AF%AE%E7%90%83%E5%9C%BA&t=r

	"南校区篮球场": {39.728026, 116.169304},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E9%95%87%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%8D%97%E6%A0%A1%E5%8C%BA&auid=1118368566963044&ll=39.728026,116.169304&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E5%8D%97%E6%A0%A1%E5%8C%BA%E7%AF%AE%E7%90%83%E5%9C%BA&t=r

	"南校区排球场": {39.727381, 116.169604}, // 没有,给的大头针
	// https://maps.apple.com/?auid=1118368566963044&ll=39.727381,116.169604&lsp=57879&q=%E5%B7%B2%E6%A0%87%E8%AE%B0%E4%BD%8D%E7%BD%AE&t=r

	"南校区足球场": {39.729583, 116.169174},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E9%95%87%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%8D%97%E6%A0%A1%E5%8C%BA&auid=1118368566964115&ll=39.729583,116.169174&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E5%8D%97%E6%A0%A1%E5%8C%BA%E8%B6%B3%E7%90%83%E5%9C%BA&t=r

	"南校区网球场": {39.727967, 116.168370},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E9%98%B3%E5%85%89%E5%8D%97%E5%A4%A7%E8%A1%97%E4%B8%8E%E6%98%8E%E7%90%86%E8%B7%AF%E4%BA%A4%E5%8F%89%E5%8F%A3%E4%B8%9C%E5%8C%97300%E7%B1%B3&auid=1118793571067609&ll=39.727967,116.168370&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E7%BD%91%E7%90%83%E5%9C%BA%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA&t=r

	"田径场主席台": {39.729490, 116.168474}, // 没有，给的大头针
	// https://maps.apple.com/?auid=1118368566964115&ll=39.729490,116.168474&lsp=57879&q=%E5%B7%B2%E6%A0%87%E8%AE%B0%E4%BD%8D%E7%BD%AE&t=r

	"疏桐园A": {39.728834, 116.167727},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E9%AB%98%E6%95%99%E5%9B%AD%E5%8C%BA%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%8D%97%E6%A0%A1%E5%8C%BA&auid=1118368545191793&ll=39.728834,116.167727&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%8D%97%E6%A0%A1%E5%8C%BA%E7%96%8F%E6%A1%90%E5%9B%AD%E5%AD%A6%E7%94%9F%E5%85%AC%E5%AF%93A%E5%BA%A7&t=r

	"游泳馆": {39.731755, 116.177294},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF%E4%B8%8E%E8%87%B4%E7%BE%8E%E5%8C%97%E8%A1%97%E4%BA%A4%E5%8F%89%E5%8F%A3%E8%A5%BF%E5%8C%97100%E7%B1%B3&auid=1118368951729133&ll=39.731755,116.177294&lsp=57879&q=%E5%8C%97%E7%90%86%E5%B7%A5%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E6%B8%B8%E6%B3%B3%E9%A6%86&t=r

	// 杂项
	"化学实验中心": {39.727976, 116.170456},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E9%AB%98%E6%95%99%E5%9B%AD%E5%8C%BA%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA&auid=1118574501764257&ll=39.727976,116.170456&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%8C%96%E5%AD%A6%E5%AE%9E%E9%AA%8C%E4%B8%AD%E5%BF%83&t=r

	"工训楼": {39.726286, 116.173760},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%B7%A5%E8%AE%AD%E6%A5%BC&auid=1118574501792818&ll=39.726286,116.173760&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E5%B7%A5%E8%AE%AD%E6%A5%BC&t=r

	"西山阻燃楼": {40.037061, 116.232287}, // 没找到，给的西山试验区的定位
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%B5%B7%E6%B7%80%E5%8C%BA%E5%86%B7%E6%B3%89%E4%B8%9C%E8%B7%AF16%E5%8F%B7&auid=1117160143024132&ll=40.037061,116.232287&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%A5%BF%E5%B1%B1%E5%AE%9E%E9%AA%8C%E5%8C%BA&t=r

	"物理实验中心": {39.729071, 116.170698},
	// https://maps.apple.com/?address=%E4%B8%AD%E5%9B%BD%E5%8C%97%E4%BA%AC%E5%B8%82%E6%88%BF%E5%B1%B1%E5%8C%BA%E8%89%AF%E4%B9%A1%E4%B8%9C%E8%B7%AF8%E5%8F%B7%E9%99%A2%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA&auid=1118368923032042&ll=39.729071,116.170698&lsp=57879&q=%E5%8C%97%E4%BA%AC%E7%90%86%E5%B7%A5%E5%A4%A7%E5%AD%A6%E8%89%AF%E4%B9%A1%E6%A0%A1%E5%8C%BA%E7%89%A9%E7%90%86%E5%AE%9E%E9%AA%8C%E4%B8%AD%E5%BF%83&t=r

}

// 根据教室名称获取坐标（返回纬度,经度）
func getBuildingCoord(jasmc string) (float64, float64, bool) {
	// 按关键字长度倒序排序，优先匹配更长更精确的关键字
	var keys []string
	for k := range buildingMap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	for _, k := range keys {
		if strings.Contains(jasmc, k) {
			coord := buildingMap[k]
			return coord[0], coord[1], true
		}
	}
	return 0, 0, false
}

// GetCourseSchedule 获取课程表
func (s *CourseService) GetCourseSchedule(cookie, term string) (url, note string, err error) {
	schedule, err := webvpn.GetSchedule(cookie, term)
	if err != nil {
		return "", "", err
	}
	firstDay, err := time.Parse("2006-01-02", schedule.FirstDay)
	if err != nil {
		return "", "", errors.New("解析日期失败Orz")
	}

	classCount := 0
	timeCount := 0

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
				calendar += fmt.Sprintf("LOCATION:%v\\n北京理工大学\n", course.JASMC)
				if lat, lng, ok := getBuildingCoord(course.JASMC); ok {
					calendar += fmt.Sprintf("X-APPLE-STRUCTURED-LOCATION;VALUE=URI;X-ADDRESS=\"北京理工大学\";X-TITLE=\"%v\":geo:%.6f,%.6f\n", course.JASMC, lat, lng)
				}
				calendar += fmt.Sprintf("DESCRIPTION:%v | %v\n", course.SKJS, course.YPSJDD)
				start_time_table := timeTable[course.KSJC-1][0]
				start_time := firstDay.AddDate(0, 0, week*7+course.SKXQ-1).Add(time.Minute * time.Duration(60*start_time_table[0]+start_time_table[1])).Format("20060102T150405")
				calendar += fmt.Sprintf("DTSTART;TZID=Asia/Shanghai:%v\n", start_time)
				end_time_table := timeTable[course.JSJC-1][1]
				end_time := firstDay.AddDate(0, 0, week*7+course.SKXQ-1).Add(time.Minute * time.Duration(60*end_time_table[0]+end_time_table[1])).Format("20060102T150405")
				calendar += fmt.Sprintf("DTEND;TZID=Asia/Shanghai:%v\n", end_time)
				calendar += fmt.Sprintf("UID:%v\n", uuid.New())
				calendar += "END:VEVENT\n"

				classCount++
				timeCount += (course.JSJC - course.KSJC + 1) * 45
			}
		}
	}

	calendar += "END:VCALENDAR\n"

	url, err = saver.Save(fmt.Sprintf("/tmp/%v.ics", uuid.New()), []byte(calendar))
	if err != nil {
		return "", "", errors.New("保存课程表失败Orz")
	}
	note = fmt.Sprintf("一共添加了%v学期的%v节课，合计坐牢时间%v小时（雾", schedule.Term, classCount, float64(timeCount)/60)
	return url, note, nil
}
