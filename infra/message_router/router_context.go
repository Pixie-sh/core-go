package message_router

import (
	"context"

	"github.com/pixie-sh/core-go/infra/events"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/message_wrapper"
)

type RouterContext struct {
	context.Context

	Request *message_wrapper.UntypedMessage
	Values  map[string]interface{}

	//if Router.Handle functions is used
	//the responses need to be handled explicitly on the registered handlers
	Responses []message_wrapper.UntypedMessage

	//if NewNakedRouter is used the BroadcastCtx is not used default
	BroadcastCtx *BroadcastContext

	//if NewNakedRouter is used the NotificationsCtx is not used by default
	NotificationsCtx *BroadcastContext

	// if Router.Handle is used the Connection can be nil
	// use utils.Nil to check it
	Connection SourceConnection

	// if Error is set, a SSA is returned with that same error.
	// no response nor broadcast are sent
	Error  errors.E
	Logger logger.Interface
}

func NewRouterContext(ctx context.Context) *RouterContext {
	return &RouterContext{
		Context:          ctx,
		Values:           make(map[string]interface{}),
		Responses:        make([]message_wrapper.UntypedMessage, 0),
		BroadcastCtx:     NewBroadcastContext(),
		NotificationsCtx: NewBroadcastContext(),
		Logger:           logger.Logger,
		Error:            nil,
	}
}

func (rc *RouterContext) WithRequest(request message_wrapper.UntypedMessage) *RouterContext {
	rc.Request = &request
	return rc
}

func (rc *RouterContext) WithResponse(responses ...message_wrapper.UntypedMessage) *RouterContext {
	rc.Responses = append(rc.Responses, responses...)
	return rc
}

// Broadcast sends the given messages to the specified channel in the BroadcastContext.
// It adds the messages to the channel using the AddMessages method of BroadcastChannel.
// Returns the receiver RouterContext for method chaining.
func (rc *RouterContext) Broadcast(channelID string, messages ...message_wrapper.UntypedMessage) *BroadcastChannel {
	bCtx := rc.BroadcastCtx.GetChannel(channelID)
	bCtx.AddMessages(messages...)
	return bCtx
}

func (rc *RouterContext) WithConnection(connection SourceConnection) *RouterContext {
	rc.Connection = connection
	return rc
}

// Notify uses the notification broadcaster push notification to service destinations;
// it leverages the same logic as Broadcast
func (rc *RouterContext) Notify(notifications []events.UntypedEventWrapper) *RouterContext {
	for _, notification := range notifications {
		if len(notification.To) == 0 {
			logger.Logger.
				With("notification", notification).
				Warn("notification doesn't have destination channel. ignored")

			continue
		}

		for _, to := range notification.To {
			rc.NotificationsCtx.GetChannel(to).
				AddMessages(notification.UntypedMessage)
		}
	}
	return rc
}

func (rc *RouterContext) WithLogger(log logger.Interface) *RouterContext {
	rc.Logger = log
	rc.Context = pixiecontext.SetCtxLogger(rc.Context, log)
	return rc
}
