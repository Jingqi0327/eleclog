package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	smtpAuthAddress   = "smtp.qq.com"     // SMTP服务器地址
	smtpServerAddress = "smtp.qq.com:587" // SMTP服务器地址和端口
)

type EmailSender interface {
	SendEmail(
		subject string, // 邮件主题
		content string, // 邮件内容
		to []string, // 收件人列表
		cc []string, // 抄送人列表
		bcc []string, // 密送人列表
		attachFile []string, // 附件列表
	) error
}

type QQmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func NewQQmailSender(name, fromEmailAddress, fromEmailPassword string) EmailSender {
	return &QQmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

func (sender *QQmailSender) SendEmail(
	subject string, // 邮件主题
	content string, // 邮件内容
	to []string, // 收件人列表
	cc []string, // 抄送人列表
	bcc []string, // 密送人列表
	attachFile []string, // 附件列表
) error {
	e := email.NewEmail()
	e.From = sender.name + "<" + sender.fromEmailAddress + ">"
	e.Subject = subject
	e.HTML = []byte(content) // 邮件内容可以是HTML格式
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	// 添加附件
	for _, file := range attachFile {
		_, err := e.AttachFile(file)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", file, err)
		}
	}

	// 验证SMTP服务器的身份
	smtpauth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPassword, smtpAuthAddress)
	// 发送邮件
	return e.Send(smtpServerAddress, smtpauth)
}
