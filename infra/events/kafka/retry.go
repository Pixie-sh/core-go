package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
)

type RetryConfiguration struct {
	Enabled           bool    `json:"enabled"`
	MaxRetries        int     `json:"max_retries"`
	RetryTopicPrefix  string  `json:"retry_topic_prefix"` // e.g., "retry-"
	DLQTopic          string  `json:"dlq_topic"`          // Dead Letter Queue topic
	BackoffMultiplier float64 `json:"backoff_multiplier"`
}

type RetryManager struct {
	producer KafkaClient
	cfg      RetryConfiguration
}

func NewRetryManager(producer KafkaClient, cfg RetryConfiguration) *RetryManager {
	return &RetryManager{
		producer: producer,
		cfg:      cfg,
	}
}

// SendToRetry sends a message to a retry topic with incremented retry count
func (r *RetryManager) SendToRetry(ctx context.Context, record *kgo.Record, retryCount int, originalTopic string) error {
	if !r.cfg.Enabled || r.producer == nil {
		return nil
	}

	log := pixiecontext.GetCtxLogger(ctx)

	// Check if we've exceeded max retries
	if retryCount >= r.cfg.MaxRetries {
		log.Debug("max retries exceeded, sending to DLQ")
		return r.SendToDLQ(ctx, record, originalTopic, "max_retries_exceeded")
	}

	// Calculate retry topic name
	retryTopic := fmt.Sprintf("%s%s", r.cfg.RetryTopicPrefix, originalTopic)

	// Increment retry count in headers
	newHeaders := r.incrementRetryCount(record.Headers, retryCount+1)
	
	// Add original topic header
	newHeaders = append(newHeaders, kgo.RecordHeader{
		Key:   "x-original-topic",
		Value: []byte(originalTopic),
	})

	// Add retry reason if not present
	hasReason := false
	for _, header := range newHeaders {
		if header.Key == "x-retry-reason" {
			hasReason = true
			break
		}
	}
	if !hasReason {
		newHeaders = append(newHeaders, kgo.RecordHeader{
			Key:   "x-retry-reason",
			Value: []byte("processing_failed"),
		})
	}

	// Create retry record
	retryRecord := &kgo.Record{
		Topic:   retryTopic,
		Key:     record.Key,
		Value:   record.Value,
		Headers: newHeaders,
	}

	// Send to retry topic
	results := r.producer.ProduceSync(ctx, retryRecord)
	for _, result := range results {
		if result.Err != nil {
			log.With("error", result.Err).Error("failed to send message to retry topic %s", retryTopic)
			return fmt.Errorf("failed to send to retry topic: %w", result.Err)
		}
	}

	log.With("retry_topic", retryTopic).With("retry_count", retryCount+1).Debug("message sent to retry topic")
	return nil
}

// SendToDLQ sends a message to the dead letter queue
func (r *RetryManager) SendToDLQ(ctx context.Context, record *kgo.Record, originalTopic string, reason string) error {
	if !r.cfg.Enabled || r.producer == nil || r.cfg.DLQTopic == "" {
		return nil
	}

	log := pixiecontext.GetCtxLogger(ctx)

	// Add DLQ headers
	dlqHeaders := make([]kgo.RecordHeader, len(record.Headers))
	copy(dlqHeaders, record.Headers)
	
	dlqHeaders = append(dlqHeaders, kgo.RecordHeader{
		Key:   "x-original-topic",
		Value: []byte(originalTopic),
	})
	
	dlqHeaders = append(dlqHeaders, kgo.RecordHeader{
		Key:   "x-dlq-reason",
		Value: []byte(reason),
	})
	
	dlqHeaders = append(dlqHeaders, kgo.RecordHeader{
		Key:   "x-dlq-timestamp",
		Value: []byte(time.Now().UTC().Format(time.RFC3339)),
	})

	// Create DLQ record
	dlqRecord := &kgo.Record{
		Topic:   r.cfg.DLQTopic,
		Key:     record.Key,
		Value:   record.Value,
		Headers: dlqHeaders,
	}

	// Send to DLQ
	results := r.producer.ProduceSync(ctx, dlqRecord)
	for _, result := range results {
		if result.Err != nil {
			log.With("error", result.Err).Error("failed to send message to DLQ %s", r.cfg.DLQTopic)
			return fmt.Errorf("failed to send to DLQ: %w", result.Err)
		}
	}

	log.With("dlq_topic", r.cfg.DLQTopic).With("reason", reason).Debug("message sent to DLQ")
	return nil
}

// CalculateRetryDelay calculates the delay for a retry attempt
func (r *RetryManager) CalculateRetryDelay(retryCount int) time.Duration {
	baseDelay := time.Second * 30 // 30 seconds base delay
	multiplier := r.cfg.BackoffMultiplier
	if multiplier <= 0 {
		multiplier = 2.0
	}

	delay := float64(baseDelay.Nanoseconds())
	for i := 0; i < retryCount; i++ {
		delay *= multiplier
	}

	return time.Duration(int64(delay))
}

// GetRetryCount extracts retry count from headers
func (r *RetryManager) GetRetryCount(headers []kgo.RecordHeader) int {
	for _, header := range headers {
		if header.Key == "x-retry-count" {
			count, err := strconv.Atoi(string(header.Value))
			if err != nil {
				return 0
			}
			return count
		}
	}
	return 0
}

// incrementRetryCount increments the retry count in headers
func (r *RetryManager) incrementRetryCount(headers []kgo.RecordHeader, newCount int) []kgo.RecordHeader {
	var newHeaders []kgo.RecordHeader
	retryCountSet := false

	for _, header := range headers {
		if header.Key == "x-retry-count" {
			newHeaders = append(newHeaders, kgo.RecordHeader{
				Key:   "x-retry-count",
				Value: []byte(strconv.Itoa(newCount)),
			})
			retryCountSet = true
		} else {
			newHeaders = append(newHeaders, header)
		}
	}

	// If retry count header wasn't present, add it
	if !retryCountSet {
		newHeaders = append(newHeaders, kgo.RecordHeader{
			Key:   "x-retry-count",
			Value: []byte(strconv.Itoa(newCount)),
		})
	}

	return newHeaders
}

// GetOriginalTopic extracts the original topic from headers
func (r *RetryManager) GetOriginalTopic(headers []kgo.RecordHeader) string {
	for _, header := range headers {
		if header.Key == "x-original-topic" {
			return string(header.Value)
		}
	}
	return ""
}

// IsRetryTopic checks if a topic is a retry topic
func (r *RetryManager) IsRetryTopic(topic string) bool {
	return len(topic) > len(r.cfg.RetryTopicPrefix) && 
		   topic[:len(r.cfg.RetryTopicPrefix)] == r.cfg.RetryTopicPrefix
}

// IsDLQTopic checks if a topic is the DLQ topic
func (r *RetryManager) IsDLQTopic(topic string) bool {
	return topic == r.cfg.DLQTopic
}

// SetProducer allows setting the producer after retry manager creation
func (r *RetryManager) SetProducer(producer KafkaClient) {
	r.producer = producer
}