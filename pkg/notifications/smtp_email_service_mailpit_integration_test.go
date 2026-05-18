//go:build integration
// +build integration

package notifications_test

import (
	"context"
	"encoding/json"
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

// Integration test against a live Mailpit instance.
//
// Expected layout (per local dev port-forwarding):
//   - SMTP: localhost:4444
//   - HTTP API: localhost:8025
//
// Run with:
//
//	go test -tags=integration ./pkg/notifications/... \
//	    -run TestSmtpEmailService_Mailpit -count=1 -v
//
// To override ports for CI/other setups:
//
//	MAILPIT_SMTP_HOST=127.0.0.1 MAILPIT_SMTP_PORT=1025 \
//	MAILPIT_HTTP_BASE=http://127.0.0.1:8025 \
//	    go test -tags=integration ...
const (
	defaultMailpitSMTPHost = "localhost"
	defaultMailpitSMTPPort = 4444
	defaultMailpitHTTPBase = "http://localhost:8025"
)

func TestSmtpEmailService_Mailpit(t *testing.T) {
	smtpHost := envOr("MAILPIT_SMTP_HOST", defaultMailpitSMTPHost)
	smtpPort := envOrInt("MAILPIT_SMTP_PORT", defaultMailpitSMTPPort)
	httpBase := strings.TrimRight(envOr("MAILPIT_HTTP_BASE", defaultMailpitHTTPBase), "/")

	purgeMailpit(t, httpBase)

	service, err := notifications.NewSmtpEmailService(
		context.Background(),
		notifications.SmtpServiceConfiguration{
			Host:       smtpHost,
			Port:       smtpPort,
			FromSender: `"Integration Sender" <integration@example.test>`,
			AuthMode:   "none",
			TlsMode:    "none",
		},
	)
	if err != nil {
		t.Fatalf("NewSmtpEmailService() error = %v", err)
	}

	// Use a unique subject so we can locate this exact send in Mailpit even
	// if other tests share the inbox.
	subject := fmt.Sprintf("Integration smoke %d", time.Now().UnixNano())
	recipient := "qa@example.test"
	body := `<html><body><p>Hello from the <b>core-go</b> integration test.</p></body></html>`

	messageID, err := service.Send(context.Background(), recipient, subject, body)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if messageID == "" {
		t.Fatal("Send() returned empty message id")
	}
	t.Logf("delivered Message-ID: %s", messageID)

	msg := waitForMailpitMessage(t, httpBase, subject, 5*time.Second)

	if !strings.EqualFold(msg.From.Address, "integration@example.test") {
		t.Errorf("From.Address = %q, want integration@example.test", msg.From.Address)
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

// --- Mailpit HTTP helpers ----------------------------------------------------

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
	Text    string           `json:"Text"`
}

type mailpitListResponse struct {
	Messages []mailpitMessage `json:"messages"`
}

// purgeMailpit deletes all stored messages so the test starts from a clean
// inbox. Failure to purge is fatal — without it, subject-based lookup may
// match a stale message from a prior run.
func purgeMailpit(t *testing.T, base string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, base+"/api/v1/messages", nil)
	if err != nil {
		t.Fatalf("purge request build: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("purge mailpit: %v (is Mailpit reachable at %s?)", err, base)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		t.Fatalf("purge mailpit: status %d", resp.StatusCode)
	}
}

// waitForMailpitMessage polls /api/v1/search until a message matching the
// unique subject appears, or the timeout elapses.
func waitForMailpitMessage(t *testing.T, base, subject string, timeout time.Duration) mailpitMessage {
	t.Helper()
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("%s/api/v1/search?query=%s", base, urlQueryEscape(`subject:"`+subject+`"`))

	for {
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("mailpit search: %v", err)
		}
		var list mailpitListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			resp.Body.Close()
			t.Fatalf("mailpit decode: %v", err)
		}
		resp.Body.Close()

		if len(list.Messages) == 1 {
			return fetchMailpitMessage(t, base, list.Messages[0].ID)
		}
		if len(list.Messages) > 1 {
			t.Fatalf("expected 1 message for subject %q, got %d", subject, len(list.Messages))
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for Mailpit to receive subject %q", subject)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func fetchMailpitMessage(t *testing.T, base, id string) mailpitMessage {
	t.Helper()
	resp, err := http.Get(base + "/api/v1/message/" + id)
	if err != nil {
		t.Fatalf("mailpit fetch %s: %v", id, err)
	}
	defer resp.Body.Close()
	var msg mailpitMessage
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatalf("mailpit fetch decode: %v", err)
	}
	return msg
}

// fetchMailpitHeaders returns the parsed RFC 822 headers for a message.
// Mailpit exposes them as map[string][]string at /api/v1/message/{id}/headers.
func fetchMailpitHeaders(t *testing.T, base, id string) map[string][]string {
	t.Helper()
	resp, err := http.Get(base + "/api/v1/message/" + id + "/headers")
	if err != nil {
		t.Fatalf("mailpit headers %s: %v", id, err)
	}
	defer resp.Body.Close()
	headers := map[string][]string{}
	if err := json.NewDecoder(resp.Body).Decode(&headers); err != nil {
		t.Fatalf("mailpit headers decode: %v", err)
	}
	return headers
}

func headerFirst(h map[string][]string, key string) string {
	if v, ok := h[key]; ok && len(v) > 0 {
		return v[0]
	}
	// Mailpit may canonicalize differently; do a case-insensitive fallback.
	for k, v := range h {
		if strings.EqualFold(k, key) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func urlQueryEscape(s string) string { return url.QueryEscape(s) }
