package notifications

import (
	"context"

	"github.com/pixie-sh/di-go"
	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/models/serializer"
)

const (
	EmailServiceDriverMailgun = "mailgun"
	EmailServiceDriverSMTP    = "smtp"
)

var _ di.Configuration = &EmailServiceConfiguration{}

type EmailServiceConfiguration struct {
	Driver        string `json:"driver"`
	Configuration any    `json:"configuration"`
}

func (e *EmailServiceConfiguration) LookupNode(lookupPath string) (any, error) {
	return di.ConfigurationNodeLookup(e, lookupPath)
}

// IsEmpty reports whether the configuration has no driver selected. Callers
// (e.g., NotificationsBusinessLayer) use this to tolerate deployments that
// ship without an email_service section.
func (e EmailServiceConfiguration) IsEmpty() bool {
	return e.Driver == ""
}

// NewEmailService builds the concrete EmailService for the configured driver.
// An empty Driver returns (nil, nil) so callers can treat the service as
// optional, matching the legacy behavior where a missing email config was
// logged as a warning rather than failing startup.
func NewEmailService(ctx context.Context, config EmailServiceConfiguration) (EmailService, error) {
	switch config.Driver {
	case "":
		return nil, nil
	case EmailServiceDriverMailgun:
		var mailgunConfig MailgunServiceConfiguration
		if err := serializer.ToStruct(config.Configuration, &mailgunConfig); err != nil {
			return nil, errors.NewWithError(err, "failed to parse mailgun email service configuration")
		}

		return NewMailgunService(ctx, mailgunConfig)
	case EmailServiceDriverSMTP:
		var smtpConfig SmtpServiceConfiguration
		if err := serializer.ToStruct(config.Configuration, &smtpConfig); err != nil {
			return nil, errors.NewWithError(err, "failed to parse smtp email service configuration")
		}

		return NewSmtpEmailService(ctx, smtpConfig)
	default:
		return nil, errors.New("unsupported email service driver: %s", config.Driver).WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}
}
