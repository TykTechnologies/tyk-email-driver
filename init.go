package emaildriver

import (
	"errors"
	"github.com/TykTechnologies/tykcommon-logger"
	"html/template"
)

type TykTemplateName string

type EmailMeta struct {
	RecipientEmail string
	RecipientName  string
	FromEmail      string
	FromName       string
	Subject        string
}

var log = logger.GetLogger().WithField("prefix", "email")

var PortalEmailTemplatesHTML *template.Template
var PortalEmailTemplatesTXT *template.Template

type EmailBackend interface {
	Init(map[string]string) error
	Send(EmailMeta, interface{}, TykTemplateName, TykTemplateName, string, string) error
}

var EmailBackendCodes = map[string]EmailBackend{
	"mandrill":   &MandrillEmailBackend{},
	"sendgrid":   &SendGridEmailBackend{},
	"mailgun":    &MailgunEmailBackend{},
	"amazonses":  &AmazonSESEmailBackend{},
	"smtp":       &SmtpEmailBackend{},
	"mock":       &MockEmailBackend{},
}

func GetEmailBackend(code string) (EmailBackend, error) {
	var thisInterface EmailBackend
	var ok bool

	if thisInterface, ok = EmailBackendCodes[code]; !ok {
		return nil, errors.New("No backend with this code was found")
	}

	return thisInterface, nil
}
