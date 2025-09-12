package microservice

import (
	"context"

	"github.com/pixie-sh/core-go/pkg/comm/http"
)

func SetupHealthController(_ context.Context, server http.Server) error {
	server.Group("health").Get("/", ok)
	return nil
}

func ok(ctx http.ServerCtx) error {
	return http.Response(ctx, 200, "Ok")
}
