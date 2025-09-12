package http_middlewares

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/pixie-sh/core-go/pkg/comm/http"
	"github.com/pixie-sh/core-go/pkg/metrics"
	"github.com/pixie-sh/logger-go/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func Tracing(skipPaths ...string) []http.ServerHandler {
	var skip map[string]struct{}
	if len(skipPaths) == 0 {
		skip = map[string]struct{}{}
		for _, path := range skipPaths {
			skip[path] = struct{}{}
		}
	}

	return []http.ServerHandler{
		otelfiber.Middleware(
			otelfiber.WithNext(func(ctx http.ServerCtx) bool {
				_, ok := skip[ctx.Path()]
				return !ok
			}),
			otelfiber.WithMeterProvider(otel.GetMeterProvider()),
		),
		func(ctx http.ServerCtx) error {
			propagator := otel.GetTextMapPropagator()
			tracer := otel.Tracer("http-server")
			carrier := propagation.HeaderCarrier{}

			ctx.Request().Header.VisitAll(func(key, value []byte) {
				carrier[string(key)] = []string{string(value)}
			})

			localCtx := propagator.Extract(ctx.Context(), carrier)
			localCtx, span := tracer.Start(localCtx, "http-request")
			defer span.End()

			traceCtx := trace.SpanContextFromContext(localCtx)
			propagator.Inject(localCtx, carrier)

			ctx.SetUserContext(localCtx)
			ctx.Locals(http.LocalsMetricsTraceID, traceCtx.TraceID().String())
			return ctx.Next()
		},
	}
}

var fiberProm *fiberprometheus.FiberPrometheus

func Metrics(registry metrics.Registry) http.ServerHandler {
	if fiberProm != nil {
		return fiberProm.Middleware
	}

	fiberProm = fiberprometheus.NewWithRegistry(
		registry,
		env.EnvAppName(),
		"http",
		"http-server",
		nil,
	)

	return fiberProm.Middleware
}
