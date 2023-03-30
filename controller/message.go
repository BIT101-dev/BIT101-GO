/*
 * @Author: flwfdd
 * @Date: 2023-03-30 08:55:28
 * @LastEditTime: 2023-03-30 16:29:26
 * @Description: _(:з」∠)_
 */
package controller

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"
	"time"

	"github.com/gin-gonic/gin"
)

// 推送消息
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
	if err := database.DB.Where("obj = ? AND uid = ?", obj, to_uid).Limit(1).Find(&summary).Error; err != nil {
		return err
	}
	if summary.ID == 0 {
		summary = database.MessageSummary{
			Uid: to_uid,
			Obj: obj,
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
	return nil
}

// 获取未读点赞数
func MessageGetUnreadLikeNum(c *gin.Context) {
	var summary database.MessageSummary
	if err := database.DB.Where("uid = ? AND obj = ?", c.GetString("uid"), "like").Limit(1).Find(&summary).Error; err != nil {
		c.JSON(500, gin.H{"msg": "获取未读点赞数失败Orz"})
		return
	}
	if summary.ID == 0 {
		c.JSON(200, gin.H{"unread_num": 0})
		return
	}
	c.JSON(200, gin.H{"unread_num": summary.UnreadNum})
}

// 获取消息列表请求结构
type MessageGetListQuery struct {
	Obj    string `form:"obj" binding:"required"` // 消息对象
	LastID uint   `form:"last_id" default:"0"`    // 上次查询最后一条消息的ID 为0则不限制
}

type MessageGetListResponseItem struct {
	database.Message
	FromUser UserAPI `json:"from_user"`
}

// 获取点赞消息列表
func MessageGetList(c *gin.Context) {
	var query MessageGetListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误Orz"})
		return
	}

	var messages []database.Message
	q := database.DB.Where("to_uid = ? AND obj = ?", c.GetString("uid"), query.Obj)
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
		if err := database.DB.Where("uid = ? AND obj = ?", c.GetString("uid"), query.Obj).Limit(1).Find(&summary).Error; err != nil || summary.ID == 0 {
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
			Message:  message,
			FromUser: user_map[message.FromUid],
		})
	}

	c.JSON(200, res)
}

// 获取未读评论数
func MessageGetUnreadCommentNum(c *gin.Context) {
	var summary database.MessageSummary
	if err := database.DB.Where("uid = ? AND obj = ?", c.GetString("uid"), "comment").Limit(1).Find(&summary).Error; err != nil {
		c.JSON(500, gin.H{"msg": "获取未读评论数失败Orz"})
		return
	}
	if summary.ID == 0 {
		c.JSON(200, gin.H{"unread_num": 0})
		return
	}
	c.JSON(200, gin.H{"unread_num": summary.UnreadNum})
}

// 获取总未读消息数
func MessageGetUnreadNum(c *gin.Context) {
	var count int64
	if err := database.DB.Model(&database.MessageSummary{}).Select("SUM(unread_num)").Where("uid = ?", c.GetString("uid")).Pluck("sum", &count).Error; err != nil {
		c.JSON(500, gin.H{"msg": "获取未读消息数失败Orz"})
		return
	}
	c.JSON(200, gin.H{"unread_num": count})
}
