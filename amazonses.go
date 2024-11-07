package emaildriver

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type AmazonSESEmailBackend struct {
	isEnabled       bool
	endpoint        string
	accessKeyId     string
	secretAccessKey string
	sesClient       *ses.Client
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

	s.region, _ = conf["Region"]
	endPoint, _ := conf["Endpoint"]
	if s.region == "" && endPoint == "" {
		return errors.New("No Amazon SES region or endpoint defined, emails will fail")
	}

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

	// Create a new AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s.region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s.accessKeyId, s.secretAccessKey, "")),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	// Create a new SES client
	s.sesClient = ses.NewFromConfig(cfg)

	s.isEnabled = true
	return nil
}

// Send Sends emails using Amazon SES
func (s *AmazonSESEmailBackend) Send(emailMeta EmailMeta, emailData interface{}, textTemplateName TykTemplateName, htmlTemplateName TykTemplateName, OrgId string, Styles string) error {
	if !s.isEnabled {
		log.Warning("Please check your email settings and restart Tyk Dashboard in order for notifications to work")
		return errors.New("amazon SES email driver not initialised correctly")
	}

	var htmlDoc, txtDoc bytes.Buffer

	type superEmailData struct {
		Data   interface{}
		Styles string
	}

	thisData := superEmailData{Data: emailData, Styles: Styles}

	htmlErr := PortalEmailTemplatesHTML.ExecuteTemplate(&htmlDoc, string(htmlTemplateName), thisData)
	if htmlErr != nil {
		log.Error("HTML Template error: ", htmlErr)
		return htmlErr
	}

	txtErr := PortalEmailTemplatesTXT.ExecuteTemplate(&txtDoc, string(textTemplateName), emailData)
	if txtErr != nil {
		log.Error("Text Template error: ", txtErr)
		return txtErr
	}

	from := "\"" + emailMeta.FromName + "\" <" + emailMeta.FromEmail + ">"
	to := "\"" + emailMeta.RecipientName + "\" <" + emailMeta.RecipientEmail + ">"

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String(s.charSet),
					Data:    aws.String(htmlDoc.String()),
				},
				Text: &types.Content{
					Charset: aws.String(s.charSet),
					Data:    aws.String(txtDoc.String()),
				},
			},
			Subject: &types.Content{
				Charset: aws.String(s.charSet),
				Data:    aws.String(emailMeta.Subject),
			},
		},
		Source: aws.String(from),
	}

	_, err := s.sesClient.SendEmail(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
