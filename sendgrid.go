package emaildriver

import (
	"bytes"
	"errors"

	"github.com/sendgrid/sendgrid-go"
)

type SendGridEmailBackend struct {
	clientKey      string
	isEnabled      bool
	sendGridClient *sendgrid.SGClient
}

func (m *SendGridEmailBackend) Init(conf map[string]string) error {
	var ok bool

	if conf == nil {
		return errors.New("SendGridEmailBackend requires a configuration map")
	}

	m.clientKey, ok = conf["ClientKey"]
	if !ok {
		return errors.New("No SendGrid client key defined, emails will fail")
	}

	m.sendGridClient = sendgrid.NewSendGridClientWithApiKey(m.clientKey)
	m.isEnabled = true
	return nil
}

func (m *SendGridEmailBackend) Send(emailMeta EmailMeta, emailData interface{}, textTemplateName TykTemplateName, htmlTemplateName TykTemplateName, OrgId string, Styles string) error {

	if !m.isEnabled {
		log.Warning("Plese check your email settings and restart Tyk Dashboard in order for notifications to work")
		return errors.New("Driver not initialised correctly")
	}

	// Generate strings from templates
	var htmlDoc, txtDoc bytes.Buffer

	type superEmailData struct {
		Data   interface{}
		Styles string
	}

	thisData := superEmailData{Data: emailData}
	thisData.Styles = Styles

	htmlErr := PortalEmailTemplatesHTML.ExecuteTemplate(&htmlDoc, string(htmlTemplateName), thisData)
	if htmlErr != nil {
		log.Error("HTML Template error: ", htmlErr)
		return htmlErr
	}

	txtErr := PortalEmailTemplatesTXT.ExecuteTemplate(&txtDoc, string(textTemplateName), emailData)
	if txtErr != nil {
		log.Error("HTML Template error: ", txtErr)
		return txtErr
	}

	// Do sending
	message := sendgrid.NewMail()
	message.AddTo(emailMeta.RecipientEmail)
	message.AddToName(emailMeta.RecipientName)
	message.SetSubject(emailMeta.Subject)
	message.SetText(txtDoc.String())
	message.SetHTML(htmlDoc.String())
	message.SetFrom(emailMeta.FromEmail)
	message.SetFromName(emailMeta.FromName)
	err := m.sendGridClient.Send(message)

	log.Info("Email sending (SendGrid): ", emailMeta.RecipientEmail)
	if err != nil {
		log.Error("SendGrid API error: ", err)
		return err
	}

	return nil
}
