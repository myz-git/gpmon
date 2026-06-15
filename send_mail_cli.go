// send_mail_cli.go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"time"

	"gpmon/utils"

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
	wd := path.Dir(filename)
	/*** End ***/

	/*** 设定dbfile路径 ***/
	dbf := wd + "/messages.db"

	var err error
	db, err := sql.Open("sqlite3", dbf)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()
	/*** End ***/

	// 查询邮件配置
	var cfg MailConfig
	err = db.QueryRow("SELECT sender, recipient, cc, smtp_server, smtp_port, smtp_user, smtp_password FROM mail_cfg WHERE isenable = 1 LIMIT 1").Scan(
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

// SendEmailWithHTML 发送支持HTML格式的邮件
func SendEmailWithHTML(subject, textBody, htmlBody, htmlFile string, customRecipient string) error {
	cfg, err := ReadMailConfig()
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no enabled mail configuration found")
		}
		return fmt.Errorf("failed to read mail config: %v", err)
	}

	m := gomail.NewMessage()

	// 设置发件人
	m.SetHeader("From", cfg.Sender)

	// 设置收件人（优先使用自定义收件人）
	recipient := cfg.Recipient
	if customRecipient != "" {
		recipient = customRecipient
	}
	if toAddrs := utils.ParseEmailAddresses(recipient); len(toAddrs) > 0 {
		m.SetHeader("To", toAddrs...)
	}
	if ccAddrs := utils.ParseEmailAddresses(cfg.CC); len(ccAddrs) > 0 {
		m.SetHeader("Cc", ccAddrs...)
	}

	// 设置主题
	m.SetHeader("Subject", subject)

	// 设置邮件内容
	if htmlBody != "" {
		// 如果有HTML内容，设置多部分邮件
		m.SetBody("text/plain", textBody)
		m.AddAlternative("text/html", htmlBody)
	} else if htmlFile != "" {
		// 如果有HTML文件，读取并设置
		htmlContent, err := ioutil.ReadFile(htmlFile)
		if err != nil {
			return fmt.Errorf("failed to read HTML file: %v", err)
		}
		m.SetBody("text/plain", textBody)
		m.AddAlternative("text/html", string(htmlContent))
	} else {
		// 纯文本邮件
		m.SetBody("text/plain", textBody)
	}

	// 设置SMTP配置
	d := gomail.NewDialer(cfg.SMTPServer, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword)

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func main() {
	var (
		subject    = flag.String("subject", "", "邮件主题")
		textBody   = flag.String("text", "", "文本内容")
		textFile   = flag.String("text-file", "", "文本内容文件路径")
		htmlBody   = flag.String("html", "", "HTML内容")
		htmlFile   = flag.String("html-file", "", "HTML内容文件路径")
		recipient  = flag.String("to", "", "收件人（可选，默认使用数据库配置）")
		testConfig = flag.Bool("test-config", false, "测试邮件配置")
		help       = flag.Bool("help", false, "显示帮助信息")
	)

	flag.Parse()

	if *help {
		fmt.Println("GPMon 邮件发送工具")
		fmt.Println("")
		fmt.Println("用法:")
		fmt.Println("  send_mail_cli [选项]")
		fmt.Println("")
		fmt.Println("选项:")
		fmt.Println("  -subject string      邮件主题")
		fmt.Println("  -text string         文本内容")
		fmt.Println("  -text-file string    文本内容文件路径")
		fmt.Println("  -html string         HTML内容")
		fmt.Println("  -html-file string    HTML内容文件路径")
		fmt.Println("  -to string           收件人（可选）")
		fmt.Println("  -test-config         测试邮件配置")
		fmt.Println("  -help                显示帮助信息")
		fmt.Println("")
		fmt.Println("示例:")
		fmt.Println("  # 发送纯文本邮件")
		fmt.Println("  send_mail_cli -subject \"测试邮件\" -text \"这是测试内容\"")
		fmt.Println("")
		fmt.Println("  # 发送HTML邮件")
		fmt.Println("  send_mail_cli -subject \"监控报告\" -text-file report.txt -html-file report.html")
		fmt.Println("")
		fmt.Println("  # 测试邮件配置")
		fmt.Println("  send_mail_cli -test-config")
		fmt.Println("")
		return
	}

	if *testConfig {
		fmt.Println("测试邮件配置...")
		cfg, err := ReadMailConfig()
		if err != nil {
			fmt.Printf("❌ 读取邮件配置失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ 邮件配置读取成功:\n")
		fmt.Printf("   发件人: %s\n", cfg.Sender)
		fmt.Printf("   收件人: %s\n", cfg.Recipient)
		fmt.Printf("   抄送: %s\n", cfg.CC)
		fmt.Printf("   SMTP服务器: %s:%d\n", cfg.SMTPServer, cfg.SMTPPort)
		fmt.Printf("   SMTP用户: %s\n", cfg.SMTPUser)

		// 发送测试邮件
		testSubject := "GPMon 邮件配置测试"
		testMessage := fmt.Sprintf("这是一封测试邮件，用于验证 GPMon 邮件配置是否正确。\n\n发送时间: %s\n配置测试成功！",
			time.Now().Format("2006-01-02 15:04:05"))

		err = SendEmailWithHTML(testSubject, testMessage, "", "", "")
		if err != nil {
			fmt.Printf("❌ 发送测试邮件失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ 测试邮件发送成功！")
		return
	}

	if *subject == "" {
		fmt.Println("错误: 必须指定邮件主题")
		fmt.Println("使用 -help 查看帮助信息")
		os.Exit(1)
	}

	// 读取文本内容
	var textContent string
	if *textFile != "" {
		content, err := ioutil.ReadFile(*textFile)
		if err != nil {
			fmt.Printf("错误: 无法读取文本文件 %s: %v\n", *textFile, err)
			os.Exit(1)
		}
		textContent = string(content)
	} else if *textBody != "" {
		textContent = *textBody
	} else {
		textContent = "GPMon 自动生成的邮件"
	}

	// 发送邮件
	err := SendEmailWithHTML(*subject, textContent, *htmlBody, *htmlFile, *recipient)
	if err != nil {
		fmt.Printf("❌ 发送邮件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 邮件发送成功！")
}
