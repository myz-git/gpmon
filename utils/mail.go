// utils/mail.go

package utils

import (
	"gpmon/db"
	"log"

	"gopkg.in/gomail.v2"
)

func SendEmail(ip, dbType, dbName, subject, message string) error {
	if !db.ShouldSendEmail(ip, dbType, dbName) {
		log.Printf("Email already sent recently. Skipping sending email for IP: %s", ip)
		return nil
	}
	m := gomail.NewMessage()

	// 设置发件人、收件人、主题和内容
	m.SetHeader("From", "mon@dbhome.cc")
	m.SetHeader("To", "mayz@vastdata.com.cn")
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	// 设置SMTP服务器的地址、端口和登录凭据
	d := gomail.NewDialer("smtp.qiye.aliyun.com", 587, "mon@dbhome.cc", "Welcome1")
	// d := gomail.NewDialer("smtp.qiye.aliyun.com", 5871, "abc@dbhome.cc", "123")

	// 尝试发送邮件
	if err := d.DialAndSend(m); err != nil {
		log.Println("Failed to send email:", err)
		// 邮件发送失败，设置 ismail=0
		if updateErr := db.UpdateClientInfoIsMail(ip, dbType, dbName, 0); updateErr != nil {
			log.Println("Failed to update ismail after email send failure:", updateErr)
		}
		return err
	}

	// 邮件发送成功，设置 ismail=1
	if updateErr := db.UpdateClientInfoIsMail(ip, dbType, dbName, 1); updateErr != nil {
		log.Println("Failed to update ismail after email send success:", updateErr)
	}

	// 邮件发送成功后更新 last_email_sent
	if err := db.UpdateLastEmailSent(ip, dbType, dbName); err != nil {
		log.Printf("Failed to update last email sent timestamp for IP: %s", ip)
	}
	return nil
}
