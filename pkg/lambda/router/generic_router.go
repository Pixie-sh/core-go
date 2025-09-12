package router

import (
	"context"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
)

// GenericHandlerType is an interface for all handler types.
type GenericHandlerType interface {
	APIHandler | SQSHandler //| KafkaHandler //TODO TBD later
}

// GenericHandler is a type for handling generic requests.
type GenericHandler func(*pixiecontext.GenericContext) error

func (h GenericHandler) Handle(ctx *pixiecontext.GenericContext) error {
	return h(ctx)
}

// GenericRouteKey is a struct to which will hold the handler routing combination
type GenericRouteKey interface {
	Key() string
}

// GenericRouter is a struct that holds the routing logic.
type GenericRouter[T GenericRouteKey, H GenericHandlerType] struct {
	routes      map[string][]H
	gates       []H
	routePrefix string
}

// NewGenericRouter creates a new GenericRouter instance.
// routePrefix default is "" not used for now
func NewGenericRouter[T GenericRouteKey, H GenericHandlerType](_ context.Context, routePrefix string) *GenericRouter[T, H] {
	return &GenericRouter[T, H]{
		routes:      make(map[string][]H),
		gates:       nil,
		routePrefix: routePrefix,
	}
}

func (router *GenericRouter[K, H]) PreAllocate(expectedRoutes int) *GenericRouter[K, H] {
	router.routes = make(map[string][]H, expectedRoutes)
	return router
}

func (router *GenericRouter[K, H]) Middlewares(gates ...H) *GenericRouter[K, H] {
	if len(router.gates) > 0 {
		router.gates = append(router.gates, gates...)
	} else {
		router.gates = gates
	}
	return router
}
