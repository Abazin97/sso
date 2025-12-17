package email

import (
	"bytes"
	"errors"
	"html/template"
	log "log/slog"
)

type Sender interface {
	Send(input SendEmailInput) error
}

type SendEmailInput struct {
	To      string
	Subject string
	Body    string
}

func (e *SendEmailInput) GenerateBodyFromHTML(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		log.Info("failed to parse file %s:%s", templateFileName, err.Error())

		return err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	e.Body = buf.String()

	return nil
}

func (e *SendEmailInput) Validate() error {
	if e.To == "" {
		return errors.New("empty email address")
	}
	if e.Subject == "" {
		return errors.New("empty email subject")
	}
	if e.Body == "" {
		return errors.New("empty email body")
	}

	return nil
}
