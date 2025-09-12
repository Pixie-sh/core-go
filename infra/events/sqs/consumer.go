package sqs

import (
	"context"
	"runtime/debug"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/events"
	"github.com/pixie-sh/core-go/infra/message_factory"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	pixietypes "github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/types/slices"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type ConsumerConfiguration struct {
	QueueURL                  string `json:"queue_url"`
	MaxNumberOfMessages       int32  `json:"max_number_of_messages"`
	WaitTimeSeconds           int32  `json:"wait_time_seconds"`
	RequeueBackoffTimeSeconds int32  `json:"requeue_backoff_time_seconds"`
	RequeueMaxRetries         int    `json:"requeue_max_retries"`
	WithoutScope              bool   `json:"without_scope,omitempty"`
}

type Consumer struct {
	cfg          ConsumerConfiguration
	client       *SQSClient
	allowedScope func(types.Message) bool
}

func NewConsumer(_ context.Context, client *SQSClient, cfg ConsumerConfiguration) (*Consumer, error) {
	return &Consumer{
		client: client,
		cfg:    cfg,
		allowedScope: func(message types.Message) bool {
			if cfg.WithoutScope {
				return true
			}

			reqScope, ok := message.MessageAttributes[env.Scope]
			if !ok {
				return false
			}

			scope := env.EnvScope()
			return *reqScope.StringValue == scope
		},
	}, nil
}

// ConsumeBatch it's blocking call
func (s *Consumer) ConsumeBatch(ctx context.Context, handler func(context.Context, events.UntypedEventWrapper) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.With("stack_trace", pixietypes.UnsafeString(debug.Stack())).Error("consumer recovered from panic: %+v", r)
			err = s.ConsumeBatch(ctx, handler) // TODO: rafa improve this
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			output, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:                    aws.String(s.cfg.QueueURL),
				MaxNumberOfMessages:         s.cfg.MaxNumberOfMessages,
				WaitTimeSeconds:             s.cfg.WaitTimeSeconds,
				MessageAttributeNames:       []string{env.Scope},
				MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameApproximateReceiveCount},
			})
			if err != nil {
				return err
			}

			if len(output.Messages) == 0 {
				continue
			}

			var wrappers []events.UntypedEventWrapper

			log := pixiecontext.GetCtxLogger(ctx)
			for _, message := range output.Messages {
				wrapper, err := message_factory.Singleton.CreateFromString(ctx, *message.Body)
				if err != nil {
					innerlog := log.With("sqs_message", message).With("error", err)

					herr, haz := errors.Has(err, errors.InvalidTypeErrorCode)
					if haz {
						field := slices.Find(herr.FieldErrors, func(ferr *errors.FieldError) bool {
							return ferr != nil && ferr.Rule == "invalidPayloadType"
						})

						innerlog = innerlog.With(field.Field, field.Param)
					}

					innerlog.Error("error deserializing message")
					s.requeueOrDelete(ctx, log, errors.New(err.Error(), errors.NoRetryErrorCode), message)
					continue
				}

				if !s.allowedScope(message) {
					log.With("sqs_message", message).With("error", err).Error("scope is invalid")
					s.requeueOrDelete(ctx, log, errors.New("invalid scope", errors.InvalidScopeRequeueErrorCode), message)
					continue
				}

				wrapper.SetHeader("sqs.message", message) //TODO: maybe change this to event.locals instead of headers
				wrapper.SetHeader("sqs.receipt_handle", message.ReceiptHandle)
				wrapper.SetHeader("sqs.approximate_receive_count", message.Attributes[string(types.MessageSystemAttributeNameApproximateReceiveCount)])

				wrappers = append(wrappers, events.NewUntypedEventWrapperFromMessage(wrapper))
			}

			for i := range wrappers {
				traceID := uid.NewUUID()
				requestLog := pixiecontext.GetCtxLogger(ctx)
				requestCtx := pixiecontext.SetCtxLogger(
					context.Background(),
					requestLog.With(logger.TraceID, traceID).With("event_message", wrappers[i]),
				)
				requestCtx = pixiecontext.SetCtxTraceID(requestCtx, traceID)

				err = handler(requestCtx, wrappers[i])
				if err != nil {
					requestLog.With("error", err).Error("error processing batch messages")
					s.requeueOrDelete(ctx, requestLog, err, wrappers[i].GetHeader("sqs.message").(types.Message))
					continue
				}

				err = s.Delete(ctx, wrappers[i].GetHeader("sqs.receipt_handle").(*string))
				if err != nil {
					requestLog.Error("error deleting message %s", wrappers[i].GetHeader("sqs.receipt_handle").(*string))
				}

			}
		}
	}
}

