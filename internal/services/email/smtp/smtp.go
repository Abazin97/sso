package smtp

import (
	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"

	"sso/internal/services/email"
)

type SMTPService struct {
	from string
	pass string
	host string
	port int
}

func NewSMTPService(from string, pass, host string, port int) (*SMTPService, error) {
	return &SMTPService{from: from, pass: pass, host: host, port: port}, nil
}

func (s *SMTPService) Send(input email.SendEmailInput) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", s.from)
	msg.SetHeader("To", input.To)
	msg.SetHeader("Subject", input.Subject)
	msg.SetBody("text/html", input.Body)

	dialer := gomail.NewDialer(s.host, s.port, s.from, s.pass)
	if err := dialer.DialAndSend(msg); err != nil {
		return errors.Wrap(err, "failed to send email via smtp")
	}

	return nil
}
