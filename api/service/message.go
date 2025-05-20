/*
 * @Author: flwfdd
 * @Date: 2025-03-07 14:51:53
 * @LastEditTime: 2025-03-19 00:48:02
 * @Description: _(:з」∠)_
 */
package service

import (
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MessageService 消息模块服务
type MessageService struct {
	UserSvc *UserService
}

// NewMessageService 创建消息模块服务
func NewMessageService(userSvc *UserService) *MessageService {
	return &MessageService{userSvc}
}

// SetUserService 设置用户服务
func (s *MessageService) SetUserService(userSvc *UserService) {
	s.UserSvc = userSvc
}

// Send 发送消息
// obj 触发消息的对象，如 "poster1" "comment2"
// typ 为消息类型，如 "like" "comment" "follow"
// link_obj 为消息链接对象，如 "poster1" "paper2"
func (s *MessageService) Send(from_uid int, to_uid uint, objID string, typ types.MessageType, linkObjID string, content string) error {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 创建消息
		short_msg := []rune(content)
		if len(short_msg) > 233 {
			short_msg = short_msg[:233]
		}
		message := database.Message{
			Obj:     objID,
			FromUid: from_uid,
			ToUid:   to_uid,
			Type:    string(typ),
			LinkObj: linkObjID,
			Content: string(short_msg),
		}
		if err := tx.Create(&message).Error; err != nil {
			return err
		}

		// 更新消息摘要
		var summary database.MessageSummary
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("uid = ? AND type = ?", to_uid, typ).Limit(1).Find(&summary).Error; err != nil {
			return err
		}
		if summary.ID == 0 {
			summary = database.MessageSummary{
				Uid:  to_uid,
				Type: string(typ),
			}
			if err := tx.Create(&summary).Error; err != nil {
				return err
			}
		}
		summary.UnreadNum++
		summary.LastTime = time.Now()
		summary.Content = content
		if err := tx.Save(&summary).Error; err != nil {
			return err
		}

		// Webpush
		ser, err := json.Marshal(summary)
		if err != nil {
			return err
		}
		s.WebpushSend(to_uid, ser)
		return nil
	})
	return err
}

// GetList 获取消息列表
// typ 为消息类型，如 "like" "comment" "follow"
// lastID 为上次查询最后一条消息的ID 为0则查询最新的
func (s *MessageService) GetList(uid uint, typ types.MessageType, lastID uint) ([]types.MessageAPI, error) {
	var messages []database.Message
	q := database.DB.Where("to_uid = ? AND type = ?", uid, typ)
	if lastID != 0 {
		q = q.Where("id < ?", lastID)
	}
	q.Order("id DESC").Limit(int(config.Get().MessagePageSize))
	if err := q.Find(&messages).Error; err != nil {
		return nil, errors.New("获取消息列表失败Orz")
	}

	// 清空未读
	if lastID == 0 && len(messages) > 0 {
		var summary database.MessageSummary
		if err := database.DB.Where("uid = ? AND type = ?", uid, typ).Limit(1).Find(&summary).Error; err != nil || summary.ID == 0 {
			return nil, errors.New("清空未读失败Orz")
		}
		summary.UnreadNum = 0
		if err := database.DB.Save(&summary).Error; err != nil {
			return nil, errors.New("清空未读失败Orz")
		}
	}

	// 获取用户信息
	res, err := FillUsers(
		s.UserSvc,
		messages,
		func(message database.Message) int {
			return message.FromUid
		},
		func(message database.Message, user types.UserAPI) types.MessageAPI {
			return types.MessageAPI{
				FromUser:   user,
				ID:         message.ID,
				LinkObj:    message.LinkObj,
				Obj:        message.Obj,
				Text:       message.Content,
				UpdateTime: message.UpdatedAt,
			}
		})
	if err != nil {
		return nil, errors.New("获取用户信息失败Orz")
	}
	return res, nil
}

// GetUnreadNums 获取各类消息未读数
func (s *MessageService) GetUnreadNums(uid uint) (map[string]uint, error) {
	var summaries []database.MessageSummary
	if err := database.DB.Where("uid = ?", uid).Find(&summaries).Error; err != nil {
		return nil, errors.New("获取未读消息数失败Orz")
	}
	var unreadNums = make(map[string]uint)
	for typ, _ := range types.MessageTypeMap {
		unreadNums[string(typ)] = 0
	}
	for _, summary := range summaries {
		unreadNums[summary.Type] = summary.UnreadNum
	}
	return unreadNums, nil
}