// Consume it's blocking call
func (s *Consumer) Consume(ctx context.Context, handler func(context.Context, events.UntypedEventWrapper) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("consumer recovered from panic: %+v", r)
			err = s.Consume(ctx, handler) // TODO: rafa improve this
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			output, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:                    aws.String(s.cfg.QueueURL),
				MaxNumberOfMessages:         1,
				WaitTimeSeconds:             s.cfg.WaitTimeSeconds,
				MessageAttributeNames:       []string{env.Scope},
				MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameApproximateReceiveCount},
			})
			if err != nil {
				return err
			}

			if len(output.Messages) == 0 {
				continue
			}

			message := output.Messages[0]
			log := pixiecontext.GetCtxLogger(ctx).With("sqs_message", message)

			wrapper, err := message_factory.Singleton.CreateFromString(ctx, *message.Body)
			if err != nil {
				innerlog := log.With("sqs_message", message).With("error", err)

				herr, haz := errors.Has(err, errors.InvalidTypeErrorCode)
				if haz {
					field := slices.Find(herr.FieldErrors, func(ferr *errors.FieldError) bool {
						return ferr != nil && ferr.Rule == "invalidPayloadType" //TODO this kind of fields should start be consts
					})

					innerlog = innerlog.With(field.Field, field.Param)
				}

				innerlog.Error("error deserializing message")
				s.requeueOrDelete(ctx, log, errors.New(err.Error(), errors.NoRetryErrorCode), message)
				continue
			}

			if !s.allowedScope(message) {
				log.With("sqs_message", message).With("error", err).Error("scope is invalid")
				s.requeueOrDelete(ctx, log, errors.New("invalid scope", errors.InvalidScopeRequeueErrorCode), message)
				continue
			}

			wrapper.SetHeader("sqs.receipt_handle", *message.ReceiptHandle)
			wrapper.SetHeader("sqs.approximate_receive_count", message.Attributes[string(types.MessageSystemAttributeNameApproximateReceiveCount)])

			requestLog := pixiecontext.GetCtxLogger(ctx)
			traceID := uid.NewUUID()
			requestCtx := pixiecontext.SetCtxLogger(
				context.Background(),
				requestLog.With(logger.TraceID, traceID).With("event_message", wrapper).With("sqs_message", message),
			)
			requestCtx = pixiecontext.SetCtxTraceID(requestCtx, traceID)

			err = handler(requestCtx, events.NewUntypedEventWrapperFromMessage(wrapper))
			if err != nil {
				requestLog.With("error", err).Error("error processing message %s", *message.ReceiptHandle)
				s.requeueOrDelete(ctx, log, err, message)
				continue
			}

			err = s.Delete(ctx, message.ReceiptHandle)
			if err != nil {
				log.Error("error deleting message %s", message.ReceiptHandle)
			}
		}
	}
}

func (s *Consumer) Delete(ctx context.Context, handle *string) error {
	// delete processed message
	_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.cfg.QueueURL),
		ReceiptHandle: handle,
	})

	return err
}

func (s *Consumer) Requeue(ctx context.Context, handle *string) error {
	_, err := s.client.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(s.cfg.QueueURL),
		ReceiptHandle:     handle,
		VisibilityTimeout: s.cfg.RequeueBackoffTimeSeconds,
	})

	return err
}

func (s *Consumer) requeueOrDelete(ctx context.Context, log logger.Interface, err error, message types.Message) {
	_, hasScopeCode := errors.Has(err, errors.InvalidScopeRequeueErrorCode)
	_, has := errors.Has(err, errors.ProcessFailedDoNotRequeueErrorCode)
	_, haz := errors.Has(err, errors.NoRetryErrorCode)

	receiveCount, err := strconv.Atoi(message.Attributes[string(types.MessageSystemAttributeNameApproximateReceiveCount)])
	if err != nil {
		//the message may not have the attributes, handle as error only
		receiveCount = s.cfg.RequeueMaxRetries
	}

	log.Log(
		"evaluating if it's for deletion. scopeInvalid: %t ; processFailedNotRequeue: %t ; nonRetriableError: %t ",
		hasScopeCode,
		has,
		haz,
	)
	if (!haz || !has || hasScopeCode) && receiveCount <= s.cfg.RequeueMaxRetries {
		log.Debug("executing requeue")
		err = s.Requeue(ctx, message.ReceiptHandle)
		if err == nil {
			return
		}

		log.With("error", err).Error("error requeue-ing message %s", message.ReceiptHandle)
	}

	log.Debug("executing deletion")
	err = s.Delete(ctx, message.ReceiptHandle)
	if err != nil {
		log.With("error", err).Error("error deleting message %s", message.ReceiptHandle)
		return
	}
}
