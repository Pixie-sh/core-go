package metrics

import (
	"context"

	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

type ExporterConfiguration struct {
	CollectorURL string `json:"collector_url"`
	Secure       bool   `json:"secure"`
}

type Exporter struct {
	*otlptrace.Exporter
}

type TracerConfiguration struct {
	ServiceName string `json:"service_name"`
	SchemaURL   string `json:"schema_url"`
}

type TracerSpan struct {
	trace.Span
}

type Propagator = propagation.TextMapPropagator
type Tracer struct {
	*oteltrace.TracerProvider

	Propagator Propagator
}

func NewExporter(_ context.Context, conf ExporterConfiguration) (Exporter, error) {
	var secureOption otlptracegrpc.Option

	if conf.Secure {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(conf.CollectorURL),
		),
	)
	if err != nil {
		return Exporter{}, err
	}

	return Exporter{exporter}, nil
}

func NewTracer(ctx context.Context, exporter Exporter, conf TracerConfiguration) (Tracer, error) {
	if len(conf.SchemaURL) == 0 {
		conf.SchemaURL = semconv.SchemaURL
	}

	tp := oteltrace.NewTracerProvider(
		oteltrace.WithSampler(oteltrace.AlwaysSample()),
		oteltrace.WithBatcher(exporter),
		oteltrace.WithResource(
			resource.NewWithAttributes(
				conf.SchemaURL,
				semconv.ServiceNameKey.String(conf.ServiceName),
				semconv.TelemetrySDKLanguageGo,
			)),
	)

	return Tracer{tp, propagation.TraceContext{}}, nil
}
