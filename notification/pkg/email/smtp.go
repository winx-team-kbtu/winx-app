package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPMailer struct {
	host      string
	port      int
	username  string
	password  string
	fromEmail string
	fromName  string
}

func NewSMTPMailer(host string, port int, username, password, fromEmail, fromName string) *SMTPMailer {
	return &SMTPMailer{
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

func (m *SMTPMailer) FromName() string {
	if strings.TrimSpace(m.fromName) != "" {
		return m.fromName
	}

	return "Winx Notifications"
}

func (m *SMTPMailer) IsConfigured() bool {
	host := strings.TrimSpace(m.host)

	return host != "" && host != "smtp.example.com" && m.port != 0 && strings.TrimSpace(m.fromEmail) != ""
}

func (m *SMTPMailer) Send(_ context.Context, to, subject, body string) error {
	if strings.TrimSpace(m.host) == "" {
		return fmt.Errorf("smtp host is required")
	}
	if m.port == 0 {
		return fmt.Errorf("smtp port is required")
	}
	if strings.TrimSpace(m.fromEmail) == "" {
		return fmt.Errorf("smtp from_email is required")
	}
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("recipient email is required")
	}

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		m.fromHeader(),
		to,
		subject,
		body,
	))

	var auth smtp.Auth
	if strings.TrimSpace(m.username) != "" || strings.TrimSpace(m.password) != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	return smtp.SendMail(addr, auth, m.fromEmail, []string{to}, msg)
}

func (m *SMTPMailer) fromHeader() string {
	if strings.TrimSpace(m.fromName) == "" {
		return m.fromEmail
	}

	return fmt.Sprintf("%s <%s>", m.fromName, m.fromEmail)
}
