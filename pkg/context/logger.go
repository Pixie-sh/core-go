package context

import (
	"context"

	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/uid"
)

const LocalsRequestLogger = "request_logger" //refer models.LocalsRequestLogger

func GetCtxLogger(ctx context.Context, noClone ...bool) logger.Interface {
	if ctx == nil {
		return logger.Logger
	}

	log, ok := ctx.Value(LocalsRequestLogger).(logger.Interface)
	if !ok {
		return logger.Logger
	}

	if len(noClone) > 0 && noClone[0] {
		return log
	}

	return log.Clone()
}

func SetCtxLogger(ctx context.Context, log logger.Interface) context.Context {
	return context.WithValue(ctx, LocalsRequestLogger, log.WithCtx(ctx))
}

func SetCtxTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, logger.TraceID, traceID)
}

func GetCtxTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(logger.TraceID).(string)
	if !ok {
		return uid.NewULID()
	}

	return traceID
}
