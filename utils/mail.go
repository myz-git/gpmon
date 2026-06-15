// utils/mail.go

package utils

import (
	"database/sql"
	"log"
	"net/mail"
	"path"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/gomail.v2"
)

// MailConfig 结构体映射 mail_cfg 表的配置
type MailConfig struct {
	Sender       string
	Recipient    string
	CC           string
	SMTPServer   string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
}

// ReadMailConfig 从数据库中读取邮件配置
func ReadMailConfig() (*MailConfig, error) {
	/*** 获取项目根路径 ***/
	_, filename, _, _ := runtime.Caller(0)
	wd := path.Dir(path.Dir(filename))
	// log.Printf("wd:  %s", wd)
	/*** End ***/

	/*** 设定dbfile路径 ***/
	dbf := wd + "/messages.db"

	var err error
	db, err := sql.Open("sqlite3", dbf)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	/*** End ***/

	// 查询邮件配置
	var cfg MailConfig
	err = db.QueryRow("SELECT sender, recipient, cc, smtp_server, smtp_port, smtp_user, smtp_password  FROM mail_cfg where isenable = 1 LIMIT 1").Scan(
		&cfg.Sender,
		&cfg.Recipient,
		&cfg.CC,
		&cfg.SMTPServer,
		&cfg.SMTPPort,
		&cfg.SMTPUser,
		&cfg.SMTPPassword,
	)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ParseEmailAddresses 将逗号/分号分隔的邮箱字符串解析为地址列表
func ParseEmailAddresses(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	if list, err := mail.ParseAddressList(s); err == nil {
		addrs := make([]string, 0, len(list))
		for _, a := range list {
			addrs = append(addrs, a.Address)
		}
		return addrs
	}

	s = strings.ReplaceAll(s, ";", ",")
	parts := strings.Split(s, ",")
	addrs := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if a, err := mail.ParseAddress(p); err == nil {
			addrs = append(addrs, a.Address)
		} else {
			addrs = append(addrs, p)
		}
	}
	return addrs
}

// SendEmail 发送邮件并根据发送结果更新数据库
func SendEmail(ip, dbType, dbName, subject, message string) error {
	cfg, err := ReadMailConfig()
	if err != nil {
		// 如果没有找到启用的邮件配置，则不发送邮件
		if err == sql.ErrNoRows {
			log.Println("No enabled mail configuration found. Email not sent.")
			return nil // 返回 nil 以避免错误，但根据您的业务逻辑您可能想返回一个特定的错误
		}
		// 如果是其他错误，返回它
		return err
	}

	m := gomail.NewMessage()

	// 设置发件人、收件人、主题和内容
	m.SetHeader("From", cfg.Sender)
	if toAddrs := ParseEmailAddresses(cfg.Recipient); len(toAddrs) > 0 {
		m.SetHeader("To", toAddrs...)
	}
	if ccAddrs := ParseEmailAddresses(cfg.CC); len(ccAddrs) > 0 {
		m.SetHeader("Cc", ccAddrs...)
	}
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", message)

	// // 设置SMTP服务器的地址、端口和登录凭据
	d := gomail.NewDialer(cfg.SMTPServer, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword)

	// // 设置发件人、收件人、主题和内容
	// m.SetHeader("From", "mon@dbhome.cc")
	// m.SetHeader("To", "mayz@vastdata.com.cn")
	// m.SetHeader("Subject", subject)
	// m.SetBody("text/plain", message)

	// // 设置SMTP服务器的地址、端口和登录凭据
	// d := gomail.NewDialer("smtp.qiye.aliyun.com", 587, "mon@dbhome.cc", "Welcome1")

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
