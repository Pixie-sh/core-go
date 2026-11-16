package notifications

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/pixie-sh/di-go"
	"github.com/pixie-sh/errors-go"
)

const (
	smtpAuthModeNone  = "none"
	smtpAuthModePlain = "plain"

	smtpTLSModeNone     = "none"
	smtpTLSModeStartTLS = "starttls"
	smtpTLSModeImplicit = "implicit"
	// smtpTLSModeTLS is a synonym for smtpTLSModeImplicit.
	smtpTLSModeTLS = "tls"
)

// EmailServiceUnsupportedOperationErrorCode flags operations a driver does
// not implement (e.g., SMTP Resend). Distinct from
// errors.ErrorCreatingDependencyErrorCode which signals startup/DI failures.
var EmailServiceUnsupportedOperationErrorCode = errors.NewErrorCode(
	"EmailServiceUnsupportedOperationErrorCode",
	errors.UserInputErrorCode+errors.HTTPInvalidData,
)

var _ di.Configuration = &SmtpServiceConfiguration{}

type SmtpServiceConfiguration struct {
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	FromSender    string `json:"from_sender"`
	AuthMode      string `json:"auth_mode"`
	TlsMode       string `json:"tls_mode"`
	TlsSkipVerify bool   `json:"tls_skip_verify"`
}

func (s *SmtpServiceConfiguration) LookupNode(lookupPath string) (any, error) {
	return di.ConfigurationNodeLookup(s, lookupPath)
}

type SmtpEmailService struct {
	config         SmtpServiceConfiguration
	fromEnvelope   string // bare addr-spec for SMTP MAIL FROM
	fromHeaderLine string // RFC 5322 From header value (display + addr)
}

