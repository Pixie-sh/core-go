package kafka

import (
	"context"
	"encoding/base64"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/pixie-sh/core-go/infra/events"
	"github.com/pixie-sh/core-go/infra/message_factory"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	pixietypes "github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/types/slices"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type ConsumerConfiguration struct {
	Topics            []string `json:"topics"`
	ConsumerGroup     string   `json:"consumer_group"`
	RequeueMaxRetries int      `json:"requeue_max_retries"`
	WithoutScope      bool     `json:"without_scope,omitempty"`
	AutoCommit        bool     `json:"auto_commit"`
	StartOffset       string   `json:"start_offset"` // "earliest", "latest"
}

type Consumer struct {
	cfg          *ConsumerConfiguration
	client       *Client
	allowedScope func(*kgo.Record) bool
	retryManager *RetryManager
}

func NewConsumer(_ context.Context, client *Client, cfg *ConsumerConfiguration) (*Consumer, error) {
	// Configure consumer group and topics
	opts := []kgo.Opt{
		kgo.ConsumerGroup(cfg.ConsumerGroup),
		kgo.ConsumeTopics(cfg.Topics...),
	}

	// Set start offset
	switch cfg.StartOffset {
	case "earliest":
		opts = append(opts, kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	case "latest":
		opts = append(opts, kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()))
	}

	// Auto commit configuration
	if !cfg.AutoCommit {
		opts = append(opts, kgo.DisableAutoCommit())
	}

	// Create a new client for consuming (separate from producer client)
	consumerClient, err := kgo.NewClient(append(buildKgoOpts(client.cfg), opts...)...)
	if err != nil {
		return nil, err
	}

	// Replace the client's kgoClient with the consumer-configured one
	client.kgoClient.Close()
	client.kgoClient = consumerClient

	consumer := &Consumer{
		client: client,
		cfg:    cfg,
		allowedScope: func(record *kgo.Record) bool {
			if cfg.WithoutScope {
				return true
			}

			for _, header := range record.Headers {
				if header.Key == env.Scope {
					scope := env.EnvScope()
					return string(header.Value) == scope
				}
			}
			return false
		},
	}

	// Initialize retry manager if we have requeue configuration
	if cfg.RequeueMaxRetries > 0 {
		retryConfig := RetryConfiguration{
			Enabled:           true,
			MaxRetries:        cfg.RequeueMaxRetries,
			RetryTopicPrefix:  "retry-",
			DLQTopic:          "dlq",
			BackoffMultiplier: 2.0,
		}

		// Create a producer for retry operations - we'll need to pass this separately
		// For now, we'll set it to nil and require it to be set later
		consumer.retryManager = NewRetryManager(nil, retryConfig)
	}

	return consumer, nil
}

// ConsumeBatch it's blocking call - matches SQS interface exactly
func (c *Consumer) ConsumeBatch(ctx context.Context, handler func(context.Context, events.UntypedEventWrapper) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.With("stack_trace", pixietypes.UnsafeString(debug.Stack())).Error("consumer recovered from panic: %+v", r)
			err = c.ConsumeBatch(ctx, handler) // TODO: improve this
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fetches := c.client.kgoClient.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				// Handle errors
				for _, fetchErr := range errs {
					pixiecontext.GetCtxLogger(ctx).With("error", fetchErr.Err).Error("error fetching from kafka")
				}
				continue
			}

			if fetches.NumRecords() == 0 {
				continue
			}

			var wrappers []events.UntypedEventWrapper
			log := pixiecontext.GetCtxLogger(ctx)

			fetches.EachPartition(func(p kgo.FetchTopicPartition) {
				for _, record := range p.Records {
					wrapper, err := c.processRecord(ctx, log, record)
					if err != nil {
						continue // Error was already logged in processRecord
					}
					if wrapper != nil {
						wrappers = append(wrappers, *wrapper)
					}
				}
			})

			// In batch mode, to enforce Behavior A, if any record in the batch fails,
			// we will not commit any offsets from this batch. If all succeed, we commit
			// the last record per topic/partition.
			batchFailed := false
			lastRecordPerTP := make(map[string]*kgo.Record) // key: topic-partition

			for i := range wrappers {
				traceID := uid.NewUUID()
				requestLog := pixiecontext.GetCtxLogger(ctx)
				requestCtx := pixiecontext.SetCtxLogger(
					context.Background(),
					requestLog.With(logger.TraceID, traceID).With("event_message", wrappers[i]),
				)
				requestCtx = pixiecontext.SetCtxTraceID(requestCtx, traceID)

				err = handler(requestCtx, wrappers[i])
				rec := wrappers[i].GetHeader("kafka.record").(*kgo.Record)
				if err != nil {
					batchFailed = true
					requestLog.With("error", err).Error("error processing batch messages")
					c.requeueOrDelete(ctx, requestLog, err, rec)
					continue
				}

				// Track highest offset per topic/partition for successful records
				key := rec.Topic + ":" + strconv.Itoa(int(rec.Partition))
				if existing, ok := lastRecordPerTP[key]; !ok || rec.Offset > existing.Offset {
					lastRecordPerTP[key] = rec
				}
			}

			// Commit only if all records in the batch succeeded and auto-commit is disabled
			if !c.cfg.AutoCommit && !batchFailed {
				for _, rec := range lastRecordPerTP {
					if err := c.commitRecord(ctx, rec); err != nil {
						log.Error("error committing message offset")
					}
				}
			}
		}
	}
}

