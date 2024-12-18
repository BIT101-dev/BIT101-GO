/*
 * @Author: flwfdd
 * @Date: 2023-03-20 13:44:32
 * @LastEditTime: 2023-03-20 14:27:33
 * @Description: _(:з」∠)_
 */
package mail

import (
	"fmt"
	"net/smtp"
	"strings"

	"BIT101-GO/util/config"
)

// Send 发送邮件
func Send(to, subject, body string) error {
	// 从配置加载邮件服务器信息
	mailHost := config.Config.Mail.Host
	mailUser := config.Config.Mail.User
	mailPass := config.Config.Mail.Password
	mailPort := config.Config.Mail.Port // 使用配置中的端口

	// 验证配置信息是否完整
	if mailHost == "" || mailUser == "" || mailPass == "" || mailPort == "" {
		return fmt.Errorf("邮件配置不完整")
	}

	// 构建SMTP认证信息
	auth := smtp.PlainAuth("", mailUser, mailPass, mailHost)

	// 构建邮件内容
	msg := fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		to, mailUser, subject, body,
	)

	// 发送邮件
	serverAddr := fmt.Sprintf("%s:%s", mailHost, mailPort)
	recipients := strings.Split(to, ",") // 支持多个收件人
	return smtp.SendMail(serverAddr, auth, mailUser, recipients, []byte(msg))
}
