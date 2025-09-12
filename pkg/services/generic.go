package services

import (
	"github.com/pixie-sh/database-helpers-go/database"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/caller"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/types"
)

type Service[T any] struct {
	Instance    *T
	newInstance func(Service[T]) (*T, *Service[T])

	Tx *database.DB
}

func NewService[T any](instance *T, newInstance func(Service[T]) (*T, *Service[T])) Service[T] {
	svc := Service[T]{
		Instance: instance,
		Tx:       nil,
	}

	svc.newInstance = newInstance
	return svc
}

// WithTx creates a copy of current service, setting with txDB *DB as service.tx
func (svc Service[T]) WithTx(txDB *database.DB) T {
	if types.Nil(svc.newInstance) {
		logger.Logger.With("four_steps_caller", caller.NewCaller(caller.FourHopsCallerDepth)).Error("wrong call to WithTx on Service[T]")
		panic(errors.New("wrong call to WithTx on Service[%s]", types.NameOf(svc.Instance)))
	}

	i, s := svc.newInstance(svc)
	s.Tx = txDB
	return *i
}

func (svc Service[T]) TxNil() bool {
	return types.Nil(svc.Tx)
}
