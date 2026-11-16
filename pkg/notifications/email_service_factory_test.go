package notifications

import (
	"context"
	"testing"

	"github.com/pixie-sh/core-go/pkg/models/serializer"
)

func TestNewEmailService_Factory(t *testing.T) {
	tests := []struct {
		name    string
		config  EmailServiceConfiguration
		wantErr bool
	}{
		{
			name: "valid_mailgun",
			config: emailServiceConfig(t, EmailServiceDriverMailgun, MailgunServiceConfiguration{
				ApiKey:     "test",
				Domain:     "test.com",
				FromSender: "Test <test@test.com>",
			}),
		},
		{
			name: "valid_smtp",
			config: emailServiceConfig(t, EmailServiceDriverSMTP, SmtpServiceConfiguration{
				Host:       "127.0.0.1",
				Port:       1025,
				FromSender: "Test <test@test.com>",
				AuthMode:   smtpAuthModeNone,
				TlsMode:    smtpTLSModeNone,
			}),
		},
		{
			name: "valid_smtp_from_json_node",
			config: EmailServiceConfiguration{
				Driver: EmailServiceDriverSMTP,
				Configuration: map[string]any{
					"host":            "127.0.0.1",
					"port":            float64(1025),
					"from_sender":     "Test <test@test.com>",
					"auth_mode":       smtpAuthModeNone,
					"tls_mode":        smtpTLSModeNone,
					"tls_skip_verify": false,
				},
			},
		},
		{
			name:   "empty_driver_returns_nil",
			config: EmailServiceConfiguration{},
		},
		{
			name: "unknown_driver_error",
			config: EmailServiceConfiguration{
				Driver:        "unknown",
				Configuration: map[string]any{},
			},
			wantErr: true,
		},
		{
			name: "payload_mismatch_error",
			config: EmailServiceConfiguration{
				Driver:        EmailServiceDriverSMTP,
				Configuration: "invalid input",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewEmailService(context.Background(), tt.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewEmailService() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.config.Driver != "" && service == nil {
				t.Fatal("NewEmailService() returned nil service for valid driver")
			}
			if !tt.wantErr && tt.config.Driver == "" && service != nil {
				t.Fatal("NewEmailService() returned non-nil service for empty driver")
			}
		})
	}
}

func emailServiceConfig(t *testing.T, driver string, config any) EmailServiceConfiguration {
	t.Helper()

	payload, err := serializer.StructToMap[map[string]any](config)
	if err != nil {
		t.Fatal(err)
	}

	return EmailServiceConfiguration{
		Driver:        driver,
		Configuration: payload,
	}
}
