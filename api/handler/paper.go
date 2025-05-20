/*
 * @Author: flwfdd
 * @Date: 2023-03-21 17:34:55
 * @LastEditTime: 2025-03-18 18:54:22
 * @Description: _(:з」∠)_
 */
package handler

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/middleware"
	"BIT101-GO/api/service"
	"BIT101-GO/api/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaperHandler struct {
	paperSvc *service.PaperService
}

func NewPaperHandler(paperSvc *service.PaperService) *PaperHandler {
	return &PaperHandler{paperSvc: paperSvc}
}

// GetHandler 获取文章
func (h *PaperHandler) GetHandler(c *gin.Context) {
	type Request struct {
		ID uint `uri:"id" binding:"required"`
	}
	type Response types.PaperInfo
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindUri(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint
	admin := middleware.MustGetUserContext(c).IsAdmin

	paperInfo, err := h.paperSvc.Get(query.ID, uid, admin)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(paperInfo))
}

// CreateHandler 新建文章
func (h *PaperHandler) CreateHandler(c *gin.Context) {
	type Request struct {
		Title      string `json:"title" binding:"required"`
		Intro      string `json:"intro" binding:"required"`
		Content    string `json:"content" binding:"required"`
		Anonymous  bool   `json:"anonymous"`
		PublicEdit bool   `json:"public_edit"`
	}
	type Response struct {
		ID  uint   `json:"id"`
		Msg string `json:"msg"`
	}
	var query Request
	query.PublicEdit = true
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint

	id, err := h.paperSvc.Create(query.Title, query.Intro, query.Content, query.Anonymous, query.PublicEdit, uid)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{ID: id, Msg: "创建成功OvO"})
}

// EditHandler 编辑文章
func (h *PaperHandler) EditHandler(c *gin.Context) {
	type Request struct {
		ID         uint    `uri:"id"` // 需要手动绑定
		Title      string  `json:"title" binding:"required"`
		Intro      string  `json:"intro" binding:"required"`
		Content    string  `json:"content" binding:"required"`
		Anonymous  bool    `json:"anonymous"`
		PublicEdit bool    `json:"public_edit"`
		LastTime   float64 `json:"last_time" binding:"required"`
	}
	type Response struct {
		Msg string `json:"msg"`
	}
	var query Request
	query.PublicEdit = true
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if common.HandleErrorWithCode(c, err, 400) {
		return
	}
	query.ID = uint(id)

	uid := middleware.MustGetUserContext(c).UIDUint
	admin := middleware.MustGetUserContext(c).IsAdmin

	if common.HandleError(c, h.paperSvc.Edit(query.ID, query.Title, query.Intro, query.Content, query.Anonymous, query.PublicEdit, query.LastTime, uid, admin)) {
		return
	}

	c.JSON(200, Response{Msg: "编辑成功OvO"})
}

// GetListHandler 获取文章列表
func (h *PaperHandler) GetListHandler(c *gin.Context) {
	type Request struct {
		Search string `form:"search"`
		Order  string `form:"order"` // new | like | comment
		Page   uint   `form:"page"`
	}
	// 获取文章列表返回结构
	type Response []types.PaperAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}

	paperList, err := h.paperSvc.GetList(query.Search, query.Order, query.Page)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(paperList))
}

// DeleteHandler 删除文章
func (h *PaperHandler) DeleteHandler(c *gin.Context) {
	type Request struct {
		ID uint `uri:"id" binding:"required"`
	}
	type Response struct {
		Msg string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindUri(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint
	admin := middleware.MustGetUserContext(c).IsAdmin

	err := h.paperSvc.Delete(query.ID, uid, admin)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{Msg: "删除成功OvO"})
}
