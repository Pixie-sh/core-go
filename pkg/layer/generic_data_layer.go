package layer

import (
	"context"

	"github.com/pixie-sh/database-helpers-go/database"
	"github.com/pixie-sh/logger-go/logger"
)

type GenericDataLayer struct {
	orm *database.Orm
}

func NewGenericDataLayer(orm *database.Orm) GenericDataLayer {
	return GenericDataLayer{orm: orm}
}

func (l GenericDataLayer) InitLayer(_ context.Context) error {
	logger.Debug("empty init layer for ChatDataLayer")
	return nil
}

func (l GenericDataLayer) Health(_ context.Context) HealthResponse {
	return HealthResponse{
		Err: l.orm.Ping(),
	}
}

func (l GenericDataLayer) Defer() {
	err := l.orm.Close()
	if err != nil {
		logger.With("error", err).Error("unable to close ORM connection")
	}
}

func (l GenericDataLayer) Connection(_ context.Context) *database.DB {
	return l.orm.DB
}

func (l GenericDataLayer) Transaction(ctx context.Context, handler func(db *database.DB) error, opts ...*database.TxOptions) (err error) {
	return database.NewRepository(l.Connection(ctx), func(db *database.DB) any { panic("wrong GenericDataLayer usage of transaction") }).Transaction(handler, opts...)
}
