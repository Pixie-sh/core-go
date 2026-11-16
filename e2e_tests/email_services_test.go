//go:build e2e

// Package e2e_tests holds end-to-end tests that depend on running infrastructure
// (containers, network listeners, real protocols). They are not part of the
// default `go test ./...` run.
//
// Bring up the dependencies first:
//
//	docker compose -f e2e_tests/dockerfiles/docker-compose.yaml up -d
//
// Run the tests:
//
//	go test -tags=e2e -count=1 -v ./e2e_tests/...
//
// Tear down when done:
//
//	docker compose -f e2e_tests/dockerfiles/docker-compose.yaml down
//
// Override Mailpit endpoints (defaults shown):
//
//	MAILPIT_SMTP_HOST=localhost \
//	MAILPIT_SMTP_PORT=1025 \
//	MAILPIT_HTTP_BASE=http://localhost:8025 \
//	    go test -tags=e2e ./e2e_tests/...
package e2e_tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pixie-sh/core-go/pkg/notifications"
)

const (
	defaultMailpitSMTPHost = "localhost"
	defaultMailpitSMTPPort = 1025
	defaultMailpitHTTPBase = "http://localhost:8025"
)

// TestEmailService_SMTPDriver_E2E exercises the FULL factory path:
//   notifications.NewEmailService(ctx, EmailServiceConfiguration{Driver: "smtp", ...})
// then asserts delivery via the Mailpit HTTP API. This is the contract a
// downstream microservice would invoke after loading its JSON configuration.
func TestEmailService_SMTPDriver_E2E(t *testing.T) {
	smtpHost := envOr("MAILPIT_SMTP_HOST", defaultMailpitSMTPHost)
	smtpPort := envOrInt("MAILPIT_SMTP_PORT", defaultMailpitSMTPPort)
	httpBase := strings.TrimRight(envOr("MAILPIT_HTTP_BASE", defaultMailpitHTTPBase), "/")

	requireMailpitReachable(t, httpBase)
	purgeMailpit(t, httpBase)

	cfg := notifications.EmailServiceConfiguration{
		Driver: notifications.EmailServiceDriverSMTP,
		Configuration: map[string]any{
			"host":            smtpHost,
			"port":            smtpPort,
			"username":        "",
			"password":        "",
			"from_sender":     `"E2E Sender" <e2e@example.test>`,
			"auth_mode":       "none",
			"tls_mode":        "none",
			"tls_skip_verify": false,
		},
	}

	ctx := context.Background()
	service, err := notifications.NewEmailService(ctx, cfg)
	if err != nil {
		t.Fatalf("NewEmailService() error = %v", err)
	}
	if service == nil {
		t.Fatal("NewEmailService() returned nil for driver=smtp")
	}

	subject := fmt.Sprintf("E2E SMTP %d", time.Now().UnixNano())
	recipient := "qa@example.test"
	body := `<html><body><p>Hello from the <b>core-go</b> e2e SMTP test.</p></body></html>`

	messageID, err := service.Send(ctx, recipient, subject, body)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if messageID == "" {
		t.Fatal("Send() returned empty message id")
	}
	t.Logf("delivered Message-ID: %s", messageID)

	msg := waitForMailpitMessage(t, httpBase, subject, 5*time.Second)

	if !strings.EqualFold(msg.From.Address, "e2e@example.test") {
		t.Errorf("From.Address = %q, want e2e@example.test", msg.From.Address)
	}
	if len(msg.To) != 1 || !strings.EqualFold(msg.To[0].Address, recipient) {
		t.Errorf("To = %+v, want single recipient %s", msg.To, recipient)
	}
	if len(msg.Bcc) != 0 {
		t.Errorf("Bcc = %+v, want empty (regression guard for header injection)", msg.Bcc)
	}
	if msg.Subject != subject {
		t.Errorf("Subject = %q, want %q", msg.Subject, subject)
	}
	if !strings.Contains(msg.HTML, "core-go") {
		t.Errorf("HTML body missing expected content: %q", msg.HTML)
	}

	headers := fetchMailpitHeaders(t, httpBase, msg.ID)
	if got := headerFirst(headers, "Message-ID"); got != "<"+messageID+">" {
		t.Errorf("Message-ID header = %q, want %q", got, "<"+messageID+">")
	}
	if headerFirst(headers, "Date") == "" {
		t.Error("Date header missing")
	}
	if !strings.HasPrefix(headerFirst(headers, "Content-Type"), "text/html") {
		t.Errorf("Content-Type = %q, want text/html*", headerFirst(headers, "Content-Type"))
	}
}

// TestEmailService_EmptyDriver_E2E confirms the factory's empty-driver contract:
// no SMTP traffic, no error — callers can treat the service as optional.
func TestEmailService_EmptyDriver_E2E(t *testing.T) {
	service, err := notifications.NewEmailService(context.Background(), notifications.EmailServiceConfiguration{
		Driver: "",
	})
	if err != nil {
		t.Fatalf("NewEmailService(driver=\"\") error = %v, want nil", err)
	}
	if service != nil {
		t.Fatalf("NewEmailService(driver=\"\") returned non-nil service %T, want nil", service)
	}
}

