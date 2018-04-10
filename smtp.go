package emaildriver

import (
	"gopkg.in/gomail.v2"
	"bytes"
	"crypto/tls"
	"errors"
	"strconv"
	"net"
)

type SmtpEmailBackend struct {
	isEnabled  bool
	d gomail.Dialer
	smtpAddress string
}

func (m *SmtpEmailBackend) Init(conf map[string]string) error {
	var tlsConn bool

	if conf == nil {
		return errors.New("SmtpEmailBackend requires a configuration map")
	}


	smtpUser, ok := conf["SMTPUsername"]
	if !ok {
		return errors.New("SMTPPassword not defined")
	}

	smtpPass, ok := conf["SMTPPassword"]
	if !ok {
		return errors.New("SMTPPassword not defined, emails will fail")
	}

	host, portString, err := net.SplitHostPort(conf["SMTPAddress"])
	if err != nil {
		return err
	}

	m.smtpAddress = conf["SMTPAddress"]

	port, err := strconv.Atoi(portString)
	if err != nil {
		return err
	}

	tlsConnTxt, ok := conf["tls"]
	if !ok {
		tlsConn = false
	} else {
		tlsConn, _ = strconv.ParseBool(tlsConnTxt)
	}

	d := gomail.NewDialer(host, port, smtpUser, smtpPass)
	if tlsConn {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if err != nil {
		log.Error(err)
		return err
	}

	m.isEnabled = true
	return nil
}

func (m *SmtpEmailBackend) Send(emailMeta EmailMeta, emailData interface{}, textTemplateName TykTemplateName, htmlTemplateName TykTemplateName, OrgId string, Styles string) error {

	if !m.isEnabled {
		log.Warning("SMTP driver not initialized.")
		return errors.New("Driver not initialised correctly")
	}

	// Generate strings from templates
	var txtDoc bytes.Buffer

	type superEmailData struct {
		Data   interface{}
		Styles string
	}

	thisData := superEmailData{Data: emailData}
	thisData.Styles = Styles

	txtErr := PortalEmailTemplatesTXT.ExecuteTemplate(&txtDoc, string(textTemplateName), emailData)
	if txtErr != nil {
		log.Error("HTML Template error: ", txtErr)
		return txtErr
	}

	fromStr := emailMeta.FromName + " <" + emailMeta.FromEmail + ">"
	recStr := emailMeta.RecipientName + " <" + emailMeta.RecipientEmail + ">"
	subj := emailMeta.Subject
	msg := txtDoc.String()

	message := gomail.NewMessage()
	message.SetHeader("From", fromStr)
	message.SetHeader("To", recStr)
	message.SetHeader("Subject", subj)
	message.SetBody("text/html", msg)

	if err := m.d.DialAndSend(message); err != nil {
		panic(err)
	}

	log.Debugf("email sent addr: %s, from: %s, to: %s", m.smtpAddress, fromStr, recStr)
	if err != nil {
		log.Error("SMTP error: ", err)
		return err
	}

	return nil
}
