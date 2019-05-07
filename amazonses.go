package emaildriver

import (
	"bytes"
	"errors"
	"github.com/sourcegraph/go-ses"
)

type AmazonSESEmailBackend struct {
	isEnabled        bool
	endpoint         string
	accessKeyId      string
	secretAccessKey  string
	config           ses.Config
}

func (s *AmazonSESEmailBackend) Init(conf map[string]string) error {
	var ok bool

	if conf == nil {
		return errors.New("AmazonSESEmailBackend requires a configuration map")
	}

	s.endpoint, ok = conf["Endpoint"]
	if !ok {
		return errors.New("No Amazon SES endpoint defined, emails will fail")
	}

	s.accessKeyId, ok = conf["AccessKeyId"]
	if !ok {
		return errors.New("No Amazon SES access key defined, emails will fail")
	}

	s.secretAccessKey, ok = conf["SecretAccessKey"]
	if !ok {
		return errors.New("No Amazon SES secret access key defined, emails will fail")
	}

	s.config = ses.Config{
	  Endpoint:        s.endpoint,
	  AccessKeyID:     s.accessKeyId,
	  SecretAccessKey: s.secretAccessKey,
	}
	s.isEnabled = true
	return nil
}

func (s *AmazonSESEmailBackend) Send(emailMeta EmailMeta, emailData interface{}, textTemplateName TykTemplateName, htmlTemplateName TykTemplateName, OrgId string, Styles string) error {

	if !s.isEnabled {
		log.Warning("Plese check your email settings and restart Tyk Dashboard in order for notifications to work")
		return errors.New("Amazon SES email driver not initialised correctly")
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

	from := "\"" + emailMeta.FromName + "\" <" + emailMeta.FromEmail + ">"
	to := "\"" + emailMeta.RecipientName + "\" <" + emailMeta.RecipientEmail + ">"
	response, err := s.config.SendEmailHTML(from, to, emailMeta.Subject, txtDoc.String(), htmlDoc.String())

	log.Info("Email sending (AmazonSES)")
	if err != nil {
		log.Error("Amazon SES error: ", err)
		log.Error("Response: ", response)
		return err
	}

	return nil
}
