package message_router

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/message_wrapper"
	"github.com/pixie-sh/core-go/pkg/pubsub"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/uid"
)

// innerSub implements pubsub.Subscriber[T any]
// make subscription sync to avoid mis behave on connection data input
type innerSub struct {
	connection SourceConnection
	router     *Router
}

func (sub *innerSub) Publish(msg message_wrapper.UntypedMessage) {
	sub.router.listen(sub.connection, msg)
}

func (sub *innerSub) ID() string {
	return fmt.Sprintf("router-subscriber-%s", sub.connection.ID())
}

type Router struct {
	Broadcaster

	appCtx                 context.Context
	createSSA              func(wrapper message_wrapper.UntypedMessage, err errors.E) message_wrapper.UntypedMessage
	notificationsBroadcast Broadcaster
	subscriber             *pubsub.OnProcessSubscriber[SourceSubscription]

	mu       sync.RWMutex
	handlers map[string]struct {
		handlers []MessageHandler
	}
}

func NewRouter(
	ctx context.Context,
	wsBroadcast Broadcaster,
	notificationsBroadcast Broadcaster,
	generateAck func(wrapper message_wrapper.UntypedMessage, err errors.E) message_wrapper.UntypedMessage) *Router {
	cr := Router{
		Broadcaster:            wsBroadcast,
		appCtx:                 ctx,
		createSSA:              generateAck,
		notificationsBroadcast: notificationsBroadcast,
		handlers: make(map[string]struct {
			handlers []MessageHandler
		}),
	}

	cr.subscriber = pubsub.NewOnProcessSubscriber[SourceSubscription](ctx, uid.NewUUID(), cr.subscriberHandler, 256)
	return &cr
}

type broadcaster struct {
}

func (b broadcaster) Broadcast(broadcastID string, identifier string, message ...message_wrapper.UntypedMessage) BroadcastResult {
	logger.Logger.Warn("naked broadcast being called, nothing is done")
	return BroadcastResult{}
}

func (b broadcaster) BroadcastFinalizer(result ...BroadcastResult) error {
	logger.Logger.Warn("naked broadcast finalizer being called, nothing is done")
	return nil
}

func (b broadcaster) BroadcastCtx(ctx *BroadcastContext) []BroadcastResult {
	logger.Logger.Warn("naked broadcast with ctx being called, nothing is done")
	return []BroadcastResult{}
}

func NewNakedRouter(ctx context.Context) *Router {
	cr := NewRouter(ctx, &broadcaster{}, &broadcaster{}, func(wrapper message_wrapper.UntypedMessage, err errors.E) message_wrapper.UntypedMessage {
		return wrapper
	})

	return cr
}

func (r *Router) Register(messageType string, handlers ...MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[messageType] = struct {
		handlers []MessageHandler
	}{
		handlers,
	}
}

func (r *Router) Unregister(messageType string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.handlers[messageType]
	if ok {
		delete(r.handlers, messageType)
	}
}

func (r *Router) SourceSubscriber() pubsub.Subscriber[SourceSubscription] {
	return r.subscriber
}

func (r *Router) subscriberHandler(src SourceSubscription) {
	var (
		connection = src.Connection
		added      = src.Added
	)

	if types.Nil(connection) {
		logger.Logger.Error("subscriber function at Router called with nil connection")
		return
	}

	if !added {
		logger.Logger.Debug("router connection %s removed", connection.ID())
		return
	}

	connection.Subscribe(&innerSub{
		connection: connection,
		router:     r,
	})
}