// GetUnreadNum 获取总未读消息数
func (s *MessageService) GetUnreadNum(uid uint) (uint, error) {
	var count sql.NullInt64
	if err := database.DB.Model(&database.MessageSummary{}).Select("SUM(unread_num)").Where("uid = ?", uid).Pluck("sum", &count).Error; err != nil {
		return 0, errors.New("获取未读消息数失败Orz")
	}
	if !count.Valid {
		return 0, nil
	}
	return uint(count.Int64), nil
}

// WebpushSubscribe 订阅Webpush
func (s *MessageService) WebpushSubscribe(sub webpush.Subscription, uid uint) error {
	subscription := database.WebPushSubscription{
		Uid:            uid,
		Endpoint:       sub.Endpoint,
		ExpirationTime: "null",
		Auth:           sub.Keys.Auth,
		P256dh:         sub.Keys.P256dh,
	}

	if err := database.DB.Create(&subscription).Error; err != nil {
		return errors.New("数据库错误Orz")
	}

	return nil
}

// WebpushSend 推送Webpush消息
// Web Push 3/4    服务端推送消息
func (s *MessageService) WebpushSend(uid uint, content []byte) error {
	subscriptions := []database.WebPushSubscription{}
	if err := database.DB.Where("uid = ?", uid).Find(&subscriptions).Error; err != nil {
		return errors.New("数据库错误Orz")
	}
	if len(subscriptions) == 0 {
		return nil
	}

	msg := types.WebpushMessage{
		Data:      content,
		Badge:     "https://bit101.cn/favicon.ico",
		Icon:      "https://bit101.cn/pwa-512x512.png",
		Timestamp: time.Now().UTC().Unix(),
	}

	ctx, err := json.Marshal(msg)
	if err != nil {
		return errors.New("消息错误Orz")
	}

	slog.Info("[Webpush] Send notification", "endpoint_count", len(subscriptions), "uid", uid)
	for _, subscription := range subscriptions {
		if err := s.WebpushSendToSubscription(subscription, ctx); err != nil {
			slog.Error("[Webpush] Send notification", "uid", subscription.Uid, "error", err)
			continue
		}
	}
	return nil
}

// WebpushGetRequestPubkey 获取Webpush请求公钥
func (s *MessageService) WebpushGetRequestPubkey() string {
	return config.Get().WebPushKeys.Public
}

// WebpushSendToSubscription 发送Webpush消息到订阅
func (s *MessageService) WebpushSendToSubscription(sub database.WebPushSubscription, message []byte) error {
	if sub.Endpoint == "" || sub.Auth == "" || sub.P256dh == "" {
		return errors.New("订阅错误Orz")
	}

	subscription := webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			Auth:   sub.Auth,
			P256dh: sub.P256dh,
		},
	}

	resp, err := webpush.SendNotification(message, &subscription, &webpush.Options{
		VAPIDPublicKey:  config.Get().WebPushKeys.Public,
		VAPIDPrivateKey: config.Get().WebPushKeys.Private,
		TTL:             30,
	})
	if err != nil {
		slog.Error("[Push] Send notification failed", "uid", sub.Uid, "error", err, "endpoint", sub.Endpoint[:30], "auth", sub.Auth, "p256dh", sub.P256dh)
		return errors.New("推送错误Orz")
	}
	defer resp.Body.Close()
	slog.Info("[Push] Send notification", "uid", sub.Uid, "endpoint", sub.Endpoint[:30], "status", resp.Status)
	return nil
}

// WebpushCancelRegister 取消Webpush订阅
func (s *MessageService) WebpushCancelRegister(sub webpush.Subscription, uid uint) error {
	subscription := database.WebPushSubscription{
		Uid:            uid,
		Endpoint:       sub.Endpoint,
		ExpirationTime: "null",
		Auth:           sub.Keys.Auth,
		P256dh:         sub.Keys.P256dh,
	}

	var target database.WebPushSubscription
	database.DB.Where("uid = ?", uid).Where(
		"endpoint = ?", subscription.Endpoint).Where("auth = ?", subscription.Auth).Where(
		"p256dh = ?", subscription.P256dh).Limit(1).Find(&target)

	if target.ID == 0 {
		return errors.New("订阅不存在Orz")
	}

	if err := database.DB.Delete(&target).Error; err != nil {
		return errors.New("数据库错误Orz")
	}

	return nil
}
