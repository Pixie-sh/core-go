package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

// ClientConfiguration holds the configuration for the Kafka client
type ClientConfiguration struct {
	Brokers        []string      `json:"brokers"`
	ClientID       string        `json:"client_id"`
	SASL           *SASLConfig   `json:"sasl,omitempty"`
	TLS            *TLSConfig    `json:"tls,omitempty"`
	RetryBackoff   time.Duration `json:"retry_backoff"`
	RequestTimeout time.Duration `json:"request_timeout"`
	Compression    string        `json:"compression"` // "none", "gzip", "snappy", "lz4", "zstd"
}

type SASLConfig struct {
	Mechanism string `json:"mechanism"` // "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	CertFile           string `json:"cert_file,omitempty"`
	KeyFile            string `json:"key_file,omitempty"`
	CAFile             string `json:"ca_file,omitempty"`
}

// Client is an abstraction over the franz-go Kafka client
type Client struct {
	kgoClient *kgo.Client
	cfg       *ClientConfiguration
}

// NewClient creates a new Kafka client with the given configuration
func NewClient(_ context.Context, cfg *ClientConfiguration) (*Client, error) {
	opts := buildKgoOpts(cfg)

	kgoClient, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return &Client{
		kgoClient: kgoClient,
		cfg:       cfg,
	}, nil
}

// Close closes the Kafka client
func (c *Client) Close() {
	if c.kgoClient != nil {
		c.kgoClient.Close()
	}
}

// GetKgoClient returns the underlying kgo.Client for direct access when needed
func (c *Client) GetKgoClient() *kgo.Client {
	return c.kgoClient
}

// ProduceSync implements the KafkaClient interface
func (c *Client) ProduceSync(ctx context.Context, rs ...*kgo.Record) kgo.ProduceResults {
	return c.kgoClient.ProduceSync(ctx, rs...)
}

// buildKgoOpts builds the kgo options from the configuration
func buildKgoOpts(cfg *ClientConfiguration) []kgo.Opt {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelWarn, nil)),
	}

	if len(cfg.Compression) > 0 {
		switch cfg.Compression {
		case "gzip":
			opts = append(opts, kgo.ProducerBatchCompression(kgo.GzipCompression()))
		case "snappy":
			opts = append(opts, kgo.ProducerBatchCompression(kgo.SnappyCompression()))
		case "lz4":
			opts = append(opts, kgo.ProducerBatchCompression(kgo.Lz4Compression()))
		case "zstd":
			opts = append(opts, kgo.ProducerBatchCompression(kgo.ZstdCompression()))
		}
	}

	// Set client ID if provided
	if cfg.ClientID != "" {
		opts = append(opts, kgo.ClientID(cfg.ClientID))
	}

	// Configure retry backoff
	if cfg.RetryBackoff > 0 {
		opts = append(opts, kgo.RetryBackoffFn(func(tries int) time.Duration {
			return cfg.RetryBackoff * time.Duration(tries)
		}))
	}

	// Configure request timeout
	if cfg.RequestTimeout > 0 {
		opts = append(opts, kgo.RequestTimeoutOverhead(cfg.RequestTimeout))
	}

	// Configure SASL authentication
	if cfg.SASL != nil {
		switch cfg.SASL.Mechanism {
		case "PLAIN":
			opts = append(opts, kgo.SASL(plain.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}.AsMechanism()))
		case "SCRAM-SHA-256":
			opts = append(opts, kgo.SASL(scram.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}.AsSha256Mechanism()))
		case "SCRAM-SHA-512":
			opts = append(opts, kgo.SASL(scram.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}.AsSha512Mechanism()))
		}
	}

	// Configure TLS
	if cfg.TLS != nil && cfg.TLS.Enabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		}

		// Load client certificates if provided
		if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
			if err == nil {
				tlsConfig.Certificates = []tls.Certificate{cert}
			}
		}

		// Load CA certificate if provided
		if cfg.TLS.CAFile != "" {
			caCert, err := os.ReadFile(cfg.TLS.CAFile)
			if err == nil {
				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)
				tlsConfig.RootCAs = caCertPool
			}
		}

		opts = append(opts, kgo.Dialer((&tls.Dialer{Config: tlsConfig}).DialContext))
	}

	return opts
}
