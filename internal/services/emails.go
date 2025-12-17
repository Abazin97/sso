package services

import (
	"fmt"
	"log/slog"
	"sso/internal/config"
	"sso/internal/services/email"
)

const verificationLinkTmpl = "%s/verify?code=%s"

type EmailService struct {
	log    *slog.Logger
	sender email.Sender
	config config.EmailConfig
}

type VerificationEmailInput struct {
	Email            string
	Name             string
	VerificationCode string
	Domain           string
}

func NewEmailService(log *slog.Logger, sender email.Sender, config config.EmailConfig) (*EmailService, error) {
	return &EmailService{log: log, sender: sender, config: config}, nil
}

func (s *EmailService) SendVerificationEmail(input VerificationEmailInput) error {

	subject := fmt.Sprint(s.config.Subjects.VerificationName, input.Name)

	//body := fmt.Sprintf(input.Name, input.VerificationCode)
	templateInput := VerificationEmailInput{Name: input.Name, VerificationCode: input.VerificationCode}
	sendInput := email.SendEmailInput{Subject: subject, To: input.Email}

	if err := sendInput.GenerateBodyFromHTML(s.config.Templates.VerificationCode, templateInput); err != nil {
		return err
	}

	return s.sender.Send(sendInput)
}
