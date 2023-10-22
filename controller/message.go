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
	"database/sql"
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
	return nil
}

// 获取消息列表请求结构
type MessageGetListQuery struct {
	Type   string `form:"type" binding:"required"` // 消息对象
	LastID uint   `form:"last_id"`                 // 上次查询最后一条消息的ID 为0则不限制
}

// MessageGetListResponseItem 获取消息列表返回结构
type MessageGetListResponseItem struct {
	FromUser   UserAPI `json:"from_user"`
	ID         uint    `json:"id"`
	LinkObj    string  `json:"link_obj"`
	Obj        string  `json:"obj"`
	Text       string  `json:"text"`
	UpdateTime string  `json:"update_time"`
}

// MessageGetList 获取点赞消息列表
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
			UpdateTime: message.UpdatedAt.String(),
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
	res := MessageGetUnreadNumsResponse{}
	var summary database.MessageSummary
	if err := database.DB.Where("uid = ? AND type = ?", c.GetString("uid"), "comment").Limit(1).Find(&summary).Error; err != nil {
		c.JSON(500, gin.H{"msg": "获取未读评论数失败Orz"})
		return
	}
	res.Comment = int(summary.UnreadNum)
	if err := database.DB.Where("uid = ? AND type = ?", c.GetString("uid"), "follow").Limit(1).Find(&summary).Error; err != nil {
		c.JSON(500, gin.H{"msg": "获取未读关注数失败Orz"})
		return
	}
	res.Follow = int(summary.UnreadNum)
	if err := database.DB.Where("uid = ? AND type = ?", c.GetString("uid"), "like").Limit(1).Find(&summary).Error; err != nil {
		c.JSON(500, gin.H{"msg": "获取未读点赞数失败Orz"})
		return
	}
	res.Like = int(summary.UnreadNum)
	if err := database.DB.Where("uid = ? AND type = ?", c.GetString("uid"), "system").Limit(1).Find(&summary).Error; err != nil {
		c.JSON(500, gin.H{"msg": "获取未读系统消息数失败Orz"})
		return
	}
	res.System = int(summary.UnreadNum)
	c.JSON(200, res)
}

// 获取总未读消息数
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
