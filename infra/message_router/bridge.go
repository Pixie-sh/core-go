package message_router

import (
	"context"

	"github.com/pixie-sh/core-go/pkg/types"

	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/pubsub"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type BridgeSource interface {
}

type Bridge struct {
	publisher  *pubsub.GenericPublisher[SourceSubscription]
	subscriber *pubsub.OnProcessSubscriber[SourceSubscription]
	source     BridgeSource
}

func NewBridge(ctx context.Context, source BridgeSource) *Bridge {
	b := &Bridge{
		publisher: pubsub.NewGenericPublisher[SourceSubscription](ctx),
		source:    source,
	}

	b.subscriber = pubsub.NewOnProcessSubscriber[SourceSubscription](ctx, uid.NewUUID(), b.subscriberHandler, 512)
	return b
}

// Listen blocking call that listens to BridgeSource events
func (b *Bridge) Listen(ctx context.Context) {
	// TODO: woo
	// b.publisher.NotifySubscriptions(...sdf)
}

// Subscribe implements Publisher interface
func (b *Bridge) Subscribe(sub pubsub.Subscriber[SourceSubscription]) {
	b.publisher.Subscribe(sub)
}

// SourceSubscriber returns the subscriber for Source Subscriptions
func (b *Bridge) SourceSubscriber() pubsub.Subscriber[SourceSubscription] {
	return b.subscriber
}

// subscriberHandler listener function for on process sources
// Added: sources will send that information over the bridge
// !Added: means it removed
func (b *Bridge) subscriberHandler(src SourceSubscription) {
	if types.Nil(src.Connection) {
		logger.Logger.Warn("nil connection at subscriber")
		return
	}

	if src.Added {
		// coiso
	}

	if !src.Added {
		// cenas
	}

	return
}
