package sqs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/events"
	"github.com/pixie-sh/core-go/infra/message_wrapper"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/models/serializer"
	utils "github.com/pixie-sh/core-go/pkg/types"
)

type ProducerConfiguration struct {
	ProducerID string             `json:"producer_id"`
	QueueURL   string             `json:"queue_url"`
	IsFIFO     bool               `json:"is_fifo"`
	CheckSize  func([]byte) error `json:"check_size"`
}

type Client interface {
	SendMessage(ctx context.Context, input *sqs.SendMessageInput, opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	SendMessageBatch(ctx context.Context, params *sqs.SendMessageBatchInput, opts ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)
}

type Producer struct {
	cfg    ProducerConfiguration
	client Client
}

func NewProducer(_ context.Context, client Client, cfg ProducerConfiguration) (*Producer, error) {
	if cfg.CheckSize == nil {
		cfg.CheckSize = func(b []byte) error {
			const (
				maxMessageSize = 262143 // 256 KiB - 1B
			)
			if len(b) > maxMessageSize {
				return errors.New("message size %d bytes exceeds maximum allowed size of %d bytes",
					len(b),
					maxMessageSize,
					errors.FieldError{
						Field:   "payload_size",
						Rule:    "payloadTooLong",
						Param:   fmt.Sprintf("%d", len(b)),
						Message: fmt.Sprintf("payload size %d bytes exceeds maximum allowed size of %d bytes", len(b), maxMessageSize),
					},
					errors.ProducerErrorCode)
			}
			return nil
		}
	}

	return &Producer{
		client: client,
		cfg:    cfg,
	}, nil
}

func (s *Producer) ProduceBatch(ctx context.Context, wrappers ...events.UntypedEventWrapper) error {
	var log = pixiecontext.GetCtxLogger(ctx)
	var entries []types.SendMessageBatchRequestEntry

	log.Debug("entry point for sqs producer. generating message attributes... ")
	messageAttributes := s.createMessageAttributes(ctx)

	log.With("message_attributes", messageAttributes).Debug("generated message attributes")
	for _, wrapper := range wrappers {
		id := aws.String(wrapper.ID)
		payload, err := serializer.Serialize(wrapper.UntypedMessage)
		if err != nil {
			pixiecontext.GetCtxLogger(ctx).
				With("event_wrapper", wrapper).
				Warn("issue serializing payload", wrapper.PayloadType)
			continue
		}

		err = s.cfg.CheckSize(payload)
		if err != nil {
			pixiecontext.GetCtxLogger(ctx).
				With("error", err).
				With("event_wrapper", wrapper).
				Error("batch issue checking payload size", wrapper.PayloadType)

			return err
		}

		entry := types.SendMessageBatchRequestEntry{
			Id:                id,
			MessageBody:       aws.String(string(payload)),
			MessageAttributes: s.appendPayloadType(ctx, wrapper.PayloadType, wrapper.ID, messageAttributes),
		}

		if s.cfg.IsFIFO {
			entry.MessageGroupId = id
			entry.MessageDeduplicationId = s.dedupID(ctx, &wrapper.UntypedMessage)
		}
		entries = append(entries, entry)
	}

	//TODO Create specific error to have failures from send batch
	log.Debug("generated sqs entries len(%d) for queue url %s", len(entries), s.cfg.QueueURL)
	res, err := s.client.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(s.cfg.QueueURL),
		Entries:  entries,
	})

	log.With("batch.result", res).With("error", err).Debug("batch produced.")
	return err
}

func (s *Producer) Produce(ctx context.Context, wrapper events.UntypedEventWrapper) error {
	var log = pixiecontext.GetCtxLogger(ctx)
	payload, err := serializer.Serialize(wrapper.UntypedMessage)
	if err != nil {
		return err
	}

	err = s.cfg.CheckSize(payload)
	if err != nil {
		pixiecontext.GetCtxLogger(ctx).
			With("error", err).
			With("event_wrapper", wrapper).
			Error("issue checking payload size", wrapper.PayloadType)

		return err
	}

	messageAttributes := s.createMessageAttributes(ctx)

	id := aws.String(wrapper.ID)
	input := &sqs.SendMessageInput{
		QueueUrl:          aws.String(s.cfg.QueueURL),
		MessageBody:       aws.String(utils.UnsafeString(payload)),
		MessageAttributes: s.appendPayloadType(ctx, wrapper.PayloadType, wrapper.ID, messageAttributes),
	}

	if s.cfg.IsFIFO {
		input.MessageGroupId = id
		input.MessageDeduplicationId = s.dedupID(ctx, &wrapper.UntypedMessage)
	}

	res, err := s.client.SendMessage(ctx, input)
	log.With("send.result", res).With("error", err).Debug("event produced.")
	return err
}

func (s *Producer) ProduceWithQueue(ctx context.Context, wrapper message_wrapper.UntypedMessage, queueUrl string, isFIFO bool) error {
	var log = pixiecontext.GetCtxLogger(ctx)

	payload, err := serializer.Serialize(wrapper)
	if err != nil {
		return err
	}

	err = s.cfg.CheckSize(payload)
	if err != nil {
		pixiecontext.GetCtxLogger(ctx).
			With("error", err).
			With("event_wrapper", wrapper).
			Error("issue checking payload size", wrapper.PayloadType)

		return err
	}

	id := aws.String(wrapper.ID)
	input := &sqs.SendMessageInput{
		QueueUrl:          aws.String(queueUrl),
		MessageBody:       aws.String(utils.UnsafeString(payload)),
		MessageAttributes: s.appendPayloadType(ctx, wrapper.PayloadType, wrapper.ID, nil),
	}

	if isFIFO {
		input.MessageGroupId = id
		input.MessageDeduplicationId = s.dedupID(ctx, &wrapper)
	}

	res, err := s.client.SendMessage(ctx, input)
	log.With("send.result", res).With("error", err).Debug("event produced.")

	return err
}

func (s *Producer) ID() string {
	return s.cfg.ProducerID
}

func (s *Producer) createMessageAttributes(ctx context.Context) map[string]types.MessageAttributeValue {
	messageAttributes := make(map[string]types.MessageAttributeValue)
	messageAttributes[env.Scope] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(env.EnvScope()),
	}
	messageAttributes[logger.TraceID] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(pixiecontext.GetCtxTraceID(ctx)),
	}
	return messageAttributes
}

func (s *Producer) appendPayloadType(ctx context.Context, payloadType string, eventID string, attributes map[string]types.MessageAttributeValue) map[string]types.MessageAttributeValue {
	if attributes == nil {
		attributes = s.createMessageAttributes(ctx)
	}

	attributes["x-payload-type"] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(payloadType),
	}

	attributes["x-event-id"] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(eventID),
	}

	return attributes
}

func (s *Producer) dedupID(_ context.Context, wrapper *message_wrapper.UntypedMessage) *string {
	return aws.String(fmt.Sprintf("%s:%d", wrapper.ID, wrapper.Timestamp.UnixMilli()))
}
