/**
* @author:YHCnb
* Package:
* @date:2023/10/16 21:52
* Description:
 */
package handler

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/middleware"
	"BIT101-GO/api/types"
	"BIT101-GO/database"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PosterHandler struct {
	posterSvc types.PosterService
}

func NewPosterHandler(posterSvc types.PosterService) *PosterHandler {
	return &PosterHandler{posterSvc: posterSvc}
}

// GetHandler 获取帖子接口
func (h *PosterHandler) GetHandler(c *gin.Context) {
	type Request struct {
		ID uint `uri:"id" binding:"required"`
	}
	type Response types.PosterInfo
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindUri(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint
	admin := middleware.MustGetUserContext(c).IsAdmin

	posterInfo, err := h.posterSvc.Get(query.ID, uid, admin)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(posterInfo))
}

// CreateHandler 发布帖子
func (h *PosterHandler) CreateHandler(c *gin.Context) {
	type Request struct {
		Title     string   `json:"title" binding:"required"`
		Text      string   `json:"text"`
		ImageMids []string `json:"image_mids"`
		Plugins   string   `json:"plugins"`
		Anonymous bool     `json:"anonymous"`
		Tags      []string `json:"tags"`
		ClaimID   uint     `json:"claim_id"`
		Public    bool     `json:"public"`
	}
	type Response struct {
		ID  uint   `json:"id"`
		Msg string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint
	admin := middleware.MustGetUserContext(c).IsAdmin

	id, err := h.posterSvc.Create(query.Title, query.Text, query.ImageMids, query.Plugins, query.Anonymous, query.Tags, query.ClaimID, query.Public, uid, admin)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{ID: id, Msg: "发布成功OvO"})
}

// EditHandler 修改帖子
func (h *PosterHandler) EditHandler(c *gin.Context) {
	type Request struct {
		ID        uint     `uri:"id"`
		Title     string   `json:"title" binding:"required"`
		Text      string   `json:"text"`
		ImageMids []string `json:"image_mids"`
		Plugins   string   `json:"plugins"`
		Anonymous bool     `json:"anonymous"`
		Tags      []string `json:"tags"`
		ClaimID   uint     `json:"claim_id"`
		Public    bool     `json:"public"`
	}
	type Response struct {
		Msg string `json:"msg"`
	}
	var query Request
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

	if common.HandleError(c, h.posterSvc.Edit(query.ID, query.Title, query.Text, query.ImageMids, query.Plugins, query.Anonymous, query.Tags, query.ClaimID, query.Public, uid, admin)) {
		return
	}

	c.JSON(200, Response{Msg: "编辑成功OvO"})

}

// GetListHandler 获取帖子列表
func (h *PosterHandler) GetListHandler(c *gin.Context) {
	type Request struct {
		Mode   string `form:"mode"` //recommend | search | follow | hot 默认为recommend
		Page   uint   `form:"page"`
		Search string `form:"search"`
		Order  string `form:"order"` //like | new 默认为new
		Uid    int    `form:"uid"`   // 0为个人主页（显示匿名和未公开的帖子）
	}
	type Response []types.PosterAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint

	var noAnonymous, onlyPublic bool
	if query.Mode == "follow" {
		query.Uid = int(uid)
	}
	if query.Mode == "search" {
		if query.Uid < 0 {
			// 当uid<0返回public=true，anonymous任意的帖子（搜索）
			query.Uid = 0
			onlyPublic = true
		} else if query.Uid == 0 {
			// 当uid=0时，全部返回（自己主页）
			query.Uid = int(uid)
		} else {
			// 当uid>0时，仅返回public=true且anonymous=false的帖子（他人主页）
			noAnonymous = true
			onlyPublic = true
		}
	}

	posterAPIs, err := h.posterSvc.GetList(query.Mode, query.Page, query.Search, query.Order, uint(query.Uid), noAnonymous, onlyPublic)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(posterAPIs))
}

// DeleteHandler 删除帖子
func (h *PosterHandler) DeleteHandler(c *gin.Context) {
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

	err := h.posterSvc.Delete(query.ID, uid, admin)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{Msg: "删除成功OvO"})
}

// GetClaimsHandler 获取声明列表
func (h *PosterHandler) GetClaimsHandler(c *gin.Context) {
	type Response []database.Claim

	claims := h.posterSvc.GetClaims()

	c.JSON(200, Response(claims))
}
