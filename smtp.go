package emaildriver

import (
	"bytes"
	"errors"
	"net"
	"strconv"

	"github.com/TykTechnologies/logrus"
	"gopkg.in/gomail.v2"
)

type SMTPEmailBackend struct {
	isEnabled bool
	config
}

type config struct {
	host string
	port int
	user string
	pass string
}

// Init receives the configs, validates them and sets on the SMTPEmailBackend struct for use by Send function
func (m *SMTPEmailBackend) Init(conf map[string]string) error {
	log.Info("initializing SMTP email driver")

	if conf == nil {
		return errors.New("SmtpEmailBackend requires a config map")
	}

	user, ok := conf["SMTPUsername"]
	if !ok {
		user = ""
	}

	pass, ok := conf["SMTPPassword"]
	if !ok {
		pass = ""
	}

	if user == "" || pass == "" {
		log.Info("SMTPUsername and/or SMTPPassword not set - smtp driver configured for no-auth")
	}

	host, port, err := net.SplitHostPort(conf["SMTPAddress"])
	if err != nil {
		if e, ok := err.(*net.AddrError); ok && e.Err == "missing port in address" {
			if conf["SMTPPort"] == "" {
				return e
			}
			host = conf["SMTPAddress"]
			port = conf["SMTPPort"]
		} else {
			return err
		}
	}

	log.WithFields(logrus.Fields{
		"address": host,
		"port":    port,
	}).Info("SMTP settings")

	m.host = host
	m.port, _ = strconv.Atoi(port)
	m.user = user
	m.pass = pass
	m.isEnabled = true

	log.Info("SMTP email driver initialized")

	return nil
}

func (m *SMTPEmailBackend) Send(emailMeta EmailMeta, emailData interface{}, textTemplateName TykTemplateName,
	htmlTemplateName TykTemplateName, OrgId string, Styles string) error {

	if !m.isEnabled {
		log.Warning("SMTP driver not initialized.")
		return driverInitializationError
	}

	// Generate strings from template
	var txtDoc bytes.Buffer
	var htmlDoc bytes.Buffer

	type superEmailData struct {
		Data   interface{}
		Styles string
	}

	thisData := superEmailData{Data: emailData}
	thisData.Styles = Styles

	if err := PortalEmailTemplatesTXT.ExecuteTemplate(&txtDoc, string(textTemplateName), emailData); err != nil {
		log.WithError(err).Error("error executing text template")
		return err
	}

	if err := PortalEmailTemplatesHTML.ExecuteTemplate(&htmlDoc, string(htmlTemplateName), thisData); err != nil {
		log.WithError(err).Error("error executing html template")
		return err
	}

	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", emailMeta.FromEmail, emailMeta.FromName)
	msg.SetHeader("To", emailMeta.RecipientEmail)
	msg.SetHeader("Subject", emailMeta.Subject)
	msg.SetBody("text/html", htmlDoc.String())
	msg.AddAlternative("text/plain", txtDoc.String())

	var dialer *gomail.Dialer
	if m.user == "" || m.pass == "" {
		dialer = &gomail.Dialer{
			Host: m.host,
			Port: m.port,
		}
	} else {
		dialer = gomail.NewDialer(m.host, m.port, m.user, m.pass)
	}

	if err := dialer.DialAndSend(msg); err != nil {
		log.WithError(err).Error("error sending mail")
		return err
	}

	log.WithFields(logrus.Fields{
		"host": m.host,
		"from": emailMeta.FromEmail,
		"to":   emailMeta.RecipientEmail,
	}).Debug("email sent")

	return nil
}