func NewSmtpEmailService(_ context.Context, config SmtpServiceConfiguration) (*SmtpEmailService, error) {
	if config.AuthMode == "" {
		config.AuthMode = smtpAuthModeNone
	}
	if config.TlsMode == "" {
		config.TlsMode = smtpTLSModeNone
	}

	if config.Host == "" {
		return nil, errors.New("smtp host is required").WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}
	if config.Port == 0 {
		return nil, errors.New("smtp port is required").WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}
	if config.FromSender == "" {
		return nil, errors.New("smtp from_sender is required").WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}
	if config.AuthMode != smtpAuthModeNone && config.AuthMode != smtpAuthModePlain {
		return nil, errors.New("unsupported smtp auth_mode").WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}
	if config.TlsMode != smtpTLSModeNone && config.TlsMode != smtpTLSModeStartTLS && config.TlsMode != smtpTLSModeImplicit && config.TlsMode != smtpTLSModeTLS {
		return nil, errors.New("unsupported smtp tls_mode").WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	parsedFrom, err := mail.ParseAddress(config.FromSender)
	if err != nil {
		return nil, errors.NewWithError(err, "smtp from_sender is not a valid RFC 5322 address").WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	return &SmtpEmailService{
		config:         config,
		fromEnvelope:   parsedFrom.Address,
		fromHeaderLine: parsedFrom.String(),
	}, nil
}

func (s *SmtpEmailService) Send(ctx context.Context, recipient, subject, body string) (string, error) {
	if err := assertNoHeaderInjection("recipient", recipient); err != nil {
		return "", err
	}
	if err := assertNoHeaderInjection("subject", subject); err != nil {
		return "", err
	}

	parsedRecipient, err := mail.ParseAddress(recipient)
	if err != nil {
		return "", errors.NewWithError(err, "smtp recipient is not a valid RFC 5322 address").WithErrorCode(errors.InvalidFormDataCode)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	client, err := s.newClient(ctx)
	if err != nil {
		return "", errors.NewWithError(err, "failed to connect to smtp server")
	}
	defer client.Close()

	if s.config.AuthMode == smtpAuthModePlain {
		if err = client.Auth(smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)); err != nil {
			return "", errors.NewWithError(err, "failed to authenticate with smtp server")
		}
	}

	if err = client.Mail(s.fromEnvelope); err != nil {
		return "", errors.NewWithError(err, "failed to set smtp sender")
	}
	if err = client.Rcpt(parsedRecipient.Address); err != nil {
		return "", errors.NewWithError(err, "failed to set smtp recipient")
	}

	messageID := newSMTPMessageID(s.fromEnvelope)

	writer, err := client.Data()
	if err != nil {
		return "", errors.NewWithError(err, "failed to start smtp data")
	}
	if _, err = writer.Write([]byte(buildSmtpMessage(s.fromHeaderLine, parsedRecipient.String(), subject, body, messageID, time.Now()))); err != nil {
		_ = writer.Close()
		return "", errors.NewWithError(err, "failed to write smtp data")
	}
	if err = writer.Close(); err != nil {
		return "", errors.NewWithError(err, "failed to close smtp data")
	}
	if err = client.Quit(); err != nil {
		return "", errors.NewWithError(err, "failed to quit smtp session")
	}

	return messageID, nil
}

func (s *SmtpEmailService) Resend(_ context.Context, _ string, _ ...string) (string, error) {
	return "", errors.New("resend not supported by smtp driver").WithErrorCode(EmailServiceUnsupportedOperationErrorCode)
}

func (s *SmtpEmailService) newClient(ctx context.Context) (*smtp.Client, error) {
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	tlsConfig := &tls.Config{
		ServerName:         s.config.Host,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: s.config.TlsSkipVerify,
	}

	var conn net.Conn
	var err error
	if s.config.TlsMode == smtpTLSModeImplicit || s.config.TlsMode == smtpTLSModeTLS {
		dialer := tls.Dialer{Config: tlsConfig}
		conn, err = dialer.DialContext(ctx, "tcp", address)
	} else {
		dialer := net.Dialer{}
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return nil, err
	}

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if s.config.TlsMode == smtpTLSModeStartTLS {
		if err = client.StartTLS(tlsConfig); err != nil {
			_ = client.Close()
			return nil, err
		}
	}

	return client, nil
}

func buildSmtpMessage(fromHeader, toHeader, subject, body, messageID string, now time.Time) string {
	headers := []string{
		fmt.Sprintf("From: %s", fromHeader),
		fmt.Sprintf("To: %s", toHeader),
		fmt.Sprintf("Subject: %s", mime.QEncoding.Encode("utf-8", subject)),
		fmt.Sprintf("Date: %s", now.UTC().Format(time.RFC1123Z)),
		fmt.Sprintf("Message-ID: <%s>", messageID),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}

	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body + "\r\n"
}

// assertNoHeaderInjection rejects values containing CR or LF, which would
// allow header smuggling (Bcc:, additional MIME parts, etc.) when the value
// is interpolated into an outbound RFC 5322 header line.
func assertNoHeaderInjection(field, value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return errors.New("smtp %s contains illegal CR/LF characters", field).WithErrorCode(errors.InvalidFormDataCode)
	}
	return nil
}

// newSMTPMessageID returns a Message-ID local part + domain suitable for
// RFC 5322. Uses 16 random bytes + nanosecond timestamp to avoid concurrent
// collisions; falls back to timestamp-only on rand failure.
func newSMTPMessageID(fromEnvelope string) string {
	domain := messageIDDomain(fromEnvelope)
	var random [16]byte
	if _, err := rand.Read(random[:]); err == nil {
		return fmt.Sprintf("%x.%d@%s", random[:], time.Now().UnixNano(), domain)
	}
	return fmt.Sprintf("%d@%s", time.Now().UnixNano(), domain)
}

func messageIDDomain(fromEnvelope string) string {
	if at := strings.LastIndex(fromEnvelope, "@"); at >= 0 && at < len(fromEnvelope)-1 {
		return fromEnvelope[at+1:]
	}
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		return hostname
	}
	return "localhost"
}