// TestEmailService_UnknownDriver_E2E confirms the factory rejects bogus drivers.
func TestEmailService_UnknownDriver_E2E(t *testing.T) {
	service, err := notifications.NewEmailService(context.Background(), notifications.EmailServiceConfiguration{
		Driver: "carrier-pigeon",
	})
	if err == nil {
		t.Fatal("NewEmailService(driver=carrier-pigeon) returned nil error, want error")
	}
	if service != nil {
		t.Fatalf("NewEmailService(driver=carrier-pigeon) returned non-nil service %T, want nil", service)
	}
}

// ──────────────────────────────────────────────────────────────────────────
// Mailpit HTTP helpers
// ──────────────────────────────────────────────────────────────────────────

type mailpitAddress struct {
	Name    string `json:"Name"`
	Address string `json:"Address"`
}

type mailpitMessage struct {
	ID      string           `json:"ID"`
	From    mailpitAddress   `json:"From"`
	To      []mailpitAddress `json:"To"`
	Bcc     []mailpitAddress `json:"Bcc"`
	Subject string           `json:"Subject"`
	HTML    string           `json:"HTML"`
}

type mailpitListItem struct {
	ID      string `json:"ID"`
	Subject string `json:"Subject"`
}

type mailpitListResponse struct {
	Messages []mailpitListItem `json:"messages"`
}

func requireMailpitReachable(t *testing.T, httpBase string) {
	t.Helper()
	resp, err := http.Get(httpBase + "/api/v1/info")
	if err != nil {
		t.Fatalf("mailpit unreachable at %s/api/v1/info: %v\n\nStart it with: docker compose -f e2e_tests/dockerfiles/docker-compose.yaml up -d", httpBase, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("mailpit /api/v1/info returned HTTP %d", resp.StatusCode)
	}
}

func purgeMailpit(t *testing.T, httpBase string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, httpBase+"/api/v1/messages", nil)
	if err != nil {
		t.Fatalf("build purge request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("purge mailpit: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("purge mailpit returned HTTP %d", resp.StatusCode)
	}
}

func waitForMailpitMessage(t *testing.T, httpBase, subject string, timeout time.Duration) mailpitMessage {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		msg, ok := findMailpitMessage(t, httpBase, subject)
		if ok {
			return msg
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for mailpit message with subject %q", subject)
	return mailpitMessage{}
}

func findMailpitMessage(t *testing.T, httpBase, subject string) (mailpitMessage, bool) {
	t.Helper()
	listURL := fmt.Sprintf("%s/api/v1/search?query=%s", httpBase, url.QueryEscape("subject:\""+subject+"\""))
	resp, err := http.Get(listURL)
	if err != nil {
		t.Fatalf("mailpit search: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("mailpit search returned HTTP %d", resp.StatusCode)
	}

	var list mailpitListResponse
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode mailpit search response: %v", err)
	}
	for _, item := range list.Messages {
		if item.Subject == subject {
			return fetchMailpitMessage(t, httpBase, item.ID), true
		}
	}
	return mailpitMessage{}, false
}

func fetchMailpitMessage(t *testing.T, httpBase, id string) mailpitMessage {
	t.Helper()
	resp, err := http.Get(httpBase + "/api/v1/message/" + id)
	if err != nil {
		t.Fatalf("fetch mailpit message: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("fetch mailpit message returned HTTP %d", resp.StatusCode)
	}
	var msg mailpitMessage
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatalf("decode mailpit message: %v", err)
	}
	return msg
}

func fetchMailpitHeaders(t *testing.T, httpBase, id string) map[string][]string {
	t.Helper()
	resp, err := http.Get(httpBase + "/api/v1/message/" + id + "/headers")
	if err != nil {
		t.Fatalf("fetch mailpit headers: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("fetch mailpit headers returned HTTP %d", resp.StatusCode)
	}
	var headers map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&headers); err != nil {
		t.Fatalf("decode mailpit headers: %v", err)
	}
	return headers
}

func headerFirst(headers map[string][]string, name string) string {
	if vs, ok := headers[name]; ok && len(vs) > 0 {
		return vs[0]
	}
	// Mailpit normalizes header keys to MIME canonical form (e.g. "Message-Id"),
	// which doesn't match Go's textproto canonical form for "Message-ID".
	for k, vs := range headers {
		if strings.EqualFold(k, name) && len(vs) > 0 {
			return vs[0]
		}
	}
	return ""
}

// ──────────────────────────────────────────────────────────────────────────
// Env helpers
// ──────────────────────────────────────────────────────────────────────────

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			panic(errors.New(key + " must be an integer: " + err.Error()))
		}
		return n
	}
	return fallback
}
