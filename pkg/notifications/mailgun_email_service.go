package notifications

import (
	"context"
	"time"

	"github.com/pixie-sh/di-go"
	"github.com/pixie-sh/errors-go"

	"github.com/mailgun/mailgun-go/v4"
)

var _ di.Configuration = &MailgunServiceConfiguration{}

type MailgunServiceConfiguration struct {
	ApiKey     string `json:"api_key"`
	Domain     string `json:"domain"`
	FromSender string `json:"from_sender"`
}

// LookupNode implements di.Configuration.
func (m *MailgunServiceConfiguration) LookupNode(lookupPath string) (any, error) {
	return di.ConfigurationNodeLookup(m, lookupPath)
}

// MailgunService implements EmailService using Mailgun
type MailgunService struct {
	config        MailgunServiceConfiguration
	mailgunClient *mailgun.MailgunImpl
}

// NewMailgunService creates a new MailgunService
func NewMailgunService(_ context.Context, config MailgunServiceConfiguration) (*MailgunService, error) {
	mailgunInstance := mailgun.NewMailgun(config.Domain, config.ApiKey)
	mailgunInstance.SetAPIBase(mailgun.APIBaseEU)

	return &MailgunService{
		config:        config,
		mailgunClient: mailgunInstance,
	}, nil
}

// Send sends an email using Mailgun
func (m *MailgunService) Send(ctx context.Context, recipient, subject, body string) (string, error) {
	message := m.mailgunClient.NewMessage(m.config.FromSender, subject, "", recipient)
	message.SetHtml(body)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, id, err := m.mailgunClient.Send(ctx, message)
	if err != nil {
		return "", errors.NewWithError(err, "failed to send email")
	}

	return id, nil
}

// Resend sends an email using Mailgun
func (m *MailgunService) Resend(ctx context.Context, id string, recipients ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, id, err := m.mailgunClient.ReSend(ctx, id, recipients...)
	if err != nil {
		return "", errors.NewWithError(err, "failed to resend email")
	}

	return id, nil
}
