package emaildriver

import (
	"bytes"
	"errors"
	"net/smtp"
	"crypto/tls"
	"log"
	"net"
)

type SmtpEmailBackend struct {
	isEnabled  bool
	smtpAuth   smtp.Auth
	serverName string
	conn       net.Dialer
}

func (m *SmtpEmailBackend) Init(conf map[string]string) error {
	var ok bool
	var tlsConn bool
	var login string
	var password string
	var port string
	var err error

	if conf == nil {
		return errors.New("SmtpEmailBackend requires a configuration map")
	}

	m.serverName, ok = conf["servername"]
	if !ok {
		return errors.New("No servername defined, emails will fail")
	}

	login, ok = conf["login"]
	if !ok {
		return errors.New("No login defined, emails will fail")
	}

	password, ok = conf["password"]
	if !ok {
		return errors.New("No password defined, emails will fail")
	}

	port, ok = conf["port"]
	if !ok {
		port = 25
	}

	tlsConn, ok = conf["tls"]
	if !ok {
		tlsConn = false
	}

	m.smtpAuth = smtp.PlainAuth("", login, password, m.serverName)

	if (tlsConn) {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName: m.serverName,
		}
		m.conn, err = tls.Dial("tcp", m.serverName + ":" + port, tlsconfig)
	} else {
		m.conn, err = smtp.Dial(m.serverName + ":" + port)
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
	msg := txtDoc.String()

	c, err := smtp.NewClient(m.conn, m.serverName)
	if err != nil {
		log.Error("Unable to get SMTP Client :", err)
		return err
	}

	if err = c.Auth(m.smtpAuth); err != nil {
		log.Error("Unable to authenticate with SMTP Client :", err)
		return err
	}

	if err = c.Mail(fromStr); err != nil {
		log.Error("Unable to set from for SMTP Client :", err)
		return err
	}

	if err = c.Rcpt(recStr); err != nil {
		log.Error("Unable to set recipient for SMTP Client :",err)
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		log.Error("Unable to set data for SMTP Client :",err)
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		log.Error("Unable to write data for SMTP Client :",err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Error("Unable to close data for SMTP Client :",err)
		return err
	}

	c.Quit()

	log.Info("Email sending (SMTP) ")
	if err != nil {
		log.Error("SMTP error: ", err)
		return err
	}

	return nil
}