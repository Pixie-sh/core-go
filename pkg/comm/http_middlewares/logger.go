package http_middlewares

import (
	"strings"
	"time"

	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/comm/http"
)

type loggerMiddleware struct {
	skipPaths map[string]struct{}
}

func Logger(skipPaths ...string) http.ServerHandler {
	var skip = map[string]struct{}{}
	if len(skipPaths) > 0 {
		skip = make(map[string]struct{}, len(skipPaths))
		for _, path := range skipPaths {
			skip[path] = struct{}{}
		}
	}

	return loggerMiddleware{skipPaths: skip}.Handler
}

// Handler log what we need from ctx
// use the ctx logger set on requestUtils mdw
func (l loggerMiddleware) Handler(ctx http.ServerCtx) error {
	if _, ok := l.skipPaths[ctx.Path()]; ok {
		return ctx.Next()
	}

	start := time.Now()
	reqLog := http.RequestLog{
		Method:  ctx.Method(),
		URL:     ctx.OriginalURL(),
		Headers: ctx.Request().Header.Header(),
		IP:      ctx.IP(),
	}

	if !strings.Contains(string(ctx.Request().Header.ContentType()), "multipart/form-data") {
		reqLog.Body = ctx.Body()
	}

	log := ctx.Locals(http.LocalsRequestLogger).(logger.Interface)
	log.With("request", reqLog)

	ctx.Locals("request_start", start)
	log.Log("request %s %s", reqLog.Method, reqLog.URL)

	err := ctx.Next()

	resLog := http.ResponseLog{
		Status:  ctx.Response().StatusCode(),
		Headers: ctx.Response().Header.Header(),
		Body:    ctx.Response().Body(),
	}

	if err != nil {
		resLog.Error = err
	}

	log.
		With("response", resLog).
		Log("request finished: %s; took %s", ctx.Path(), time.Since(start).String())
	return err
}
