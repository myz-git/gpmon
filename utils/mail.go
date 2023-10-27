// utils/mail.go

package utils

import (
	"log"

	"gopkg.in/gomail.v2"
)

func SendEmail(subject, message string) error {
	m := gomail.NewMessage()

	// 设置发件人、收件人、主题和内容
	m.SetHeader("From", "mon@dbhome.cc")
	m.SetHeader("To", "mayz@vastdata.com.cn")
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	// 设置SMTP服务器的地址、端口和登录凭据
	// d := gomail.NewDialer("smtp.qiye.aliyun.com", 587, "mon@dbhome.cc", "Welcome1")
	d := gomail.NewDialer("smtp.qiye.aliyun.com", 5871, "abc@dbhome.cc", "123")

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		log.Println("Failed to send email:", err)
		return err
	}
	return nil
}
