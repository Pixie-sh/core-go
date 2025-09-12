package microservice

import (
	"context"

	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pixie-sh/core-go/pkg/comm/http"
	"github.com/pixie-sh/core-go/pkg/metrics"
)

func SetupMetricsController(_ context.Context, server http.Server) error {
	server.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(
		metrics.GlobalRegistry,
		promhttp.HandlerOpts{Registry: metrics.GlobalRegistry},
	)))

	return nil
}