func (r *Router) routing(ctx *RouterContext) {
	request := ctx.Request

	r.mu.RLock()
	defer r.mu.RUnlock()

	// get payload type handlers
	handlers, ok := r.handlers[request.PayloadType]
	if !ok {
		handlers, ok = r.handlers[types.PayloadTypeFallback]
		if !ok {
			ssa := r.createSSA(request, errors.New("no handlers provided").WithErrorCode(errors.ErrorPerformingRequestErrorCode))

			logger.Logger.
				With("request", request).
				With("response", ssa).
				Error("message_router no handlers provided for %s (%s)", request.ID, request.PayloadType)

			ctx.Responses = []message_wrapper.UntypedMessage{ssa}
			return
		}
	}

	if len(handlers.handlers) > 0 {
		r.iterateHandlers(ctx, handlers.handlers)
	}

	if ctx.Error != nil {
		ssa := r.createSSA(request, ctx.Error)
		logger.Logger.
			With("request", request).
			With("error", ctx.Error).
			With("response", ssa).
			Error("message_router error processing %s (%s)", request.ID, request.PayloadType)

		ctx.Responses = []message_wrapper.UntypedMessage{ssa}
		return
	}

	//prepend ssa
	ctx.Responses = append([]message_wrapper.UntypedMessage{r.createSSA(request, nil)}, ctx.Responses...)
}

func (r *Router) iterateHandlers(rc *RouterContext, handlers []MessageHandler) {
	for _, h := range handlers {
		h(rc)
	}
}

func (r *Router) processResponses(connection SourceConnection, responses []message_wrapper.UntypedMessage) {
	if !types.Nil(connection) {
		for _, wrapper := range responses {
			connection.Publish(wrapper)
		}
	}
}

func (r *Router) listen(connection SourceConnection, request message_wrapper.UntypedMessage) {
	log := logger.Logger.With(logger.TraceID, request.ID)

	defer func() {
		if r := recover(); r != nil {
			log.
				With("stack_trace", debug.Stack()).
				With("recover", r).
				Error("recovered from panic at message router for %s", request.PayloadType)
		}
	}()

	routerContext := NewRouterContext(context.Background()).
		WithRequest(request).
		WithConnection(connection).
		WithLogger(log)

	r.routing(routerContext)
	r.processResponses(connection, routerContext.Responses)
	if routerContext.Error == nil {
		if len(routerContext.BroadcastCtx.messagePerChannel) > 0 {
			routerContext.BroadcastCtx.BroadcastID = connection.ID()
			go r.Broadcaster.BroadcastCtx(routerContext.BroadcastCtx)
		}

		if len(routerContext.NotificationsCtx.messagePerChannel) > 0 {
			routerContext.NotificationsCtx.BroadcastID = connection.ID()
			go r.notificationsBroadcast.BroadcastCtx(routerContext.NotificationsCtx)
		}
	}
}

// Handle the routing logic for SourceConnection do not apply here
// the broadcast are not called per si nor a SSA generated
// usually used with NewNakedRouter constructor
func (r *Router) Handle(ctx context.Context, message message_wrapper.UntypedMessage) error {
	log := pixiecontext.GetCtxLogger(ctx).With("message_id", message.ID).With("message", message)

	defer func() {
		if r := recover(); r != nil {
			log.
				With("stack_trace", debug.Stack()).
				With("recover", r).
				Error("recovered from panic at message router for %s", message.PayloadType)
		}
	}()

	routerContext := NewRouterContext(ctx).
		WithRequest(message).
		WithLogger(log)

	r.mu.RLock()
	defer r.mu.RUnlock()

	// get payload type handlers
	handlers, ok := r.handlers[routerContext.Request.PayloadType]
	if !ok {
		handlers, ok = r.handlers[types.PayloadTypeFallback]
		if !ok {
			log.Error("message_router no handlers provided for %s (%s)", routerContext.Request.ID, routerContext.Request.PayloadType)
			return errors.New("no handlers provided").WithErrorCode(errors.ErrorPerformingRequestErrorCode)
		}
	}

	if len(handlers.handlers) > 0 {
		r.iterateHandlers(routerContext, handlers.handlers)
	}

	if routerContext.Error != nil {
		log.
			With("error", routerContext.Error).
			Error("message_router error processing %s (%s)", routerContext.Request.ID, routerContext.Request.PayloadType)

		return routerContext.Error
	}

	return nil
}
