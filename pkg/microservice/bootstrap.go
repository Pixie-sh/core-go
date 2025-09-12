package microservice

import (
	"context"

	"github.com/pixie-sh/di-go"
	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/configuration"
)

type Bootstrapper[T Starter, CT di.Configuration] struct {
	Ctx di.Context

	Starter       T
	Configuration CT
}

func BootstrapMust[T Starter, CT di.Configuration](ctx context.Context, token ...di.InjectionToken) Bootstrapper[T, CT] {
	bootstrap, err := Bootstrap[T, CT](ctx, token...)
	errors.Must(err)
	return bootstrap
}

func Bootstrap[T Starter, CT di.Configuration](ctx context.Context, token ...di.InjectionToken) (Bootstrapper[T, CT], error) {
	var ms T
	var cfg CT
	var err error
	var diCtx di.Context
	var options []func(opts *di.RegistryOpts)

	configuration.Setup(&cfg, true)

	ctx, err = Setup(ctx)
	if err != nil {
		return Bootstrapper[T, CT]{}, err
	}

	if len(token) > 0 {
		options = append(options, di.WithToken(token[0]))
	}

	diCtx = di.NewContext(ctx, cfg)
	ms, err = di.Create[T](diCtx, options...)

	return Bootstrapper[T, CT]{
		Ctx:           diCtx,
		Starter:       ms,
		Configuration: cfg,
	}, err
}

func (c Bootstrapper[T, CT]) Start() {
	defer c.Starter.PanicHandler()
	StartAsync(c.Ctx, c.Starter)
	WithCancelFunctions(c.Starter.Defer)
	WaitSignalAndCancel()
}
