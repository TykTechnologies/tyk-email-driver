package emaildriver

import (
	"testing"
)

func TestSMTPEmailBackend_Init(t *testing.T) {
	testCases := map[string]struct {
		conf        map[string]string
		expectError bool
	}{
		"nil map`": {
			conf:        nil,
			expectError: true,
		},
		"empty map": {
			conf:        map[string]string{},
			expectError: true,
		},
		"junk address": {
			conf: map[string]string{
				"SMTPAddress": "junk",
			},
			expectError: true,
		},
		"valid just host": {
			conf: map[string]string{
				"SMTPAddress": "abc.com:123",
			},
			expectError: false,
		},
		"give port twice (ignore SMTPPort)": {
			conf: map[string]string{
				"SMTPAddress": "abc.com:123",
				"SMTPPort":    "456",
			},
			expectError: false,
		},
		"valid with port": {
			conf: map[string]string{
				"SMTPAddress": "abc.com",
				"SMTPPort":    "587",
			},
			expectError: false,
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			emailBackend := SMTPEmailBackend{}
			err := emailBackend.Init(tc.conf)
			if (tc.expectError == true && err == nil) || (tc.expectError == false && err != nil) {
				t.Fatalf("Did not get an error per expectation: %#v", err)
			}
		})
	}
}
