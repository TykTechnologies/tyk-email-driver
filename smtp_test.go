package emaildriver

import "testing"

var baseSMTPMap = map[string]string{}

func TestSMTPEmailBackend_Init(t *testing.T) {

	emailBackend := SMTPEmailBackend{}

	if err := emailBackend.Init(nil); err == nil {
		t.Error("expected driver initialization error with nil config")
	}

	if err := emailBackend.Init(baseSMTPMap); err == nil {
		t.Error("expected error with missing SMTPUsername")
	}

	baseSMTPMap["SMTPAddress"] = "junk"
	if err := emailBackend.Init(baseSMTPMap); err == nil {
		t.Error("expected error when unable to get host and port from SMTPAddress")
	}

	baseSMTPMap["SMTPAddress"] = "abc.com"
	if err := emailBackend.Init(baseSMTPMap); err == nil {
		t.Error("expected error with missing port in SMTPAddress")
	}

	baseSMTPMap["SMTPAddress"] = "abc.com:123"
	if err := emailBackend.Init(baseSMTPMap); err != nil {
		t.Error("valid config map Init should return nil")
	}
}
