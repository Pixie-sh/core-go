package lambda

import (
	"context"
	"fmt"
	"os"

	"github.com/pixie-sh/errors-go"
	loggerEnv "github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/version"
	"github.com/pixie-sh/core-go/pkg/env"
)

func initLambda(ctx context.Context) error {
	newLogger, err := logger.NewLogger(
		ctx,
		os.Stdout, //io.NewAsyncWriter(ctx, os.Stdout, -1),
		loggerEnv.EnvAppName(),
		loggerEnv.EnvScope(),
		fmt.Sprintf("%s-%s", version.Version, version.Commit),
		env.LogLevel(),
		[]string{logger.TraceID},
	)

	if err != nil {
		return errors.Wrap(err, "lambda init failed", errors.LambdaInitFailedErrorCode)
	}

	logger.Logger = newLogger
	newLogger.Log("starting lambda")
	return nil
}
