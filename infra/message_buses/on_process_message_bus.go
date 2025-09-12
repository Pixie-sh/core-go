package message_buses

import (
	"context"

	"github.com/pixie-sh/core-go/infra/message_wrapper"

	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/pubsub"
)

type OnProcessMessageBus struct {
	publisher *pubsub.GenericPublisher[message_wrapper.UntypedMessage]
}

func NewOnProcessMessageBus(ctx context.Context) *OnProcessMessageBus {
	return &OnProcessMessageBus{
		publisher: pubsub.NewGenericPublisher[message_wrapper.UntypedMessage](ctx),
	}
}

func (bus *OnProcessMessageBus) countSubscriptions() int {
	return bus.publisher.CountSubscriptions()
}

func (bus *OnProcessMessageBus) Subscribe(sub pubsub.Subscriber[message_wrapper.UntypedMessage]) string {
	bus.publisher.Subscribe(sub)
	return sub.ID()
}

func (bus *OnProcessMessageBus) Unsubscribe(subID string) {
	bus.publisher.Unsubscribe(subID)
}

func (bus *OnProcessMessageBus) Publish(fromID string, messages ...message_wrapper.UntypedMessage) {
	if len(fromID) == 0 {
		logger.Logger.
			Debug("ignored publish from unknown publisher")

		return
	}

	for _, message := range messages {
		bus.publisher.NotifySubscriptions(message)
	}
}
