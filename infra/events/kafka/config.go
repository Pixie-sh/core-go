package kafka

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// LoadClientConfigFromEnv loads client configuration from environment variables
func LoadClientConfigFromEnv() ClientConfiguration {
	cfg := ClientConfiguration{
		Brokers:        strings.Split(getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), ","),
		ClientID:       getEnvOrDefault("KAFKA_CLIENT_ID", "pixie-kafka-client"),
		RetryBackoff:   parseDurationOrDefault(getEnvOrDefault("KAFKA_RETRY_BACKOFF", "1s")),
		RequestTimeout: parseDurationOrDefault(getEnvOrDefault("KAFKA_REQUEST_TIMEOUT", "30s")),
	}

	// Configure SASL if enabled
	saslMechanism := os.Getenv("KAFKA_SASL_MECHANISM")
	if saslMechanism != "" {
		cfg.SASL = &SASLConfig{
			Mechanism: saslMechanism,
			Username:  os.Getenv("KAFKA_SASL_USERNAME"),
			Password:  os.Getenv("KAFKA_SASL_PASSWORD"),
		}
	}

	// Configure TLS if enabled
	if parseBoolOrDefault(os.Getenv("KAFKA_TLS_ENABLED")) {
		cfg.TLS = &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: parseBoolOrDefault(os.Getenv("KAFKA_TLS_INSECURE")),
			CertFile:           os.Getenv("KAFKA_TLS_CERT_FILE"),
			KeyFile:            os.Getenv("KAFKA_TLS_KEY_FILE"),
			CAFile:             os.Getenv("KAFKA_TLS_CA_FILE"),
		}
	}

	return cfg
}

// LoadProducerConfigFromEnv loads producer configuration from environment variables
func LoadProducerConfigFromEnv(topic string, producerID ...string) ProducerConfiguration {
	id := "pixie-kafka-producer"
	if len(producerID) > 0 && producerID[0] != "" {
		id = producerID[0]
	}

	cfg := ProducerConfiguration{
		ProducerID:     id,
		Topic:          topic,
		Compression:    getEnvOrDefault("KAFKA_COMPRESSION", "none"),
		Idempotent:     parseBoolOrDefault(os.Getenv("KAFKA_IDEMPOTENT_PRODUCER")),
		MaxMessageSize: parseIntOrDefault(getEnvOrDefault("KAFKA_MAX_MESSAGE_SIZE", "1048576")), // 1MB default
	}

	// Set default CheckSize function based on MaxMessageSize
	if cfg.MaxMessageSize > 0 {
		maxSize := cfg.MaxMessageSize
		cfg.CheckSize = func(b []byte) error {
			if len(b) > maxSize {
				return fmt.Errorf("message size %d bytes exceeds maximum allowed size of %d bytes", len(b), maxSize)
			}
			return nil
		}
	}

	return cfg
}

// LoadConsumerConfigFromEnv loads consumer configuration from environment variables
func LoadConsumerConfigFromEnv(consumerGroup string, topics ...string) ConsumerConfiguration {
	cfg := ConsumerConfiguration{
		Topics:                    topics,
		ConsumerGroup:             consumerGroup,
		MaxBatchSize:              int32(parseIntOrDefault(getEnvOrDefault("KAFKA_MAX_BATCH_SIZE", "100"))),
		PollTimeout:               parseDurationOrDefault(getEnvOrDefault("KAFKA_POLL_TIMEOUT", "5s")),
		RequeueBackoffTimeSeconds: int32(parseIntOrDefault(getEnvOrDefault("KAFKA_REQUEUE_BACKOFF_SECONDS", "30"))),
		RequeueMaxRetries:         parseIntOrDefault(getEnvOrDefault("KAFKA_REQUEUE_MAX_RETRIES", "3")),
		WithoutScope:              parseBoolOrDefault(os.Getenv("KAFKA_WITHOUT_SCOPE")),
		AutoCommit:                parseBoolOrDefault(getEnvOrDefault("KAFKA_AUTO_COMMIT", "true")),
		StartOffset:               getEnvOrDefault("KAFKA_START_OFFSET", "latest"),
	}

	return cfg
}

