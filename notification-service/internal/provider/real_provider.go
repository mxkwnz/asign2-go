package provider

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type SMTPEmailSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPEmailSender() *SMTPEmailSender {
	return &SMTPEmailSender{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

func (s *SMTPEmailSender) Send(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := []byte(fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\n\r\n%s",
		to, subject, body,
	))

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, msg); err != nil {
		return fmt.Errorf("SMTP send failed: %w", err)
	}

	log.Printf("[SMTPProvider] Email sent to=%s subject=%q", to, subject)
	return nil
}
