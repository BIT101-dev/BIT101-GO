/*
 * @Author: flwfdd
 * @Date: 2023-06-15 14:51:53
 * @LastEditTime: 2025-03-19 00:01:02
 * @Description: 订阅服务
 */
package service

import (
	"BIT101-GO/api/types"
	"BIT101-GO/config"
	"BIT101-GO/database"
	"errors"
	"log/slog"
)

// 检查是否实现了SubscriptionService接口
var _ types.SubscriptionService = (*SubscriptionService)(nil)

// SubscriptionService 订阅服务
type SubscriptionService struct {
	messageSvc types.MessageService
}

// NewSubscriptionService 创建订阅服务
func NewSubscriptionService(messageSvc types.MessageService) *SubscriptionService {
	return &SubscriptionService{messageSvc}
}

// GetSubscriptionLevel 获取订阅级别
func (s *SubscriptionService) GetSubscriptionLevel(uid uint, objID string) (database.SubscriptionLevel, error) {
	var subscription database.Subscription
	if err := database.DB.Where("uid = ? AND obj = ?", uid, objID).Limit(1).Find(&subscription).Error; err != nil {
		return 0, errors.New("数据库错误Orz")
	}
	if subscription.ID == 0 {
		return subscription.Level, errors.New("订阅不存在Orz")
	}

	return subscription.Level, nil
}

// NotifySubscription 通知订阅者
func (s *SubscriptionService) NotifySubscription(objID string, level database.SubscriptionLevel, text string) error {
	var subscriptions []database.Subscription
	if err := database.DB.Where("obj = ? AND level >= ?", objID, level).Find(&subscriptions).Error; err != nil {
		slog.Error("subscription: get subscription list failed", "objID", objID, "error", err.Error())
		return errors.New("获取订阅列表失败Orz")
	}

	for _, subscription := range subscriptions {
		if err := s.messageSvc.Send(0, subscription.Uid, objID, types.MessageTypeSubscription, objID, text); err != nil {
			slog.Error("subscription: send message failed", "uid", subscription.Uid, "objID", objID, "error", err.Error())
			return errors.New("发送消息失败Orz")
		}
	}
	return nil
}

// Subscribe 订阅
func (s *SubscriptionService) Subscribe(uid uint, objID string, level int) error {
	// 检查订阅级别
	if level < 0 || level > 3 {
		return errors.New("订阅级别错误Orz")
	}

	// 检查是否已经存在订阅
	var existingSub database.Subscription
	if err := database.DB.Where("uid = ? AND obj = ?", uid, objID).Limit(1).Find(&existingSub).Error; err != nil {
		return errors.New("数据库错误Orz")
	}

	if level == 0 {
		// 取消订阅
		if existingSub.ID == 0 {
			return errors.New("订阅不存在Orz")
		}
		existingSub.Level = 0
		if err := database.DB.Save(&existingSub).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		return nil
	}

	// 获取订阅对象简介
	obj, err := types.NewObj(objID)
	if err != nil {
		return errors.New("获取订阅对象失败Orz")
	}
	text, err := obj.GetText()
	if err != nil {
		return errors.New("获取订阅对象简介失败Orz")
	}

	// 更新已有订阅
	if existingSub.ID != 0 {
		existingSub.Level = database.SubscriptionLevel(level)
		existingSub.Text = text
		if err := database.DB.Save(&existingSub).Error; err != nil {
			return errors.New("数据库错误Orz")
		}
		return nil
	}

	// 创建新订阅
	subscription := database.Subscription{
		Uid:   uid,
		Obj:   objID,
		Level: database.SubscriptionLevel(level),
		Text:  text,
	}

	if err := database.DB.Create(&subscription).Error; err != nil {
		return errors.New("数据库错误Orz")
	}

	return nil
}

// GetSubscriptions 获取订阅列表
func (s *SubscriptionService) GetSubscriptions(uid uint, page uint) ([]types.SubscriptionAPI, error) {
	var subscriptions []database.Subscription
	if err := database.DB.Where("uid = ? AND level > 0", uid).Order("updated_at DESC").Offset(int(page * config.Get().SubscriptionPageSize)).Limit(int(config.Get().SubscriptionPageSize)).Find(&subscriptions).Error; err != nil {
		return nil, errors.New("获取订阅列表失败Orz")
	}

	result := make([]types.SubscriptionAPI, len(subscriptions))
	for i, subscription := range subscriptions {
		result[i] = types.SubscriptionAPI{
			ID:               subscription.ID,
			SubscriptionTime: subscription.UpdatedAt,
			Obj:              subscription.Obj,
			Level:            int(subscription.Level),
			Text:             subscription.Text,
		}
	}

	return result, nil
}
