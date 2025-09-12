package metrics

import (
	"context"
	"runtime/debug"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"

	"go.opentelemetry.io/otel/trace"

	"github.com/pixie-sh/logger-go/env"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var GlobalRegistry Registry
var GlobalTracer Tracer

func init() {
	GlobalRegistry = Registry{prometheus.NewRegistry()}
	GlobalRegistry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{
				Namespace:    env.EnvAppName(),
				ReportErrors: true,
			},
		),
	)
}

func SetTraceProvider(t Tracer) error {
	otel.SetTracerProvider(t.TracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(t.Propagator))

	GlobalTracer = t
	return nil
}

var Defer = func(ctx context.Context, tracers ...Tracer) {
	if len(tracers) == 0 {
		tracers = append(tracers, GlobalTracer)
	}

	for _, tracer := range tracers {
		err := tracer.Shutdown(ctx)
		if err != nil {
			pixiecontext.GetCtxLogger(ctx).
				With("stack_trace", debug.Stack()).
				Error("Failed GlobalTracer defer: %v", err)
		}
	}
}

func Span(ctx context.Context) TracerSpan {
	span := trace.SpanFromContext(ctx)
	return TracerSpan{span}
}