// LoadRetryConfigFromEnv loads retry configuration from environment variables
func LoadRetryConfigFromEnv() RetryConfiguration {
	return RetryConfiguration{
		Enabled:           parseBoolOrDefault(getEnvOrDefault("KAFKA_RETRY_ENABLED", "true")),
		MaxRetries:        parseIntOrDefault(getEnvOrDefault("KAFKA_RETRY_MAX_RETRIES", "3")),
		RetryTopicPrefix:  getEnvOrDefault("KAFKA_RETRY_TOPIC_PREFIX", "retry-"),
		DLQTopic:          getEnvOrDefault("KAFKA_DLQ_TOPIC", "dlq"),
		BackoffMultiplier: parseFloatOrDefault(getEnvOrDefault("KAFKA_RETRY_BACKOFF_MULTIPLIER", "2.0")),
	}
}

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseBoolOrDefault(value string) bool {
	if value == "" {
		return false
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return result
}

func parseIntOrDefault(value string) int {
	if value == "" {
		return 0
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return result
}

func parseFloatOrDefault(value string) float64 {
	if value == "" {
		return 0.0
	}
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0
	}
	return result
}

func parseDurationOrDefault(value string) time.Duration {
	if value == "" {
		return 0
	}
	result, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return result
}

// Environment variable reference:
/*
KAFKA_BROKERS                   = comma-separated broker list (default: localhost:9092)
KAFKA_CLIENT_ID                 = client identifier (default: pixie-kafka-client)
KAFKA_RETRY_BACKOFF             = retry backoff duration (default: 1s)
KAFKA_REQUEST_TIMEOUT           = request timeout duration (default: 30s)

// SASL Authentication
KAFKA_SASL_MECHANISM            = SASL mechanism: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
KAFKA_SASL_USERNAME             = SASL username
KAFKA_SASL_PASSWORD             = SASL password

// TLS Configuration
KAFKA_TLS_ENABLED               = enable TLS (true/false)
KAFKA_TLS_INSECURE              = skip TLS verification (true/false)
KAFKA_TLS_CERT_FILE             = client certificate file path
KAFKA_TLS_KEY_FILE              = client private key file path
KAFKA_TLS_CA_FILE               = CA certificate file path

// Producer Configuration
KAFKA_COMPRESSION               = compression algorithm: none, gzip, snappy, lz4, zstd (default: none)
KAFKA_IDEMPOTENT_PRODUCER       = enable idempotent producer (true/false)
KAFKA_MAX_MESSAGE_SIZE          = maximum message size in bytes (default: 1048576 = 1MB)

// Consumer Configuration
KAFKA_MAX_BATCH_SIZE            = maximum batch size for batch consumption (default: 100)
KAFKA_POLL_TIMEOUT              = polling timeout duration (default: 5s)
KAFKA_REQUEUE_BACKOFF_SECONDS   = requeue backoff time in seconds (default: 30)
KAFKA_REQUEUE_MAX_RETRIES       = maximum number of retries (default: 3)
KAFKA_WITHOUT_SCOPE             = disable scope validation (true/false)
KAFKA_AUTO_COMMIT               = enable auto commit (default: true)
KAFKA_START_OFFSET              = start offset: earliest, latest (default: latest)

// Retry Configuration
KAFKA_RETRY_ENABLED             = enable retry mechanism (default: true)
KAFKA_RETRY_MAX_RETRIES         = maximum retry attempts (default: 3)
KAFKA_RETRY_TOPIC_PREFIX        = prefix for retry topics (default: retry-)
KAFKA_DLQ_TOPIC                 = dead letter queue topic name (default: dlq)
KAFKA_RETRY_BACKOFF_MULTIPLIER  = backoff multiplier for retry delays (default: 2.0)
*/