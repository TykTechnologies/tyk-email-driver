package emaildriver

import (
	"bytes"
	"errors"

	ma "github.com/keighl/mandrill"
)

type MandrillEmailBackend struct {
	clientKey string
	isEnabled bool
}

func (m *MandrillEmailBackend) Init(conf map[string]string) error {
	var ok bool

	if conf == nil {
		return errors.New("MandrillEmailBackend requires a configuration map")
	}

	m.clientKey, ok = conf["ClientKey"]
	if !ok {
		return errors.New("No Mandrill client key defined, emails will fail")
	}

	m.isEnabled = true
	return nil
}

func (m *MandrillEmailBackend) Send(emailMeta EmailMeta, emailData interface{}, textTemplateName TykTemplateName, htmlTemplateName TykTemplateName, OrgId string, Styles string) error {

	if !m.isEnabled {
		log.Warning("Plese check your email settings and restart Tyk Dashboard in order for notifications to work")
		return errors.New("Driver not initialised correctly")
	}

	// Generate strings from templates
	var htmlDoc, txtDoc bytes.Buffer

	htmlBody := emailMeta.HTMLBody

	if htmlBody == "" {
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

		htmlBody = htmlDoc.String()
	}

	txtBody := emailMeta.TextBody
	if txtBody == "" {
		txtErr := PortalEmailTemplatesTXT.ExecuteTemplate(&txtDoc, string(textTemplateName), emailData)
		if txtErr != nil {
			log.Error("HTML Template error: ", txtErr)
			return txtErr
		}

		txtBody = txtDoc.String()
	}

	// Do sending
	client := ma.ClientWithKey(m.clientKey)
	message := &ma.Message{}
	message.AddRecipient(emailMeta.RecipientEmail, emailMeta.RecipientName, "to")
	message.FromEmail = emailMeta.FromEmail
	message.FromName = emailMeta.FromName
	message.Subject = emailMeta.Subject
	message.HTML = htmlBody
	message.Text = txtBody

	responses, err := client.MessagesSend(message)

	//log.Warning("Email sending (Mandrill): ", responses)
	if err != nil {
		log.Error("Mandrill error: ", err)
		log.Error("Responses: ", responses)
		return err
	}

	return nil
}
