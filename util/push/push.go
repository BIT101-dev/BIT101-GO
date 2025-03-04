package push

import (
	"BIT101-GO/database"
	"BIT101-GO/util/config"

	"errors"
	"fmt"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func GetRequestPubkey() (string, error) {
	if config.GetConfig().WebPushKeys.Public == "" || config.GetConfig().WebPushKeys.Private == "" {
		return "", errors.New("配置错误Orz")
	}
	return config.GetConfig().WebPushKeys.Public, nil
}

func HandleRegister(sub webpush.Subscription, uid uint) error {
	if config.GetConfig().WebPushKeys.Public == "" || config.GetConfig().WebPushKeys.Private == "" {
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
	if config.GetConfig().WebPushKeys.Public == "" || config.GetConfig().WebPushKeys.Private == "" {
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

	fmt.Printf("[Push] Trying sending notification to uid: %d, endpoint: %s..\n", sub.Uid, sub.Endpoint[:30])

	resp, err := webpush.SendNotification(message, &subscription, &webpush.Options{
		VAPIDPublicKey:  config.GetConfig().WebPushKeys.Public,
		VAPIDPrivateKey: config.GetConfig().WebPushKeys.Private,
		TTL:             30,
	})

	if err != nil {
		fmt.Printf("[Push] Send notification to uid: %d failed: %s\n", sub.Uid, err)
		fmt.Printf("[Push] With endpoint: %s.., auth: %s, p256dh: %s\n", sub.Endpoint[:30], sub.Auth, sub.P256dh)
		return errors.New("推送错误Orz")
	}

	defer resp.Body.Close()
	fmt.Printf("[Push] Send notification to uid: %d, endpoint: %s.. succeed\n", sub.Uid, sub.Endpoint[:30])
	return nil
}

func HandleUnregister(sub webpush.Subscription, uid uint) error {
	if config.GetConfig().WebPushKeys.Public == "" || config.GetConfig().WebPushKeys.Private == "" {
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
