package push

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"

	"errors"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func GetRequestPubkey() (string, error) {
	if config.Config.WebPushKeys.Public == "" || config.Config.WebPushKeys.Private == "" {
		return "", errors.New("配置错误Orz")
	}
	return config.Config.WebPushKeys.Public, nil
}

func HandleRegister(sub webpush.Subscription, uid uint) error {
	if config.Config.WebPushKeys.Public == "" || config.Config.WebPushKeys.Private == "" {
		return errors.New("配置错误Orz")
	}

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

func Send(sub database.WebPushSubscription, message []byte) error {
	if config.Config.WebPushKeys.Public == "" || config.Config.WebPushKeys.Private == "" {
		return errors.New("配置错误Orz")
	}
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
		VAPIDPublicKey:  config.Config.WebPushKeys.Public,
		VAPIDPrivateKey: config.Config.WebPushKeys.Private,
		TTL:             30,
	})

	if err != nil {
		return errors.New("推送错误Orz")
	}

	defer resp.Body.Close()
	return nil
}

func HandleUnregister(sub webpush.Subscription, uid uint) error {
	if config.Config.WebPushKeys.Public == "" || config.Config.WebPushKeys.Private == "" {
		return errors.New("配置错误Orz")
	}

	subscription := database.WebPushSubscription{
		Uid:            uid,
		Endpoint:       sub.Endpoint,
		ExpirationTime: "null",
		Auth:           sub.Keys.Auth,
		P256dh:         sub.Keys.P256dh,
	}

	var target database.WebPushSubscription
	database.DB.Unscoped().Where("uid = ?", uid).Where(
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
