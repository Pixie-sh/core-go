package kafka

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

// generateTestCertAndKey generates a self-signed certificate and key for testing.
// Returns base64-encoded PEM certificate and key.
func generateTestCertAndKey() (certBase64, keyBase64 string, err error) {
	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Encode private key to PEM
	keyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	// Base64 encode
	certBase64 = base64.StdEncoding.EncodeToString(certPEM)
	keyBase64 = base64.StdEncoding.EncodeToString(keyPEM)

	return certBase64, keyBase64, nil
}

func TestDecodeBase64_ValidInput(t *testing.T) {
	original := []byte("hello world")
	encoded := base64.StdEncoding.EncodeToString(original)

	decoded, err := decodeBase64(encoded)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if string(decoded) != string(original) {
		t.Errorf("expected %q, got %q", string(original), string(decoded))
	}
}

func TestDecodeBase64_InvalidInput(t *testing.T) {
	invalidBase64 := "not-valid-base64!!!"

	_, err := decodeBase64(invalidBase64)
	if err == nil {
		t.Fatal("expected error for invalid base64, got nil")
	}
}

func TestDecodeBase64_EmptyInput(t *testing.T) {
	decoded, err := decodeBase64("")
	if err != nil {
		t.Fatalf("expected no error for empty input, got: %v", err)
	}

	if len(decoded) != 0 {
		t.Errorf("expected empty result, got %d bytes", len(decoded))
	}
}

func TestTLSConfig_Base64Fields(t *testing.T) {
	cfg := TLSConfig{
		Enabled:            true,
		InsecureSkipVerify: true,
		CertBase64:         "dGVzdC1jZXJ0",
		KeyBase64:          "dGVzdC1rZXk=",
		CABase64:           "dGVzdC1jYQ==",
	}

	if cfg.CertBase64 != "dGVzdC1jZXJ0" {
		t.Errorf("CertBase64 field not set correctly")
	}
	if cfg.KeyBase64 != "dGVzdC1rZXk=" {
		t.Errorf("KeyBase64 field not set correctly")
	}
	if cfg.CABase64 != "dGVzdC1jYQ==" {
		t.Errorf("CABase64 field not set correctly")
	}
}

func TestBuildKgoOpts_Base64TLSCertificates(t *testing.T) {
	certBase64, keyBase64, err := generateTestCertAndKey()
	if err != nil {
		t.Fatalf("failed to generate test certificates: %v", err)
	}

	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
			CertBase64:         certBase64,
			KeyBase64:          keyBase64,
		},
	}

	opts, err := buildKgoOpts(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// We should have at least the seed brokers, logger, client ID, and dialer options
	if len(opts) < 3 {
		t.Errorf("expected at least 3 options, got %d", len(opts))
	}
}

func TestBuildKgoOpts_Base64CACertificate(t *testing.T) {
	certBase64, _, err := generateTestCertAndKey()
	if err != nil {
		t.Fatalf("failed to generate test CA certificate: %v", err)
	}

	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: false,
			CABase64:           certBase64,
		},
	}

	opts, err := buildKgoOpts(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(opts) < 3 {
		t.Errorf("expected at least 3 options, got %d", len(opts))
	}
}

func TestBuildKgoOpts_InvalidBase64CertReturnsError(t *testing.T) {
	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
			CertBase64:         "invalid-base64!!!",
			KeyBase64:          "also-invalid!!!",
		},
	}

	_, err := buildKgoOpts(cfg)
	if err == nil {
		t.Fatal("expected error for invalid base64 cert, got nil")
	}

	if !containsString(err.Error(), "TLS certificate base64") {
		t.Errorf("expected error about TLS certificate base64, got: %v", err)
	}
}

func TestBuildKgoOpts_InvalidBase64KeyReturnsError(t *testing.T) {
	certBase64, _, err := generateTestCertAndKey()
	if err != nil {
		t.Fatalf("failed to generate test certificates: %v", err)
	}

	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
			CertBase64:         certBase64,
			KeyBase64:          "invalid-base64!!!",
		},
	}

	_, err = buildKgoOpts(cfg)
	if err == nil {
		t.Fatal("expected error for invalid base64 key, got nil")
	}

	if !containsString(err.Error(), "TLS key base64") {
		t.Errorf("expected error about TLS key base64, got: %v", err)
	}
}

