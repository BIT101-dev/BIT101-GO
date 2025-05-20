/*
 * @Author: flwfdd
 * @Date: 2023-03-21 23:16:18
 * @LastEditTime: 2025-03-11 15:56:41
 * @Description: _(:з」∠)_
 */
package handler

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/middleware"
	"BIT101-GO/api/service"
	"BIT101-GO/api/types"
	"errors"

	"github.com/gin-gonic/gin"
)

// ReactionHandler 交互模块响应
type ReactionHandler struct {
	ReactionSvc *service.ReactionService
}

// NewReactionHandler 创建交互模块响应
func NewReactionHandler(reactionSvc *service.ReactionService) *ReactionHandler {
	return &ReactionHandler{ReactionSvc: reactionSvc}
}

// LikeHandler 点赞接口
func (h *ReactionHandler) LikeHandler(c *gin.Context) {
	type Request struct {
		Obj string `json:"obj" binding:"required"` // 操作对象
	}
	type Response types.LikeAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	obj, err := types.NewObj(query.Obj)
	if common.HandleError(c, err) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint

	likeAPI, err := h.ReactionSvc.Like(obj, uid)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(likeAPI))
}

// CommentHandler 评论接口
func (h *ReactionHandler) CommentHandler(c *gin.Context) {
	type Request struct {
		Obj       string   `json:"obj" binding:"required"`   // 操作对象
		Text      string   `json:"text" binding:"max=23333"` // 评论内容
		Anonymous bool     `json:"anonymous"`                // 是否匿名
		ReplyUid  int      `json:"reply_uid"`                //回复用户id
		ReplyObj  string   `json:"reply_obj"`                //回复对象
		Rate      uint     `json:"rate"`                     //评分
		ImageMids []string `json:"image_mids"`               //图片
	}
	type Response types.CommentAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	if query.Text == "" && len(query.ImageMids) == 0 {
		common.HandleError(c, errors.New("内容不能为空Orz"))
		return
	}
	obj, err := types.NewObj(query.Obj)
	if common.HandleError(c, err) {
		return
	}
	var replyObj types.Obj
	if query.ReplyObj != "" {
		replyObj, err = types.NewObj(query.ReplyObj)
		if common.HandleError(c, err) {
			return
		}
	}
	uid := middleware.MustGetUserContext(c).UIDUint
	admin := middleware.MustGetUserContext(c).IsAdmin

	commentAPI, err := h.ReactionSvc.Comment(obj, query.Text, query.Anonymous, query.ReplyUid, replyObj, query.Rate, query.ImageMids, uid, admin)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(commentAPI))
}

// GetCommentsHandler 获取评论列表
func (h *ReactionHandler) GetCommentsHandler(c *gin.Context) {
	type Request struct {
		Obj   string `form:"obj" binding:"required"` // 操作对象
		Order string `form:"order"`                  // 排序方式
		Page  uint   `form:"page"`                   // 页码
	}
	type Response []types.CommentAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint
	admin := middleware.MustGetUserContext(c).IsAdmin
	obj, err := types.NewObj(query.Obj)
	if common.HandleError(c, err) {
		return
	}

	commentAPIs, err := h.ReactionSvc.GetComments(obj, query.Order, query.Page, uid, admin)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(commentAPIs))
}

// 删除评论
func (h *ReactionHandler) DeleteCommentHandler(c *gin.Context) {
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

	err := h.ReactionSvc.DeleteComment(query.ID, uid, admin)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{"删除成功OvO"})
}

// StayHandler 停留上报接口
func (h *ReactionHandler) StayHandler(c *gin.Context) {
	type Request struct {
		Obj  string `json:"obj" binding:"required"`  // 操作对象
		Time int    `json:"time" binding:"required"` // 停留时间
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	obj, err := types.NewObj(query.Obj)
	if common.HandleError(c, err) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint

	if err := h.ReactionSvc.Stay(obj, uid); err != nil {
		common.HandleError(c, err)
		return
	}

	c.JSON(200, gin.H{})
}
