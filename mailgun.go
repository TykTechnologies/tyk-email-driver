package emaildriver

import (
	"bytes"
	"errors"
	mg "github.com/mailgun/mailgun-go"
)

type MailgunEmailBackend struct {
	isEnabled        bool
	mailgunDomain    string
	privateClientKey string
	publicClientKey  string
	mailGunClient    mg.Mailgun
}

func (m *MailgunEmailBackend) Init(conf map[string]string) error {
	var ok bool

	if conf == nil {
		return errors.New("MailgunEmailBackend requires a configuration map")
	}

	m.mailgunDomain, ok = conf["Domain"]
	if !ok {
		return errors.New("No Mailgun domain defined, emails will fail")
	}

	m.privateClientKey, ok = conf["PrivateKey"]
	if !ok {
		return errors.New("No Mailgun private client key defined, emails will fail")
	}

	m.publicClientKey, ok = conf["PublicKey"]
	if !ok {
		return errors.New("No Mailgun public client key defined, emails will fail")
	}

	m.mailGunClient = mg.NewMailgun(m.mailgunDomain, m.privateClientKey, m.publicClientKey)
	m.isEnabled = true
	return nil
}

func (m *MailgunEmailBackend) Send(emailMeta EmailMeta, emailData interface{}, textTemplateName TykTemplateName, htmlTemplateName TykTemplateName, OrgId string, Styles string) error {

	if !m.isEnabled {
		log.Warning("Plese check your email settings and restart Tyk Dashboard in order for notifications to work")
		return errors.New("Driver not initialised correctly")
	}

	if !EnableEmailNotifications {
		log.Info("Email notifications disabled, skipping Send()")
		return nil
	}

	// Generate strings from templates
	var htmlDoc, txtDoc bytes.Buffer

	type superEmailData struct {
		Data   interface{}
		Styles string
	}

	thisData := superEmailData{Data: emailData}
	// Pull custom CSS
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

	fromStr := emailMeta.FromName + " <" + emailMeta.FromEmail + ">"
	recStr := emailMeta.RecipientName + " <" + emailMeta.RecipientEmail + ">"
	msg := mg.NewMessage(fromStr, emailMeta.Subject, txtDoc.String(), recStr)
	response, id, err := m.mailGunClient.Send(msg)

	log.Info("Email sending (MailGun): ", id)
	if err != nil {
		log.Error("Mailgun API error: ", err)
		log.Error("Response: ", response)
		return err
	}

	return nil
}