// Consume it's blocking call - matches SQS interface exactly
func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, events.UntypedEventWrapper) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("consumer recovered from panic: %+v", r)
			err = c.Consume(ctx, handler) // TODO: improve this
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fetches := c.client.kgoClient.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				// Handle errors
				for _, fetchErr := range errs {
					pixiecontext.GetCtxLogger(ctx).With("error", fetchErr.Err).Error("error fetching from kafka")
				}
				continue
			}

			if fetches.NumRecords() == 0 {
				continue
			}

			fetches.EachPartition(func(p kgo.FetchTopicPartition) {

				log := pixiecontext.GetCtxLogger(ctx)
				for _, record := range p.Records {
					wrapper, err := c.processRecord(ctx, log.With("kafka_record", record), record)
					if err != nil {
						continue // Error was already logged in processRecord
					}
					if wrapper == nil {
						continue
					}

					requestLog := pixiecontext.GetCtxLogger(ctx)
					traceID := uid.NewUUID()
					requestCtx := pixiecontext.SetCtxLogger(
						context.Background(),
						requestLog.With(logger.TraceID, traceID).With("event_message", wrapper).With("kafka_record", record),
					)
					requestCtx = pixiecontext.SetCtxTraceID(requestCtx, traceID)

					err = handler(requestCtx, *wrapper)
					if err != nil {
						requestLog.
							With("error", err).
							With("topic", record.Topic).
							With("partition", record.Partition).
							With("offset", record.Offset).
							Error("error processing message")

						c.requeueOrDelete(ctx, log, err, record)
						continue
					}

					// Commit the message if not auto-commit
					if !c.cfg.AutoCommit {
						err = c.commitRecord(ctx, record)
						if err != nil {
							log.With(
								"topic", record.Topic).
								With("partition", record.Partition).
								With("offset", record.Offset).
								Error("error committing message offset")
						}
					}
				}
			})
		}
	}
}

func (c *Consumer) processRecord(ctx context.Context, log logger.Interface, record *kgo.Record) (*events.UntypedEventWrapper, error) {
	wrapper, err := message_factory.Singleton.Create(ctx, record.Value)
	if err != nil {
		innerlog := log.With("kafka_record", record).With("error", err)

		herr, haz := errors.Has(err, errors.InvalidTypeErrorCode)
		if haz {
			field := slices.Find(herr.FieldErrors, func(ferr *errors.FieldError) bool {
				return ferr != nil && ferr.Rule == "invalidPayloadType"
			})
			innerlog = innerlog.With(field.Field, field.Param)
		}

		innerlog.Error("error deserializing message")
		c.requeueOrDelete(
			ctx,
			log,
			errors.Wrap(err, "error deserializing message", errors.NoRetryErrorCode),
			record,
		)
		return nil, err
	}

	if !c.allowedScope(record) {
		log.With("kafka_record", record).With("error", err).Error("scope is invalid")
		err = errors.New("invalid scope", errors.InvalidScopeRequeueErrorCode)
		c.requeueOrDelete(ctx, log, err, record)
		return nil, err
	}

	eventWrapper := events.NewUntypedEventWrapperFromMessage(wrapper)
	eventWrapper.UntypedMessage.SetHeader("kafka.record", record)
	eventWrapper.UntypedMessage.SetHeader("kafka.offset", record.Offset)
	eventWrapper.UntypedMessage.SetHeader("kafka.partition", record.Partition)
	eventWrapper.UntypedMessage.SetHeader("kafka.topic", record.Topic)

	// Add retry count from headers if present
	retryCount := c.getRetryCount(record.Headers)
	eventWrapper.UntypedMessage.SetHeader("kafka.retry_count", retryCount)

	return &eventWrapper, nil
}

