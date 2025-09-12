package microservice

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/pixie-sh/errors-go"
	loggerEnv "github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/version"
	"github.com/pixie-sh/core-go/pkg/env"
	"github.com/pixie-sh/core-go/pkg/types/io"
)

var cancelCtx context.CancelFunc
var cancelChannel chan os.Signal

var chainAfterCancel []func()

func Setup(bg context.Context) (context.Context, error) {
	var ctx context.Context

	ctx, cancelCtx = context.WithCancel(bg)
	cancelChannel = make(chan os.Signal, 1)
	signal.Notify(cancelChannel, syscall.SIGINT, syscall.SIGTERM)

	newLogger, err := logger.NewLogger(
		ctx,
		io.NewAsyncWriter(ctx, os.Stdout, -1),
		loggerEnv.EnvAppName(),
		loggerEnv.EnvScope(),
		fmt.Sprintf("%s-%s", version.Version, version.Commit),
		env.LogLevel(),
		[]string{logger.TraceID},
	)

	if err != nil {
		return ctx, errors.NewWithError(err, "unable to init logger").WithErrorCode(errors.ErrorCreatingDependencyErrorCode)
	}

	logger.Logger = newLogger
	newLogger.Log("ms '%s' setup completed", loggerEnv.EnvAppName())

	return ctx, nil
}

// WithCancelFunctions functions to be called when the ms is terminating
// they will be executed in the reverse order, similar to how defer works on Go
func WithCancelFunctions(funcs ...func()) {
	chainAfterCancel = append(chainAfterCancel, funcs...)
}

func WaitSignalAndCancel() {
	<-cancelChannel

	cancelCtx()
	for i := len(chainAfterCancel) - 1; i >= 0; i-- {
		chainAfterCancel[i]()
	}
}

func PanicHandler() {
	if r := recover(); r != nil {
		logger.With("stack_trace", debug.Stack()).Error("recovered from panic: %v", r)
		cancelChannel <- syscall.SIGINT
	}
}

func StartAsync(ctx context.Context, ms Starter) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				break
			default:
				err := ms.Start()
				if err != nil {
					panic(err)
				}
			}
		}
	}()
}