func TestBuildKgoOpts_InvalidCertKeyPairReturnsError(t *testing.T) {
	// Generate two separate cert/key pairs - cross them to create a mismatch
	certBase64, _, err := generateTestCertAndKey()
	if err != nil {
		t.Fatalf("failed to generate test certificates: %v", err)
	}
	_, keyBase64, err := generateTestCertAndKey()
	if err != nil {
		t.Fatalf("failed to generate second test certificates: %v", err)
	}

	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
			CertBase64:         certBase64,
			KeyBase64:          keyBase64, // mismatched key
		},
	}

	_, err = buildKgoOpts(cfg)
	if err == nil {
		t.Fatal("expected error for mismatched cert/key pair, got nil")
	}

	if !containsString(err.Error(), "certificate/key pair") {
		t.Errorf("expected error about certificate/key pair, got: %v", err)
	}
}

func TestBuildKgoOpts_InvalidCertFileReturnsError(t *testing.T) {
	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
			CertFile:           "/nonexistent/cert.pem",
			KeyFile:            "/nonexistent/key.pem",
		},
	}

	_, err := buildKgoOpts(cfg)
	if err == nil {
		t.Fatal("expected error for nonexistent cert files, got nil")
	}

	if !containsString(err.Error(), "certificate/key files") {
		t.Errorf("expected error about certificate/key files, got: %v", err)
	}
}

func TestBuildKgoOpts_InvalidCABase64ReturnsError(t *testing.T) {
	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: false,
			CABase64:           "invalid-base64!!!",
		},
	}

	_, err := buildKgoOpts(cfg)
	if err == nil {
		t.Fatal("expected error for invalid CA base64, got nil")
	}

	if !containsString(err.Error(), "CA certificate base64") {
		t.Errorf("expected error about CA certificate base64, got: %v", err)
	}
}

func TestBuildKgoOpts_InvalidCAFileReturnsError(t *testing.T) {
	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: false,
			CAFile:             "/nonexistent/ca.pem",
		},
	}

	_, err := buildKgoOpts(cfg)
	if err == nil {
		t.Fatal("expected error for nonexistent CA file, got nil")
	}

	if !containsString(err.Error(), "CA certificate file") {
		t.Errorf("expected error about CA certificate file, got: %v", err)
	}
}

func TestBuildKgoOpts_Base64TakesPrecedenceOverFiles(t *testing.T) {
	certBase64, keyBase64, err := generateTestCertAndKey()
	if err != nil {
		t.Fatalf("failed to generate test certificates: %v", err)
	}

	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
			// Both file and base64 provided - base64 should take precedence
			CertFile:   "/nonexistent/cert.pem",
			KeyFile:    "/nonexistent/key.pem",
			CertBase64: certBase64,
			KeyBase64:  keyBase64,
		},
	}

	opts, err := buildKgoOpts(cfg)
	if err != nil {
		t.Fatalf("expected no error (base64 should be used, files ignored), got: %v", err)
	}

	if len(opts) < 3 {
		t.Errorf("expected at least 3 options, got %d", len(opts))
	}
}

func TestBuildKgoOpts_EmptyBase64FallsBackToFiles(t *testing.T) {
	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
			CertFile:           "/nonexistent/cert.pem",
			KeyFile:            "/nonexistent/key.pem",
			CertBase64:         "",
			KeyBase64:          "",
		},
	}

	// Now returns error for nonexistent files instead of silently ignoring
	_, err := buildKgoOpts(cfg)
	if err == nil {
		t.Fatal("expected error for nonexistent cert files, got nil")
	}
}

func TestBuildKgoOpts_TLSDisabled(t *testing.T) {
	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:    false,
			CertBase64: "some-cert",
			KeyBase64:  "some-key",
		},
	}

	opts, err := buildKgoOpts(cfg)
	if err != nil {
		t.Fatalf("expected no error when TLS disabled, got: %v", err)
	}

	// Without TLS, should have fewer options (no dialer)
	// Just brokers, logger, and client ID = 3
	if len(opts) != 3 {
		t.Errorf("expected 3 options when TLS disabled, got %d", len(opts))
	}
}

