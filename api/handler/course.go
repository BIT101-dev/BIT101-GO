/*
 * @Author: flwfdd
 * @Date: 2023-03-23 16:07:43
 * @LastEditTime: 2025-03-11 15:37:49
 * @Description: _(:з」∠)_
 */
package handler

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/middleware"
	"BIT101-GO/api/types"
	"BIT101-GO/database"
	"BIT101-GO/pkg/webvpn"

	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	CourseSvc types.CourseService
}

func NewCourseHandler(courseSvc types.CourseService) *CourseHandler {
	return &CourseHandler{courseSvc}
}

// GetCoursesHandler 获取课程列表接口
func (h *CourseHandler) GetCoursesHandler(c *gin.Context) {
	type Request struct {
		Search string `form:"search"`
		Order  string `form:"order"` //like | comment | rate | new
		Page   uint   `form:"page"`
	}
	type Response []database.Course
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}

	courses, err := h.CourseSvc.GetCourses(query.Search, query.Order, query.Page)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(courses))
}

// GetCourseHandler 获取课程详情接口
func (h *CourseHandler) GetCourseHandler(c *gin.Context) {
	type Request struct {
		ID uint `uri:"id" binding:"required"`
	}
	type Response types.CourseAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindUri(&query), 400) {
		return
	}
	var uid uint
	userCtx, err := middleware.GetUserContext(c)
	if err == nil {
		uid = userCtx.UIDUint
	} else {
		uid = 0
	}

	course, err := h.CourseSvc.GetCourseAPI(query.ID, uid)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(course))
}

// GetUploadUrlHandler 获取课程资料上传链接
func (h *CourseHandler) GetUploadUrlHandler(c *gin.Context) {
	type Request struct {
		Number string `form:"number" binding:"required"` //课程编号
		Name   string `form:"name" binding:"required"`   //文件名
		Type   string `form:"type"`                      //资料类型 默认other
	}
	type Response struct {
		URL string `json:"url"`
		ID  uint   `json:"id"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint

	url, id, err := h.CourseSvc.GetUploadUrl(query.Number, query.Name, query.Type, uid)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{URL: url, ID: id})
}

// LogUploadHandler 上报课程资料上传记录
func (h *CourseHandler) LogUploadHandler(c *gin.Context) {
	type Request struct {
		ID  uint   `json:"id" binding:"required"` //上传记录ID
		Msg string `json:"msg"`                   //上传备注
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint

	err := h.CourseSvc.LogUpload(query.ID, query.Msg, uid)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, gin.H{"msg": "上传成功OvO"})
}

// GetCourseHistoryHandler 获取课程历史接口
func (h *CourseHandler) GetCourseHistoryHandler(c *gin.Context) {
	type Request struct {
		Number string `uri:"number" binding:"required"`
	}
	type Response []types.CourseHistoryAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindUri(&query), 400) {
		return
	}

	history, err := h.CourseSvc.GetCourseHistory(query.Number)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(history))
}

// CourseScheduleHandler 获取课程表
func (h *CourseHandler) CourseScheduleHandler(c *gin.Context) {
	type Request struct {
		Term string `form:"term"` // 学期
	}
	type Response struct {
		Url  string `json:"url"`
		Note string `json:"note"`
		Msg  string `json:"msg"`
	}
	var query Request
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误awa"})
		return
	}
	cookie := c.Request.Header.Get("webvpn-cookie")
	if cookie == "" {
		common.HandleErrorWithCode(c, webvpn.ErrCookieInvalid, 400)
		return
	}

	url, note, err := h.CourseSvc.GetCourseSchedule(cookie, query.Term)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{url, note, "获取成功OvO"})
}
