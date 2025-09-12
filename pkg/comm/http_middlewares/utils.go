package http_middlewares

import (
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/comm/http"
	pixicontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type requestUtils struct {
	logger logger.Interface
}

func RequestUtils(logger logger.Interface) http.ServerHandler {
	return requestUtils{logger}.Handler
}

const nullTraceVal = "00000000000000000000000000000000"

func (l requestUtils) Handler(ctx http.ServerCtx) error {
	var traceID = uid.NewULID()
	var metricsTraceID = ctx.Locals(http.LocalsMetricsTraceID)
	if metricsTraceID != nil {
		traceVal, ok := metricsTraceID.(string)
		if ok && traceVal != nullTraceVal {
			traceID = traceVal
		}
	}

	log := l.logger.Clone().
		WithCtx(ctx.Context()).
		With(http.LocalsTraceID, traceID)

	ctx.Locals(http.LocalsTraceID, traceID)
	ctx.Locals(http.LocalsRequestLogger, log)

	userCtx := pixicontext.SetCtxLogger(ctx.UserContext(), log)
	userCtx = pixicontext.SetCtxTraceID(userCtx, traceID)
	ctx.SetUserContext(userCtx)

	defer func() {
		ctx.Response().Header.Set(http.XRequestIDKey, traceID)
		ctx.Response().Header.Set("Access-Control-Expose-Headers", http.XRequestIDKey)
	}()

	return ctx.Next()
}
