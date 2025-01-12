/*
 * @Author: flwfdd
 * @Date: 2023-03-30 08:55:28
 * @LastEditTime: 2023-05-17 16:51:49
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"BIT101-GO/util/push"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
)

// MessageSend 推送消息
func MessageSend(from_uid int, to_uid uint, obj string, tp string, link_obj string, content string) error {
	// 创建消息
	short_msg := []rune(content)
	if len(short_msg) > 233 {
		short_msg = short_msg[:233]
	}
	message := database.Message{
		Obj:     obj,
		FromUid: from_uid,
		ToUid:   to_uid,
		Type:    tp,
		LinkObj: link_obj,
		Content: string(short_msg),
	}
	if err := database.DB.Create(&message).Error; err != nil {
		return err
	}

	// 更新消息摘要
	var summary database.MessageSummary
	if err := database.DB.Where("uid = ? AND type = ?", to_uid, tp).Limit(1).Find(&summary).Error; err != nil {
		return err
	}
	if summary.ID == 0 {
		summary = database.MessageSummary{
			Uid:  to_uid,
			Type: tp,
		}
		if err := database.DB.Create(&summary).Error; err != nil {
			return err
		}
	}
	summary.UnreadNum++
	summary.LastTime = time.Now()
	summary.Content = content
	if err := database.DB.Save(&summary).Error; err != nil {
		return err
	}
	ser, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	PushMessageSend(to_uid, ser)
	return nil
}

// SystemMessagePostQuery 发送系统消息请求结构
type SystemMessagePostQuery struct {
	FromUid int    `json:"from_uid"`
	LinkObj string `json:"link_obj"`
	Obj     string `json:"obj"`
	Text    string `json:"text"`
	ToUid   uint   `json:"to_uid"`
}

// SystemMessagePost 发送系统消息
func SystemMessagePost(c *gin.Context) {
	var query SystemMessagePostQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误Orz"})
		return
	}
	err := MessageSend(query.FromUid, query.ToUid, query.Obj, "system", query.LinkObj, query.Text)
	if err != nil {
		c.JSON(500, gin.H{"msg": "发送失败Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "发送成功OvO"})
}

// MessageGetListQuery 获取消息列表请求结构
type MessageGetListQuery struct {
	Type   string `form:"type" binding:"required"` // 消息对象
	LastID uint   `form:"last_id"`                 // 上次查询最后一条消息的ID 为0则不限制
}

// MessageGetListResponseItem 获取消息列表返回结构
type MessageGetListResponseItem struct {
	FromUser   UserAPI   `json:"from_user"`
	ID         uint      `json:"id"`
	LinkObj    string    `json:"link_obj"`
	Obj        string    `json:"obj"`
	Text       string    `json:"text"`
	UpdateTime time.Time `json:"update_time"`
}

// MessageGetList 获取消息列表
func MessageGetList(c *gin.Context) {
	var query MessageGetListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误Orz"})
		return
	}

	var messages []database.Message
	q := database.DB.Where("to_uid = ? AND type = ?", c.GetString("uid"), query.Type)
	if query.LastID != 0 {
		q = q.Where("id < ?", query.LastID)
	}
	q.Order("id DESC").Limit(int(config.Config.MessagePageSize))
	if err := q.Find(&messages).Error; err != nil {
		c.JSON(500, gin.H{"msg": "获取消息列表失败Orz"})
		return
	}

	// 清空未读
	if query.LastID == 0 && len(messages) > 0 {
		var summary database.MessageSummary
		if err := database.DB.Where("uid = ? AND type = ?", c.GetString("uid"), query.Type).Limit(1).Find(&summary).Error; err != nil || summary.ID == 0 {
			c.JSON(500, gin.H{"msg": "清空未读失败Orz"})
			return
		}
		summary.UnreadNum = 0
		if err := database.DB.Save(&summary).Error; err != nil {
			c.JSON(500, gin.H{"msg": "清空未读失败Orz"})
			return
		}
	}

	// 获取用户信息
	uid_map := make(map[int]bool)
	for _, message := range messages {
		uid_map[message.FromUid] = true
	}
	user_map := GetUserAPIMap(uid_map)
	res := make([]MessageGetListResponseItem, 0)
	for _, message := range messages {
		res = append(res, MessageGetListResponseItem{
			FromUser:   user_map[message.FromUid],
			ID:         message.ID,
			LinkObj:    message.LinkObj,
			Obj:        message.Obj,
			Text:       message.Content,
			UpdateTime: message.UpdatedAt,
		})
	}
	c.JSON(200, res)
}

// MessageGetUnreadNumsResponse 获取分类未读消息返回结构
type MessageGetUnreadNumsResponse struct {
	Comment int `json:"comment"`
	Follow  int `json:"follow"`
	Like    int `json:"like"`
	System  int `json:"system"`
}

// MessageGetUnreadNums 获取未读消息数
func MessageGetUnreadNums(c *gin.Context) {
	res := MessageGetUnreadNumsResponse{0, 0, 0, 0}
	var summaries []database.MessageSummary
	if err := database.DB.Where("uid = ?", c.GetString("uid")).Find(&summaries).Error; err != nil {
		c.JSON(500, gin.H{"msg": "获取未读消息数失败Orz"})
		return
	}
	for _, summary := range summaries {
		switch summary.Type {
		case "comment":
			res.Comment = int(summary.UnreadNum)
		case "follow":
			res.Follow = int(summary.UnreadNum)
		case "like":
			res.Like = int(summary.UnreadNum)
		case "system":
			res.System = int(summary.UnreadNum)
		}
	}
	c.JSON(200, res)
}

// MessageGetUnreadNum 获取总未读消息数
func MessageGetUnreadNum(c *gin.Context) {
	var count sql.NullInt64
	if err := database.DB.Model(&database.MessageSummary{}).Select("SUM(unread_num)").Where("uid = ?", c.GetString("uid")).Pluck("sum", &count).Error; err != nil {
		c.JSON(500, gin.H{"msg": "获取未读消息数失败Orz"})
		return
	}
	if !count.Valid {
		c.JSON(200, gin.H{"unread_num": 0})
		return
	}
	c.JSON(200, gin.H{"unread_num": count.Int64})
}

// PushMessageRequestKey 获取推送服务请求公钥
// Web Push 1/4    获取服务端 VAPID 公钥
func PushMessageRequestKey(c *gin.Context) {
	public_key, err := push.GetRequestPubkey()
	if err != nil {
		c.JSON(500, gin.H{"msg": "推送服务错误Orz"})
		return
	}
	c.JSON(200, gin.H{"publicKey": public_key, "about": "https://bit101.cn/message"})
}

// PushMessageSubscribe 订阅推送服务
// Web Push 2/4    客户端生成订阅 服务端保存订阅端点与公钥以便推送
func PushMessageSubscribe(c *gin.Context) {
	var query webpush.Subscription
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误Orz"})
		return
	}
	uid := c.GetUint("uid_uint")
	if err := push.HandleRegister(query, uid); err != nil {
		c.JSON(500, gin.H{"msg": "订阅失败Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "订阅成功OvO"})
}

type PushMessage struct {
	Data      []byte `json:"data"`
	Badge     string `json:"badge"`
	Icon      string `json:"icon"`
	Timestamp int64  `json:"timestamp"`
}

// Web Push 3/4    服务端推送消息
func PushMessageSend(uid uint, content []byte) error {
	subscriptions := []database.WebPushSubscription{}
	if err := database.DB.Where("uid = ?", uid).Find(&subscriptions).Error; err != nil {
		return errors.New("数据库错误Orz")
	}
	if len(subscriptions) == 0 {
		return nil
	}

	msg := PushMessage{
		Data:      content,
		Badge:     "https://bit101.cn/favicon.ico",
		Icon:      "https://bit101.cn/pwa-512x512.png",
		Timestamp: time.Now().UTC().Unix(),
	}

	ctx, err := json.Marshal(msg)

	if err != nil {
		return errors.New("消息错误Orz")
	}

	for _, subscription := range subscriptions {
		if err := push.Send(subscription, ctx); err != nil {
			return errors.New("推送错误Orz")
		}
	}
	return nil
}

// PushMessageUnsubscribe 取消订阅推送服务
// Web Push 4/4    客户端取消订阅 服务端删除订阅端点
func PushMessageUnsubscribe(c *gin.Context) {
	var query webpush.Subscription
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误Orz"})
		return
	}
	uid := c.GetUint("uid_uint")
	if err := push.HandleUnregister(query, uid); err != nil {
		c.JSON(500, gin.H{"msg": "取消订阅失败Orz"})
		return
	}
	c.JSON(200, gin.H{"msg": "取消订阅成功OvO"})
}
