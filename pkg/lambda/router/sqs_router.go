package router

import (
	"context"
	"fmt"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/lambda/lambda_sqs"

	"github.com/pixie-sh/core-go/pkg/types"

	awsEvents "github.com/aws/aws-lambda-go/events"
)

const FallbackQueue = "?"      // as fallback
const EventTypesWildcard = "*" // always

// SQSRouteKey is a struct to hold the combination of QueueName and EventType.
type SQSRouteKey struct {
	QueueName string
	EventType types.PayloadType
}

func (k SQSRouteKey) Key() string {
	return fmt.Sprintf("%s:%s", k.QueueName, k.EventType)
}

// SQSHandler is a type for handling SQS requests.
type SQSHandler func(*pixiecontext.SQSContext) (awsEvents.SQSEventResponse, error)

type SQSQueue struct {
	gates     []SQSHandler
	queueName string
	sqsRouter *SQSRouter
}

// SQSRouter is a concrete router for SQS routes.
type SQSRouter struct {
	*GenericRouter[SQSRouteKey, SQSHandler]
}

func NewSQSRouter(ctx context.Context, routePrefix string) *SQSRouter {
	return &SQSRouter{
		GenericRouter: NewGenericRouter[SQSRouteKey, SQSHandler](ctx, routePrefix),
	}
}

func (r *SQSRouter) SQSQueue(
	queueName string,
	handler ...SQSHandler) *SQSQueue {
	return &SQSQueue{
		gates:     append([]SQSHandler{}, handler...),
		queueName: queueName,
		sqsRouter: r,
	}
}

// RegisterHandler registers a handler for a specific method and path.
func (r *SQSQueue) RegisterHandler(
	_ context.Context,
	eventType types.PayloadType,
	handler ...SQSHandler,
) *SQSQueue {
	key := SQSRouteKey{
		QueueName: r.queueName,
		EventType: eventType,
	}

	if r.sqsRouter.routes == nil {
		r.sqsRouter.routes = make(map[string][]SQSHandler)
	}

	for _, handle := range handler {
		r.sqsRouter.routes[key.Key()] = append(r.sqsRouter.routes[key.Key()], handle)
	}

	return r
}

// HandleSQSMessage for handling SQS requests.
func (router *SQSRouter) HandleSQSMessage(ctx context.Context, queueName string, msgs []awsEvents.SQSMessage) (awsEvents.SQSEventResponse, error) {

	eventKey := SQSRouteKey{
		QueueName: queueName,
		EventType: EventTypesWildcard,
	}
	hasWildCard := router.checkWithFallbackQueue(ctx, &eventKey)

	log := pixiecontext.GetCtxLogger(ctx)

	// Group keys per payloadType
	keys := make(map[string][]awsEvents.SQSMessage, len(msgs))
	for _, msg := range msgs {

		// TODO: Save this as a variable, it's hardcoded in a lot of places
		payloadType, ok := msg.MessageAttributes["x-payload-type"]
		if !ok {
			log.With("queue_name", queueName).Warn("This sqs message as no payload type, skipping...")
			continue
		}

		if payloadType.StringValue == nil {
			log.With("queue_name", queueName).Warn("SQS message had payloadType but was empty? skipping...")
			continue
		}

		key := SQSRouteKey{
			QueueName: queueName,
			EventType: types.PayloadType(*payloadType.StringValue),
		}

		if hasWildCard { // If it has wildcard, all routes should be valid
			key.QueueName = eventKey.QueueName // Already knows queue fallback
			router.routes[key.Key()] = router.routes[eventKey.Key()]
		}

		if !router.checkWithFallbackQueue(ctx, &key) { // If it has a wildcard in theory this check is redundant
			continue
		}

		keys[key.Key()] = append(keys[key.Key()], msg)

	}

	batchItemFailures := make([]awsEvents.SQSBatchItemFailure, 0)
	for key, value := range keys {

		sqsCtx := &pixiecontext.SQSContext{
			GenericContext: pixiecontext.NewGenericContext(ctx),
			Event:          awsEvents.SQSEvent{Records: value}, // Only the records that have the same
		}

		for _, handler := range router.routes[key] {
			resp, err := handler(sqsCtx)

			if len(resp.BatchItemFailures) != 0 {
				batchItemFailures = append(batchItemFailures, resp.BatchItemFailures...)
			}

			if err != nil {
				log.
					With("error", err).
					With("batch_item_failures", resp.BatchItemFailures).
					Error("Error processing batch")
			}

		}

	}

	return lambda_sqs.Response(batchItemFailures)
}

// checkWithFallbackQueue checks if there's a valid route for the specified key
// it will also check with fallback and change approperly if so
func (router *SQSRouter) checkWithFallbackQueue(_ context.Context, key *SQSRouteKey) bool {

	if len(router.routes[key.Key()]) != 0 {
		return true
	}

	key.QueueName = FallbackQueue // Check for fallback queue
	return len(router.routes[key.Key()]) != 0
}
