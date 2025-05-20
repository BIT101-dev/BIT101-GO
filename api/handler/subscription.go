/*
 * @Author: flwfdd
 * @Date: 2023-06-15 14:51:53
 * @LastEditTime: 2025-03-19 01:34:06
 * @Description: 订阅处理器
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

// SubscriptionHandler 订阅处理器
type SubscriptionHandler struct {
	SubscriptionSvc *service.SubscriptionService
}

// NewSubscriptionHandler 创建订阅处理器
func NewSubscriptionHandler(subscriptionSvc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		SubscriptionSvc: subscriptionSvc,
	}
}

// SubscribeHandler 添加订阅接口
func (h *SubscriptionHandler) SubscribeHandler(c *gin.Context) {
	type Request struct {
		Obj   string `json:"obj" binding:"required"` // 订阅对象
		Level int    `json:"level"`                  // 订阅级别
	}
	type Response struct {
		Msg string `json:"msg"`
	}
	var req Request
	if common.HandleErrorWithCode(c, c.ShouldBindJSON(&req), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint

	err := h.SubscriptionSvc.Subscribe(uid, req.Obj, req.Level)
	if common.HandleError(c, err) {
		return
	}

	msg := ""
	switch database.SubscriptionLevel(req.Level) {
	case database.SubscriptionLevelNone:
		msg = "取消订阅成功OvO"
	case database.SubscriptionLevelSilent:
		msg = "悄悄订阅成功OvO"
	case database.SubscriptionLevelUpdate:
		msg = "订阅成功 将在有更新时收到通知OvO"
	case database.SubscriptionLevelComment:
		msg = "订阅成功 将在有更新或评论时收到通知OvO"
	}
	c.JSON(200, Response{msg})
}

// GetSubscriptionsHandler 获取订阅列表接口
func (h *SubscriptionHandler) GetSubscriptionsHandler(c *gin.Context) {
	type Request struct {
		Page uint `form:"page"`
	}
	type Response []types.SubscriptionAPI
	var req Request
	if common.HandleErrorWithCode(c, c.ShouldBindQuery(&req), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint

	subscriptions, err := h.SubscriptionSvc.GetSubscriptions(uid, req.Page)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(subscriptions))
}