func TestBuildKgoOpts_NilTLSConfig(t *testing.T) {
	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS:      nil,
	}

	opts, err := buildKgoOpts(cfg)
	if err != nil {
		t.Fatalf("expected no error when TLS nil, got: %v", err)
	}

	if len(opts) != 3 {
		t.Errorf("expected 3 options when TLS nil, got %d", len(opts))
	}
}

func TestBuildKgoOpts_NoCertConfigWithTLSEnabled(t *testing.T) {
	cfg := &ClientConfiguration{
		Brokers:  []string{"localhost:9092"},
		ClientID: "test-client",
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: true,
			// No cert/key configured - valid for server-only TLS
		},
	}

	opts, err := buildKgoOpts(cfg)
	if err != nil {
		t.Fatalf("expected no error when TLS enabled without client certs, got: %v", err)
	}

	// Should have brokers, logger, client ID, and dialer = 4
	if len(opts) < 4 {
		t.Errorf("expected at least 4 options with TLS dialer, got %d", len(opts))
	}
}

func TestClient_GetTopics_NilClient(t *testing.T) {
	client := &Client{
		kgoClient: nil,
		cfg:       nil,
	}

	ctx := context.Background()
	_, err := client.GetTopics(ctx)
	if err == nil {
		t.Fatal("expected error when kgoClient is nil")
	}

	expectedMsg := "kafka client is nil"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestClient_Ping_NilClient(t *testing.T) {
	client := &Client{
		kgoClient: nil,
		cfg:       nil,
	}

	ctx := context.Background()
	err := client.Ping(ctx)
	if err == nil {
		t.Fatal("expected error when kgoClient is nil")
	}

	expectedMsg := "kafka client is nil"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestClientConfiguration_ConnectTimeout_Default(t *testing.T) {
	cfg := &ClientConfiguration{}

	timeout := cfg.connectTimeout()
	if timeout != DefaultConnectTimeout {
		t.Errorf("expected default timeout %v, got %v", DefaultConnectTimeout, timeout)
	}
}

func TestValidateTopicsExist_AllTopicsExist(t *testing.T) {
	configured := []string{"topic-a", "topic-b"}
	existing := []string{"topic-a", "topic-b", "topic-c"}

	err := validateTopicsExist(configured, existing)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateTopicsExist_MissingTopics(t *testing.T) {
	configured := []string{"topic-a", "topic-missing"}
	existing := []string{"topic-a", "topic-b", "topic-c"}

	err := validateTopicsExist(configured, existing)
	if err == nil {
		t.Fatal("expected error for missing topics")
	}

	if !containsString(err.Error(), "topic-missing") {
		t.Errorf("expected error to mention missing topic, got: %v", err)
	}
}

func TestValidateTopicsExist_NoTopicsConfigured(t *testing.T) {
	configured := []string{}
	existing := []string{"topic-a", "topic-b"}

	err := validateTopicsExist(configured, existing)
	if err == nil {
		t.Fatal("expected error for empty configured topics")
	}

	if !containsString(err.Error(), "no topics configured") {
		t.Errorf("expected 'no topics configured' error, got: %v", err)
	}
}

func TestValidateTopicsExist_EmptyTopicNames(t *testing.T) {
	configured := []string{"", ""}
	existing := []string{"topic-a", "topic-b"}

	err := validateTopicsExist(configured, existing)
	if err == nil {
		t.Fatal("expected error for empty topic names")
	}

	if !containsString(err.Error(), "all configured topics are empty") {
		t.Errorf("expected 'all configured topics are empty' error, got: %v", err)
	}
}

func TestValidateTopicsExist_MixedEmptyAndValid(t *testing.T) {
	configured := []string{"", "topic-a"}
	existing := []string{"topic-a", "topic-b"}

	err := validateTopicsExist(configured, existing)
	if err != nil {
		t.Errorf("expected no error when at least one valid topic exists, got: %v", err)
	}
}

func TestValidateTopicsExist_MixedEmptyAndMissing(t *testing.T) {
	configured := []string{"", "topic-missing"}
	existing := []string{"topic-a", "topic-b"}

	err := validateTopicsExist(configured, existing)
	if err == nil {
		t.Fatal("expected error for missing topic")
	}

	if !containsString(err.Error(), "topic-missing") {
		t.Errorf("expected error to mention missing topic, got: %v", err)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
