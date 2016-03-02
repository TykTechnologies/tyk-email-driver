package emaildriver

type MockEmailBackend struct{}

func (m *MockEmailBackend) Init(conf map[string]string) error {
	return nil
}

func (m *MockEmailBackend) Send(emailMeta EmailMeta, emailData interface{}, textTemplateName TykTemplateName, htmlTemplateName TykTemplateName, OrgId string, Styles string) error {
	return nil
}
