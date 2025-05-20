/**
* @author:YHCnb
* Package:
* @date:2023/10/28 21:27
* Description:
 */
package handler

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/middleware"
	"BIT101-GO/api/service"
	"BIT101-GO/api/types"
	"BIT101-GO/database"

	"github.com/gin-gonic/gin"
)

// ManageHandler 管理模块接口
type ManageHandler struct {
	ManageSvc *service.ManageService
}

// NewManageHandler 创建管理模块接口
func NewManageHandler(s *service.ManageService) ManageHandler {
	return ManageHandler{s}
}

// ReportTypeListGet 获取举报类型列表
func (h *ManageHandler) GetReportTypesHandler(c *gin.Context) {
	type Response []database.ReportType

	reportTypes := make([]database.ReportType, 0)

	c.JSON(200, Response(reportTypes))
}

// Report 举报
func (h *ManageHandler) ReportHandler(c *gin.Context) {
	type Request struct {
		Obj    string `json:"obj"`
		TypeId uint   `json:"type_id"`
		Text   string `json:"text"`
	}
	type Response struct {
		Msg string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint

	if common.HandleError(c, h.ManageSvc.Report(uid, query.Obj, query.TypeId, query.Text)) {
		return
	}

	c.JSON(200, Response{"举报成功OvO"})
}

// UpdateReportStatusHandler 修改举报状态
func (h *ManageHandler) UpdateReportStatusHandler(c *gin.Context) {
	type Request struct {
		ID     int `form:"id"`
		Status int `form:"status"`
	}
	type Response struct {
		Msg string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}

	if common.HandleError(c, h.ManageSvc.UpdateReportStatus(uint(query.ID), query.Status)) {
		return
	}

	c.JSON(200, Response{"修改成功OvO"})
}

// GetReportsHandler 获取举报列表
func (h *ManageHandler) GetReportsHandler(c *gin.Context) {
	type Request struct {
		Page   uint   `form:"page"`
		Uid    int    `form:"uid"`
		Obj    string `form:"obj"`
		Status int    `form:"status"` // -1为全部，0为未处理 1为举报成功 2为举报失败
	}
	type Response []types.ReportAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}

	reports, err := h.ManageSvc.GetReports(query.Uid, query.Obj, query.Status, query.Page)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(reports))
}

// BanHandler 关小黑屋接口
func (h *ManageHandler) BanHandler(c *gin.Context) {
	type Request struct {
		Time string `json:"time"` // 解封时间，格式为ISO 8601带时区
		Uid  uint   `json:"uid"`
	}
	type Response struct {
		Msg string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}
	t, err := common.ParseTime(query.Time)
	if common.HandleErrorWithCode(c, err, 400) {
		return
	}

	if common.HandleError(c, h.ManageSvc.Ban(query.Uid, t)) {
		return
	}

	c.JSON(200, Response{"关小黑屋成功OvO"})
}

// GetBansHandler 获取小黑屋列表
func (h *ManageHandler) GetBansHandler(c *gin.Context) {
	type Request struct {
		Page uint `form:"page"`
		Uid  int  `form:"uid"`
	}
	type Response []types.BanAPI
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBind(&query), 400) {
		return
	}

	bans, err := h.ManageSvc.GetBans(query.Uid, query.Page)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(bans))
}
