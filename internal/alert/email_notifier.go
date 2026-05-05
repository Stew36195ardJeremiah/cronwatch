package alert

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailConfig holds SMTP configuration for email notifications.
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
	To       []string
}

// EmailNotifier sends alert notifications via email.
type EmailNotifier struct {
	cfg  EmailConfig
	dial func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error
}

// NewEmailNotifier creates a new EmailNotifier with the given SMTP config.
func NewEmailNotifier(cfg EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		cfg:  cfg,
		dial: smtp.SendMail,
	}
}

// Send dispatches an alert email to all configured recipients.
func (e *EmailNotifier) Send(a Alert) error {
	if len(e.cfg.To) == 0 {
		return fmt.Errorf("email notifier: no recipients configured")
	}

	subject := fmt.Sprintf("[cronwatch][%s] %s", strings.ToUpper(string(a.Level)), a.Job)
	body := fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s\r\n",
		strings.Join(e.cfg.To, ", "),
		e.cfg.From,
		subject,
		a.Message,
	)

	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)
	var auth smtp.Auth
	if e.cfg.Username != "" {
		auth = smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.SMTPHost)
	}

	if err := e.dial(addr, auth, e.cfg.From, e.cfg.To, []byte(body)); err != nil {
		return fmt.Errorf("email notifier: failed to send: %w", err)
	}
	return nil
}
