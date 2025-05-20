/*
 * @Author: flwfdd
 * @Date: 2023-03-30 08:55:28
 * @LastEditTime: 2025-03-11 01:51:24
 * @Description: _(:з」∠)_
 */
package handler

import (
	"BIT101-GO/api/common"
	"BIT101-GO/api/middleware"
	"BIT101-GO/api/service"
	"BIT101-GO/api/types"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
)

// MessageHandler 消息模块接口
type MessageHandler struct {
	MessageSvc *service.MessageService
}

// NewMessageHandler 创建消息模块接口
func NewMessageHandler(s *service.MessageService) MessageHandler {
	return MessageHandler{s}
}

// SendSystem 发送系统消息接口
func (h *MessageHandler) SendSystemHandler(c *gin.Context) {
	type Request struct {
		FromUid int    `json:"from_uid"`
		LinkObj string `json:"link_obj"`
		Obj     string `json:"obj"`
		Text    string `json:"text"`
		ToUid   uint   `json:"to_uid"`
	}
	type Response struct {
		Msg string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindJSON(&query), 400) {
		return
	}

	if common.HandleError(c, h.MessageSvc.Send(query.FromUid, query.ToUid, query.Obj, types.MessageTypeSystem, query.LinkObj, query.Text)) {
		return
	}

	c.JSON(200, Response{"发送成功OvO"})
}

// MessageGetList 获取消息列表
func (h *MessageHandler) GetListHandler(c *gin.Context) {
	type Request struct {
		Type   string `form:"type" binding:"required"` // 消息对象
		LastID uint   `form:"last_id"`                 // 上次查询最后一条消息的ID 为0则不限制
	}
	type GetListResponse []types.MessageAPI
	var query Request
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误Orz"})
		return
	}

	uid := middleware.MustGetUserContext(c).UIDUint
	messages, err := h.MessageSvc.GetList(uid, types.MessageType(query.Type), query.LastID)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, GetListResponse(messages))
}

// GetUnreadNum 获取总未读消息数接口
func (h *MessageHandler) GetUnreadNumHandler(c *gin.Context) {
	type Response struct {
		UnreadNum uint `json:"unread_num"`
	}
	unreadNum, err := h.MessageSvc.GetUnreadNum(middleware.MustGetUserContext(c).UIDUint)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response{unreadNum})
}

// GetUnreadNumsHandler 获取未读消息数
func (h *MessageHandler) GetUnreadNumsHandler(c *gin.Context) {
	type Response map[string]uint

	unreadNums, err := h.MessageSvc.GetUnreadNums(middleware.MustGetUserContext(c).UIDUint)
	if common.HandleError(c, err) {
		return
	}

	c.JSON(200, Response(unreadNums))
}

// WebpushRequestKeyHandler 获取推送服务请求公钥接口
// Web Push 1/4    获取服务端 VAPID 公钥
func (h *MessageHandler) WebpushRequestKeyHandler(c *gin.Context) {
	type Response struct {
		PublicKey string `json:"publicKey"`
		About     string `json:"about"`
	}
	public_key := h.MessageSvc.WebpushGetRequestPubkey()
	c.JSON(200, Response{public_key, "https://bit101.cn/message"})
}

// WebpushSubscribeHandler 订阅推送服务接口
// Web Push 2/4    客户端生成订阅 服务端保存订阅端点与公钥以便推送
func (h *MessageHandler) WebpushSubscribeHandler(c *gin.Context) {
	type Request webpush.Subscription
	type Response struct {
		Msg string `json:"msg"`
	}

	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindJSON(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint
	if common.HandleError(c, h.MessageSvc.WebpushSubscribe(webpush.Subscription(query), uid)) {
		return
	}
	c.JSON(200, Response{"订阅成功OvO"})
}

// WebpushUnsubscribeHandler 取消订阅推送服务
// Web Push 4/4    客户端取消订阅 服务端删除订阅端点
func (h *MessageHandler) WebpushUnsubscribeHandler(c *gin.Context) {
	type Request webpush.Subscription
	type Response struct {
		Msg string `json:"msg"`
	}
	var query Request
	if common.HandleErrorWithCode(c, c.ShouldBindJSON(&query), 400) {
		return
	}
	uid := middleware.MustGetUserContext(c).UIDUint
	if common.HandleErrorWithMessage(c, h.MessageSvc.WebpushCancelRegister(webpush.Subscription(query), uid), "取消订阅失败Orz") {
		return
	}
	c.JSON(200, Response{"取消订阅成功OvO"})
}
