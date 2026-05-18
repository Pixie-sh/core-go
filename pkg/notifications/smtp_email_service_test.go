package notifications

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/pixie-sh/errors-go"
)

func TestSmtpEmailService_Send(t *testing.T) {
	tests := []struct {
		name          string
		authMode      string
		tlsMode       string
		implicitTLS   bool
		startTLS      bool
		expectAuth    bool
		expectHeaders []string
	}{
		{
			name:        "no_auth_no_tls",
			authMode:    smtpAuthModeNone,
			tlsMode:     smtpTLSModeNone,
			implicitTLS: false,
			startTLS:    false,
		},
		{
			name:        "plain_auth_no_tls",
			authMode:    smtpAuthModePlain,
			tlsMode:     smtpTLSModeNone,
			implicitTLS: false,
			startTLS:    false,
			expectAuth:  true,
		},
		{
			name:        "starttls_with_plain",
			authMode:    smtpAuthModePlain,
			tlsMode:     smtpTLSModeStartTLS,
			implicitTLS: false,
			startTLS:    true,
			expectAuth:  true,
		},
		{
			name:        "implicit_tls_with_plain",
			authMode:    smtpAuthModePlain,
			tlsMode:     smtpTLSModeImplicit,
			implicitTLS: true,
			startTLS:    false,
			expectAuth:  true,
		},
		{
			name:        "mime_html_assertion",
			authMode:    smtpAuthModeNone,
			tlsMode:     smtpTLSModeNone,
			implicitTLS: false,
			startTLS:    false,
			expectHeaders: []string{
				`From: "Test" <test@test.com>`,
				"To: <user@test.com>",
				"Subject: Subject",
				"MIME-Version: 1.0",
				"Content-Type: text/html; charset=UTF-8",
				"<p>Hello</p>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newMockSMTPServer(t, tt.implicitTLS, tt.startTLS)
			defer server.Close()

			service, err := NewSmtpEmailService(context.Background(), SmtpServiceConfiguration{
				Host:          server.Host,
				Port:          server.Port,
				Username:      "user",
				Password:      "pass",
				FromSender:    "Test <test@test.com>",
				AuthMode:      tt.authMode,
				TlsMode:       tt.tlsMode,
				TlsSkipVerify: true,
			})
			if err != nil {
				t.Fatal(err)
			}

			id, err := service.Send(context.Background(), "user@test.com", "Subject", "<p>Hello</p>")
			if err != nil {
				t.Fatalf("Send() error = %v", err)
			}
			if !strings.Contains(id, "@test.com") {
				t.Fatalf("Send() id = %q, want containing @test.com", id)
			}
			if !strings.Contains(server.Data, "Message-ID: <"+id+">") {
				t.Fatalf("SMTP data missing Message-ID header for id %q in:\n%s", id, server.Data)
			}
			if !strings.Contains(server.Data, "Date: ") {
				t.Fatalf("SMTP data missing Date header in:\n%s", server.Data)
			}
			if server.AuthSeen != tt.expectAuth {
				t.Fatalf("AuthSeen = %v, want %v", server.AuthSeen, tt.expectAuth)
			}
			for _, header := range tt.expectHeaders {
				if !strings.Contains(server.Data, header) {
					t.Fatalf("SMTP data missing %q in:\n%s", header, server.Data)
				}
			}
		})
	}
}

func TestSmtpEmailService_Resend_Unsupported(t *testing.T) {
	service := &SmtpEmailService{}
	_, err := service.Resend(context.Background(), "id", "user@test.com")
	if err == nil {
		t.Fatal("Resend() error = nil, want unsupported")
	}
	e, ok := err.(errors.E)
	if !ok {
		t.Fatalf("Resend() error type = %T, want errors.E", err)
	}
	if e.Code != EmailServiceUnsupportedOperationErrorCode {
		t.Fatalf("Resend() error code = %v, want EmailServiceUnsupportedOperationErrorCode", e.Code)
	}
}

