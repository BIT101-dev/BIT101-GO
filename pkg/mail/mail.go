/*
 * @Author: flwfdd
 * @Date: 2023-03-20 13:44:32
 * @LastEditTime: 2025-03-18 17:17:57
 * @Description: _(:з」∠)_
 */
package mail

import (
	"BIT101-GO/config"
	"crypto/tls"
	"fmt"
	"net/smtp"
)

// Send 发送邮件 参考https://juejin.cn/post/7433299013875302450
func Send(to string, title string, text string) error {
	mail_host := config.Get().Mail.Host
	mail_user := config.Get().Mail.User
	mail_pass := config.Get().Mail.Password

	// 编写发送的消息
	msg := []byte("To: " + to +
		"\r\nFrom: " + mail_user +
		"\r\nSubject: " + title +
		"\r\n\r\n" + text)

	// 设置 PlainAuth
	auth := smtp.PlainAuth("", mail_user, mail_pass, mail_host)

	// 创建 tls 配置
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         mail_host,
	}

	// 连接到 SMTP 服务器
	conn, err := tls.Dial("tcp", mail_host+":465", tlsconfig)
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, mail_host)
	if err != nil {
		return fmt.Errorf("SMTP 客户端创建失败: %v", err)
	}
	defer client.Quit()

	// 使用 auth 进行认证
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("认证失败: %v", err)
	}

	// 设置发件人和收件人
	if err = client.Mail(mail_user); err != nil {
		return fmt.Errorf("发件人设置失败: %v", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("收件人设置失败: %v", err)
	}

	// 写入邮件内容
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("数据写入失败: %v", err)
	}
	defer wc.Close()

	_, err = wc.Write(msg)
	if err != nil {
		return fmt.Errorf("消息发送失败: %v", err)
	}

	return nil
}
