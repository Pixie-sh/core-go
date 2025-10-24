package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"
	
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/pixie-sh/core-go/infra/events"
	"github.com/pixie-sh/core-go/infra/message_wrapper"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/models/serializer"
	coretime "github.com/pixie-sh/core-go/pkg/time"
)

type ProducerConfiguration struct {
	ProducerID         string                                  `json:"producer_id"`
	Topic              string                                  `json:"topic"`
	PartitionKey       func(events.UntypedEventWrapper) []byte `json:"-"` // Function to extract partition key
	MaxMessageSize     int                                     `json:"max_message_size"`
	RetryUntilDuration coretime.Duration                       `json:"retry_until_duration"` // Default: 10 minutes
}

type Producer struct {
	cfg    *ProducerConfiguration
	client *Client
}

func NewProducer(_ context.Context, client *Client, cfg *ProducerConfiguration) (*Producer, error) {
	return &Producer{
		client: client,
		cfg:    cfg,
	}, nil
}

func (p *Producer) ProduceBatch(ctx context.Context, wrappers ...events.UntypedEventWrapper) error {
	var log = pixiecontext.GetCtxLogger(ctx)
	var records []*kgo.Record

	log.Debug("entry point for kafka producer. generating message headers... ")
	messageHeaders := p.createHeaders(ctx)

	log.With("message_headers", messageHeaders).Debug("generated message headers")
	for _, wrapper := range wrappers {
		payload, err := serializer.Serialize(wrapper.UntypedMessage)
		if err != nil {
			pixiecontext.GetCtxLogger(ctx).
				With("event_wrapper", wrapper).
				Warn("issue serializing payload", wrapper.PayloadType)
			continue
		}

		headers := p.appendPayloadType(ctx, wrapper.PayloadType, wrapper.ID, messageHeaders)

		var key []byte
		if p.cfg.PartitionKey != nil {
			key = p.cfg.PartitionKey(wrapper)
		}

		record := &kgo.Record{
			Topic:   p.cfg.Topic,
			Key:     key,
			Value:   payload,
			Headers: headers,
		}

		records = append(records, record)
	}

	log.Debug("generated kafka records len(%d) for topic %s", len(records), p.cfg.Topic)
	results := p.client.kgoClient.ProduceSync(ctx, records...)

	// Check for errors in batch results
	for _, result := range results {
		if result.Err != nil {
			log.With("error", result.Err).Error("failed to produce message to topic %s", result.Record.Topic)
			return errors.Wrap(result.Err, "kafka produce error")
		}
	}

	log.With("batch.results", len(results)).Debug("batch produced.")
	return nil
}

func (p *Producer) Produce(ctx context.Context, wrapper events.UntypedEventWrapper) error {
	var log = pixiecontext.GetCtxLogger(ctx)
	payload, err := serializer.Serialize(wrapper.UntypedMessage)
	if err != nil {
		return err
	}

	messageHeaders := p.createHeaders(ctx)
	headers := p.appendPayloadType(ctx, wrapper.PayloadType, wrapper.ID, messageHeaders)

	// Build partition key if function is provided
	var key []byte
	if p.cfg.PartitionKey != nil {
		key = p.cfg.PartitionKey(wrapper)
	}

	record := &kgo.Record{
		Topic:   p.cfg.Topic,
		Key:     key,
		Value:   payload,
		Headers: headers,
	}

	results := p.client.kgoClient.ProduceSync(ctx, record)
	for _, result := range results {
		if result.Err != nil {
			log.With("error", result.Err).Error("failed to produce message to topic %s", result.Record.Topic)
			return errors.Wrap(result.Err, "kafka produce error")
		}
	}

	log.With("topic", p.cfg.Topic).Debug("event produced.")
	return nil
}

func (p *Producer) ProduceWithTopic(ctx context.Context, wrapper message_wrapper.UntypedMessage, topic string, partitionKey []byte) error {
	var log = pixiecontext.GetCtxLogger(ctx)

	payload, err := serializer.Serialize(wrapper)
	if err != nil {
		return err
	}

	headers := p.appendPayloadType(ctx, wrapper.PayloadType, wrapper.ID, nil)
	record := &kgo.Record{
		Topic:   topic,
		Key:     partitionKey,
		Value:   payload,
		Headers: headers,
	}

	results := p.client.kgoClient.ProduceSync(ctx, record)
	for _, result := range results {
		if result.Err != nil {
			log.With("error", result.Err).Error("failed to produce message to topic %s", result.Record.Topic)
			return fmt.Errorf("kafka produce error: %w", result.Err)
		}
	}

	log.With("topic", topic).Debug("event produced.")
	return nil
}

func (p *Producer) ID() string {
	return p.cfg.ProducerID
}

func (p *Producer) createHeaders(ctx context.Context) []kgo.RecordHeader {
	var headers []kgo.RecordHeader

	headers = append(headers, kgo.RecordHeader{
		Key:   env.Scope,
		Value: []byte(env.EnvScope()),
	})

	headers = append(headers, kgo.RecordHeader{
		Key:   logger.TraceID,
		Value: []byte(pixiecontext.GetCtxTraceID(ctx)),
	})

	retryUntilDuration := p.cfg.RetryUntilDuration
	if retryUntilDuration == 0 {
		retryUntilDuration = coretime.Duration(10 * time.Minute)
	}

	retryUntilMillis := time.Now().UnixMilli() + retryUntilDuration.Duration().Milliseconds()
	headers = append(headers, kgo.RecordHeader{
		Key:   XRetryUntilHeader,
		Value: []byte(strconv.FormatInt(retryUntilMillis, 10)),
	})

	return headers
}

func (p *Producer) appendPayloadType(ctx context.Context, payloadType string, eventID string, headers []kgo.RecordHeader) []kgo.RecordHeader {
	if headers == nil {
		headers = p.createHeaders(ctx)
	}

	headers = append(headers, kgo.RecordHeader{
		Key:   XPayloadTypeHeader,
		Value: []byte(payloadType),
	})

	headers = append(headers, kgo.RecordHeader{
		Key:   XEventIDHeader,
		Value: []byte(eventID),
	})

	return headers
}