func (c *Consumer) commitRecord(ctx context.Context, record *kgo.Record) error {
	pixiecontext.GetCtxLogger(ctx).
		With("topic", record.Topic).
		With("partition", record.Partition).
		With("offset", record.Offset).Log("committing offset")

	return c.client.kgoClient.CommitRecords(ctx, record)
}

func (c *Consumer) requeueOrDelete(ctx context.Context, log logger.Interface, err error, record *kgo.Record) {

	var retryUntil = c.getRetryDeadline(record.Headers)
	var now = time.Now().UnixMilli()

	if retryUntil > 0 && now > retryUntil {
		base64Message := base64.StdEncoding.EncodeToString(record.Value)
		log.With("topic", record.Topic).
			With("partition", record.Partition).
			With("offset", record.Offset).
			With("message_base64", base64Message).
			Error("Message retry deadline exceeded, dropping message (deserialization failed)")

		commitErr := c.commitRecord(ctx, record)
		if commitErr != nil {
			log.With("error", commitErr).Error("error committing message after deadline exceeded %s", record.Offset)
		}
		return
	}

	_, hasScopeCode := errors.Has(err, errors.InvalidScopeRequeueErrorCode)
	_, has := errors.Has(err, errors.ProcessFailedDoNotRequeueErrorCode)
	_, haz := errors.Has(err, errors.NoRetryErrorCode)
	_, hasError := errors.Has(err, errors.ProcessingEventErrorCode)

	retryCount := c.getRetryCount(record.Headers)

	log.Log(
		"evaluating if it's for deletion. scopeInvalid: %t ; processFailedNotRequeue: %t ; nonRetriableError: %t; retryCount: %d ; maxRetries: %d",
		hasScopeCode,
		has,
		haz,
		retryCount,
		c.cfg.RequeueMaxRetries,
	)

	if (!haz || !has || hasScopeCode || hasError) && retryCount <= c.cfg.RequeueMaxRetries {
		log.Debug("executing requeue")
		err = c.requeue(ctx, record, retryCount)
		if err == nil {
			log.Debug("left uncommitted, requeue succeeded.")
			return
		}

		log.With("error", err).Error("error requeue-ing message %s", record.Offset)
	}

	log.Debug("executing deletion (commit)")
	err = c.commitRecord(ctx, record)
	if err != nil {
		log.With("error", err).Error("error committing message %s", record.Offset)
		return
	}
}

func (c *Consumer) getRetryDeadline(headers []kgo.RecordHeader) int64 {
	for _, header := range headers {
		if header.Key == XRetryUntilHeader {
			parsedValue, parseErr := strconv.ParseInt(string(header.Value), 10, 64)
			if parseErr == nil {
				return parsedValue
			}

			break
		}
	}

	return 0
}

func (c *Consumer) requeue(ctx context.Context, record *kgo.Record, currentRetryCount int) error {
	if c.retryManager != nil {
		return c.retryManager.SendToRetry(ctx, record, currentRetryCount, record.Topic)
	}

	// If no retry manager, just not commit the message, consumer will have it again
	return nil
}

func (c *Consumer) getRetryCount(headers []kgo.RecordHeader) int {
	for _, header := range headers {
		if header.Key == XRetryCountHeader {
			count, err := strconv.Atoi(string(header.Value))
			if err != nil {
				return 0
			}
			return count
		}
	}
	return 0
}

// SetRetryManager allows setting the retry manager after consumer creation
// This is needed because the retry manager requires a producer
func (c *Consumer) SetRetryManager(retryManager *RetryManager) {
	c.retryManager = retryManager
}

// Close closes the consumer
func (c *Consumer) Close() {
	if c.client != nil && c.client.kgoClient != nil {
		c.client.kgoClient.Close()
	}
}
