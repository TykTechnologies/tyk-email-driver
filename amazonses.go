package emaildriver

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

type AmazonSESEmailBackend struct {
	isEnabled       bool
	endpoint        string
	accessKeyId     string
	secretAccessKey string
	sessionData     *session.Session
	charSet         string
	region          string
}

func getRegionFromEndpoint(e string) string {
	r0 := regexp.MustCompile(`email(?:\-smtp)?\.(?P<region>[\w|\-]+?)\.amazonaws\.com`)
	m0 := r0.FindAllStringSubmatch(e, -1)
	if len(m0) > 0 {
		return m0[0][1]
	}
	return ""
} // getRegionFromEndpoint ...

func (s *AmazonSESEmailBackend) Init(conf map[string]string) error {
	var ok bool

	if conf == nil {
		return errors.New("AmazonSESEmailBackend requires a configuration map")
	}

	// we can retrieve region from the Region parameter, or from the Endpoint parameter
	s.region, _ = conf["Region"]
	endPoint, _ := conf["Endpoint"]
	if s.region == "" && endPoint == "" {
		return errors.New("No Amazon SES region or endpoint defined, emails will fail")
	}

	// Endpoint provided, try to get region from there
	if s.region == "" {
		s.region = getRegionFromEndpoint(endPoint)
		if s.region == "" {
			return errors.New("Amazon SES region could not be retrieved from endpoint, emails will fail")
		}
	}

	s.accessKeyId, ok = conf["AccessKeyId"]
	if !ok {
		return errors.New("No Amazon SES access key defined, emails will fail")
	}

	s.secretAccessKey, ok = conf["SecretAccessKey"]
	if !ok {
		return errors.New("No Amazon SES secret access key defined, emails will fail")
	}

	s.charSet, ok = conf["CharSet"]
	if !ok {
		s.charSet = "UTF-8"
	}

	// Create a new session and specify an AWS Region.
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(s.region),
		Credentials: credentials.NewStaticCredentials(s.accessKeyId, s.secretAccessKey, ""),
	})
	if err != nil {
		return errors.New("When creating Amazon SES session data")
	}
	s.sessionData = sess

	s.isEnabled = true
	return nil
} // Init ...

// Send Sends emails using Amazon SES
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

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(to),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(s.charSet),
					Data:    aws.String(htmlDoc.String()),
				},
				Text: &ses.Content{
					Charset: aws.String(s.charSet),
					Data:    aws.String(txtDoc.String()),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(s.charSet),
				Data:    aws.String(emailMeta.Subject),
			},
		},
		Source: aws.String(from),
	}

	// // Create an SES client in the session.
	svc := ses.New(s.sessionData)

	// Attempt to send the email.
	_, err := svc.SendEmail(input)

	if err != nil {
		var errorText string
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				errorText = fmt.Sprintf("%s: %v", ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				errorText = fmt.Sprintf("%s: %v", ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				errorText = fmt.Sprintf("%s: %v", ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				errorText = fmt.Sprintf("%v", aerr.Error())
			}
		} else {
			errorText = fmt.Sprintf("%v", err.Error())
		}
		return errors.New(errorText)
	} // if err != nil ...

	return nil
} // Send
