// file: email/smtp_service.go
package email

import (
	"context"
	"fmt"
	"net/smtp"
	"time"
)

type EmailService interface {
	SendVerificationEmail(ctx context.Context, to, token string) error
	SendPasswordResetEmail(ctx context.Context, to, token string) error
}

type SMTPConfig struct {
	Host     string        // e.g. "smtp.gmail.com"
	Port     int           // e.g. 587
	Username string        // SMTP username
	Password string        // SMTP password
	From     string        // e.g. "no-reply@carsawa.com"
	Timeout  time.Duration // network timeout
}

type SMTPEmailService struct {
	cfg  SMTPConfig
	auth smtp.Auth
}

func NewSMTPEmailService(cfg SMTPConfig) *SMTPEmailService {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return &SMTPEmailService{cfg: cfg, auth: auth}
}

func (s *SMTPEmailService) SendVerificationEmail(ctx context.Context, to, token string) error {
	subject := "Verify your Carsawa account"
	verifyURL := fmt.Sprintf("https://yourdomain.com/verify-email?token=%s", token)
	body := fmt.Sprintf(
		"Hello,\n\nPlease verify your email by clicking the link below:\n%s\n\nIf you didn't request this, please ignore this email.\n",
		verifyURL,
	)
	return s.send(ctx, to, subject, body)
}

func (s *SMTPEmailService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	subject := "Carsawa Password Reset"
	resetURL := fmt.Sprintf("https://yourdomain.com/reset-password?token=%s", token)
	body := fmt.Sprintf(
		"Hello,\n\nYou requested a password reset. Click the link below to set a new password:\n%s\n\nIf you didn't request this, you can safely ignore this email.\n",
		resetURL,
	)
	return s.send(ctx, to, subject, body)
}

func (s *SMTPEmailService) send(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/plain; charset=\"utf-8\"\r\n"+
			"\r\n%s",
		s.cfg.From, to, subject, body,
	))

	type dialer interface {
		SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error
	}
	ch := make(chan error, 1)
	go func() {
		ch <- smtp.SendMail(addr, s.auth, s.cfg.From, []string{to}, msg)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(s.cfg.Timeout):
		return fmt.Errorf("email send timed out after %s", s.cfg.Timeout)
	}
}
