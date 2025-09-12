package message_router

import (
	"context"

	"github.com/pixie-sh/core-go/infra/message_wrapper"

	"github.com/pixie-sh/core-go/pkg/pubsub"
)

type SourceSubscription struct {
	Connection SourceConnection
	Added      bool
}

type MessageHandler = func(ctx *RouterContext)

type SourceManager interface {
	Subscribe(listener func(src SourceSubscription))
}

type SourceConnection interface {
	Ctx() context.Context
	ID() string
	Locals(string) interface{}
	Subscribe(pubsub.Subscriber[message_wrapper.UntypedMessage])
	Publish(message_wrapper.UntypedMessage)
}

type SourceInformation struct {
	ChannelID string   // partyID: XXXXX
	RelatedTo []string // roomID: YYYY; roomID: UUUU; roomID: IIIII
}
