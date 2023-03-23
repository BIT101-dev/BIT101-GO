/*
 * @Author: flwfdd
 * @Date: 2023-03-23 16:07:43
 * @LastEditTime: 2023-03-23 16:39:29
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/nlp"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
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
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	var courses []database.Course
	q := database.DB.Model(&database.Course{})
	// 搜索
	if query.Search != "" {
		query := nlp.CutAll(query.Search)
		q = q.Scopes(database.SearchText(query))
	}
	// 排序
	if query.Order == "comment" {
		q = q.Order("comment_num DESC")
	} else if query.Order == "like" {
		q = q.Order("like_num DESC")
	} else if query.Order == "rate" {
		q = q.Order("rate DESC")
	} else { //默认new
		q = q.Order("updated_at DESC")
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
	database.DB.Save(&course)
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
	database.DB.Save(&course)
	return course.CommentNum, nil
}
