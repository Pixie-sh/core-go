package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"time"

	"github.com/pixie-sh/errors-go"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"

	"github.com/pixie-sh/core-go/pkg/types"

	"github.com/pixie-sh/core-go/pkg/base64"
	coretime "github.com/pixie-sh/core-go/pkg/time"
)

// ClientConfiguration holds the configuration for the Kafka client
type ClientConfiguration struct {
	Brokers        []string          `json:"brokers"`
	ClientID       string            `json:"client_id"`
	SASL           *SASLConfig       `json:"sasl,omitempty"`
	TLS            *TLSConfig        `json:"tls,omitempty"`
	RetryBackoff   coretime.Duration `json:"retry_backoff"`
	RequestTimeout coretime.Duration `json:"request_timeout"`
	Compression    string            `json:"compression"` // "none", "gzip", "snappy", "lz4", "zstd"
}

type SASLConfig struct {
	Mechanism string `json:"mechanism"` // "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	CertBase64         string `json:"cert_base64,omitempty"`
	CertFile           string `json:"cert_file,omitempty"`
	KeyBase64          string `json:"key_base64,omitempty"`
	KeyFile            string `json:"key_file,omitempty"`
	CABase64           string `json:"ca_base64,omitempty"`
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
		return nil, errors.New("failed to create kafka client: %w", err)
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

// GetTopics fetches the list of topics from the Kafka cluster.
// This verifies connectivity and returns available topics.
// Useful for eager connection validation at startup.
func (c *Client) GetTopics(ctx context.Context) ([]string, error) {
	if c.kgoClient == nil {
		return nil, errors.New("kafka client is nil")
	}

	// Create a metadata request to fetch all topics (nil Topics = all topics)
	req := kmsg.NewMetadataRequest()
	req.Topics = nil // nil means fetch all topics

	resp, err := c.kgoClient.Request(ctx, &req)
	if err != nil {
		return nil, errors.New("failed to fetch metadata from kafka: %w", err)
	}

	metadataResp, ok := resp.(*kmsg.MetadataResponse)
	if !ok {
		return nil, errors.New("unexpected response type from kafka metadata request")
	}

	topics := make([]string, 0, len(metadataResp.Topics))
	for _, topic := range metadataResp.Topics {
		if topic.Topic != nil {
			topics = append(topics, *topic.Topic)
		}
	}

	return topics, nil
}

// GetKgoClient returns the underlying kgo.Client for direct access when needed
func (c *Client) GetKgoClient() *kgo.Client {
	return c.kgoClient
}

// ProduceSync implements the KafkaClient interface
func (c *Client) ProduceSync(ctx context.Context, rs ...*kgo.Record) kgo.ProduceResults {
	return c.kgoClient.ProduceSync(ctx, rs...)
}

// decodeBase64 decodes a base64-encoded string and returns the decoded bytes.
// Returns nil and an error if decoding fails.
func decodeBase64(encoded string) ([]byte, error) {
	decoded, err := base64.Decode(encoded)
	if err != nil {
		return nil, errors.New("failed to decode base64: %w", err)
	}
	return types.UnsafeBytes(decoded), nil
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
			return cfg.RetryBackoff.Duration() * time.Duration(tries)
		}))
	}

	// Configure request timeout
	if cfg.RequestTimeout > 0 {
		opts = append(opts, kgo.RequestTimeoutOverhead(cfg.RequestTimeout.Duration()))
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

		// Load client certificates - base64 takes precedence over file paths
		if cfg.TLS.CertBase64 != "" && cfg.TLS.KeyBase64 != "" {
			// Decode base64-encoded certificate and key
			certPEM, err := decodeBase64(cfg.TLS.CertBase64)
			if err == nil {
				keyPEM, err := decodeBase64(cfg.TLS.KeyBase64)
				if err == nil {
					cert, err := tls.X509KeyPair(certPEM, keyPEM)
					if err == nil {
						tlsConfig.Certificates = []tls.Certificate{cert}
					}
				}
			}
		} else if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			// Fall back to file-based certificates
			cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
			if err == nil {
				tlsConfig.Certificates = []tls.Certificate{cert}
			}
		}

		// Load CA certificate - base64 takes precedence over file paths
		if !cfg.TLS.InsecureSkipVerify {
			if cfg.TLS.CABase64 != "" {
				// Decode base64-encoded CA certificate
				caCertPEM, err := decodeBase64(cfg.TLS.CABase64)
				if err == nil {
					caCertPool := x509.NewCertPool()
					caCertPool.AppendCertsFromPEM(caCertPEM)
					tlsConfig.RootCAs = caCertPool
				}
			} else if cfg.TLS.CAFile != "" {
				// Fall back to file-based CA certificate
				caCert, err := os.ReadFile(cfg.TLS.CAFile)
				if err == nil {
					caCertPool := x509.NewCertPool()
					caCertPool.AppendCertsFromPEM(caCert)
					tlsConfig.RootCAs = caCertPool
				}
			}
		}

		opts = append(opts, kgo.Dialer((&tls.Dialer{Config: tlsConfig}).DialContext))
	}

	return opts
}
