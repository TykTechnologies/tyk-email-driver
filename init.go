package emaildriver

import (
	"errors"
	"html/template"

	"github.com/TykTechnologies/tykcommon-logger"
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

var emailBackendCodeError = errors.New("no backend with this code was found")
var driverInitializationError = errors.New("email driver initialization error")

type EmailBackend interface {
	Init(map[string]string) error
	Send(EmailMeta, interface{}, TykTemplateName, TykTemplateName, string, string) error
}

var EmailBackendCodes = map[string]EmailBackend{
	"mandrill":  &MandrillEmailBackend{},
	"sendgrid":  &SendGridEmailBackend{},
	"mailgun":   &MailgunEmailBackend{},
	"amazonses": &AmazonSESEmailBackend{},
	"smtp":      &SMTPEmailBackend{},
	"mock":      &MockEmailBackend{},
}

func GetEmailBackend(code string) (EmailBackend, error) {
	var thisInterface EmailBackend
	var ok bool

	if thisInterface, ok = EmailBackendCodes[code]; !ok {
		return nil, emailBackendCodeError
	}

	return thisInterface, nil
}
