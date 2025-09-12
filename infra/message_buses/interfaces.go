package message_buses

import (
	"context"

	"github.com/pixie-sh/core-go/infra/message_wrapper"

	"github.com/pixie-sh/core-go/pkg/pubsub"
)

type MessageBus interface {
	Subscribe(pubsub.Subscriber[message_wrapper.UntypedMessage]) string
	Unsubscribe(subID string)
	Publish(fromID string, messages ...message_wrapper.UntypedMessage)

	countSubscriptions() int
}

type BusPool interface {
	Get(ctx context.Context, key string) MessageBus
}
