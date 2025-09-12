package layer

import (
	"context"

	"github.com/pixie-sh/database-helpers-go/database"
)

type HealthResponse struct {
	Err error
}

type Interface interface {
	InitLayer(ctx context.Context) error
	Health(ctx context.Context) HealthResponse
	Defer()
}

type DataInterface interface {
	Connection(ctx context.Context) *database.DB
	Transaction(ctx context.Context, f func(*database.DB) error, opts ...*database.TxOptions) error
}
