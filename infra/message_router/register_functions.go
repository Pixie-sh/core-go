package message_router

import (
	"context"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/infra/message_wrapper"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/types"
)

func Register[M any](handler func(context.Context, *M) error, router *Router) {
	pt := types.PayloadTypeOf[M]()
	router.Register(pt.String(), func(ctx *RouterContext) {
		log := pixiecontext.GetCtxLogger(ctx).With("event_id", ctx.Request.ID)
		log.Log("processing untyped event '%s' of type '%s' on handler of '%s'", ctx.Request.ID, ctx.Request.PayloadType, pt.String())
		typedMessage := message_wrapper.MessageOf[M](ctx, *ctx.Request)
		typedData := typedMessage.Data()

		log.With("typed_message", typedData).Log("typed data gathered, stimulating handler")
		err := handler(ctx.Context, &typedData)
		if err != nil {
			log.With("error", err).Log("message processed with error")
			castedErr, ok := errors.As(err)
			if ok {
				ctx.Error = castedErr
			}

			if !ok {
				log.Error("error is not unknown, unable to process it")
			}
		}
	})
}

func RegisterFallback(handler func(context.Context, *message_wrapper.UntypedMessage) error, router *Router) {
	router.Register(types.PayloadTypeFallback, func(ctx *RouterContext) {
		log := pixiecontext.GetCtxLogger(ctx).With("event_id", ctx.Request.ID)
		log.Log("processing untyped event '%s' of type '%s' on handler of '%s'", ctx.Request.ID, ctx.Request.PayloadType, types.PayloadTypeFallback)
		err := handler(ctx.Context, ctx.Request)
		if err != nil {
			log.With("error", err).Log("message processed with error")
			castedErr, ok := errors.As(err)
			if ok {
				ctx.Error = castedErr
			}

			if !ok {
				log.Error("error is not unknown, unable to process it")
			}
		}
	})
}
