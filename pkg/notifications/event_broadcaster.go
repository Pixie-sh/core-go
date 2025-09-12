package notifications

import (
	"context"
	"time"

	"github.com/pixie-sh/logger-go/logger"

	events2 "github.com/pixie-sh/core-go/infra/events"
	"github.com/pixie-sh/core-go/infra/message_router"
	"github.com/pixie-sh/core-go/infra/message_wrapper"
	"github.com/pixie-sh/core-go/pkg/types"
)

// EventBroadcaster is a custom implementation that replaces the deprecated notifications.Service
// It implements the message_router.Broadcaster interface and uses events.Emit for broadcasting
type EventBroadcaster struct {
}

// NewEventBroadcaster creates a new event broadcaster instance
func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{}
}

// Broadcast implements the message_router.Broadcaster interface
// It processes messages and emits them as events using the events.Emit function
func (eb *EventBroadcaster) Broadcast(broadcastID string, identifier string, wrappers ...message_wrapper.UntypedMessage) message_router.BroadcastResult {
	log := logger.Logger.With("broadcast_id", broadcastID).With("identifier", identifier)

	ctx := context.Background()

	for _, message := range wrappers {
		log := log.With("message_id", message.ID).With("payload_type", message.PayloadType)

		// Create an event wrapper from the message
		eventWrapper := events2.NewUntypedEventWrapper(
			message.ID,
			broadcastID,
			time.Now(),
			message.PayloadType,
			message.Payload,
		)

		// Set the destination based on the identifier
		if len(eventWrapper.To) == 0 {
			eventWrapper.To = []string{identifier}
		}

		// Set any headers from the original message
		for key, value := range message.Headers {
			eventWrapper.UntypedMessage.SetHeader(key, value)
		}
		// Emit the event
		err := events2.Emit(ctx, eventWrapper)
		if err != nil {
			log.With("error", err).Error("failed to emit event for broadcast")
			continue
		}

		log.Debug("successfully emitted event for broadcast")
	}

	return message_router.BroadcastResult{}
}

// BroadcastCtx implements the message_router.Broadcaster interface for context-based broadcasting
func (eb *EventBroadcaster) BroadcastCtx(ctx *message_router.BroadcastContext) []message_router.BroadcastResult {
	if types.Nil(ctx) {
		logger.Logger.Error("BroadcastCtx called with nil context")
		return nil
	}

	log := logger.Logger.With("broadcast_id", ctx.BroadcastID)
	results := make([]message_router.BroadcastResult, 0, len(ctx.GetChannelKeys()))

	for _, key := range ctx.GetChannelKeys() {
		channel := ctx.GetChannel(key)
		if channel == nil {
			log.With("channel_key", key).Warn("got nil channel for key")
			continue
		}

		result := eb.Broadcast(ctx.BroadcastID, channel.ChannelIdentifier, channel.Messages...)
		results = append(results, result)
	}

	return results
}

// BroadcastFinalizer implements the message_router.Broadcaster interface
func (eb *EventBroadcaster) BroadcastFinalizer(results ...message_router.BroadcastResult) error {
	logger.Logger.With("results_count", len(results)).Debug("EventBroadcaster finalizer called")
	return nil
}