func TestSmtpEmailService_Send_RejectsHeaderInjection(t *testing.T) {
	service, err := NewSmtpEmailService(context.Background(), SmtpServiceConfiguration{
		Host:       "127.0.0.1",
		Port:       2525,
		FromSender: "Test <test@test.com>",
		AuthMode:   smtpAuthModeNone,
		TlsMode:    smtpTLSModeNone,
	})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name      string
		recipient string
		subject   string
	}{
		{name: "recipient_lf", recipient: "user@test.com\nBcc: evil@test.com", subject: "Subject"},
		{name: "recipient_cr", recipient: "user@test.com\rBcc: evil@test.com", subject: "Subject"},
		{name: "subject_lf", recipient: "user@test.com", subject: "Subject\nX-Injected: yes"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := service.Send(context.Background(), tc.recipient, tc.subject, "body"); err == nil {
				t.Fatal("expected header-injection rejection, got nil")
			}
		})
	}
}

type mockSMTPServer struct {
	Host     string
	Port     int
	Data     string
	AuthSeen bool
	listener net.Listener
	done     chan struct{}
	t        *testing.T
	cert     tls.Certificate
	startTLS bool
}

func newMockSMTPServer(t *testing.T, implicitTLS bool, startTLS bool) *mockSMTPServer {
	t.Helper()

	cert := testTLSCertificate(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if implicitTLS {
		listener = tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12})
	}

	addr := listener.Addr().(*net.TCPAddr)
	server := &mockSMTPServer{
		Host:     "127.0.0.1",
		Port:     addr.Port,
		listener: listener,
		done:     make(chan struct{}),
		t:        t,
		cert:     cert,
		startTLS: startTLS,
	}

	go server.serve()
	return server
}

func (s *mockSMTPServer) Close() {
	_ = s.listener.Close()
	<-s.done
}

func (s *mockSMTPServer) serve() {
	defer close(s.done)

	conn, err := s.listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	writeSMTPLine(writer, "220 localhost ESMTP")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				s.t.Errorf("read smtp command: %v", err)
			}
			return
		}
		command := strings.TrimSpace(line)

		upperCommand := strings.ToUpper(command)
		switch {
		case strings.HasPrefix(upperCommand, "EHLO") || strings.HasPrefix(upperCommand, "HELO"):
			writeSMTPLine(writer, "250-localhost")
			if s.startTLS {
				writeSMTPLine(writer, "250-STARTTLS")
			}
			writeSMTPLine(writer, "250 AUTH PLAIN")
		case upperCommand == "STARTTLS":
			writeSMTPLine(writer, "220 Ready to start TLS")
			conn = tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{s.cert}, MinVersion: tls.VersionTLS12})
			reader = bufio.NewReader(conn)
			writer = bufio.NewWriter(conn)
		case strings.HasPrefix(upperCommand, "AUTH PLAIN"):
			s.AuthSeen = true
			writeSMTPLine(writer, "235 Authentication successful")
		case strings.HasPrefix(upperCommand, "MAIL FROM"):
			writeSMTPLine(writer, "250 OK")
		case strings.HasPrefix(upperCommand, "RCPT TO"):
			writeSMTPLine(writer, "250 OK")
		case upperCommand == "DATA":
			writeSMTPLine(writer, "354 End data with <CR><LF>.<CR><LF>")
			s.Data = readSMTPData(s.t, reader)
			writeSMTPLine(writer, "250 OK")
		case upperCommand == "QUIT":
			writeSMTPLine(writer, "221 Bye")
			return
		default:
			writeSMTPLine(writer, "250 OK")
		}
	}
}

func writeSMTPLine(writer *bufio.Writer, line string) {
	_, _ = writer.WriteString(line + "\r\n")
	_ = writer.Flush()
}

func readSMTPData(t *testing.T, reader *bufio.Reader) string {
	t.Helper()

	var data strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(line) == "." {
			return data.String()
		}
		data.WriteString(line)
	}
}

func testTLSCertificate(t *testing.T) tls.Certificate {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}

	return cert
}
