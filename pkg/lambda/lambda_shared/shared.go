package lambda_shared

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pixie-sh/database-helpers-go/database"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/version"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/lambda/lambda_api"
)

const XRequestIDKey = "X-Request-ID"

var lambdaSpecificOrm *database.Orm = nil
var chatSpecificOrm *database.Orm = nil

func orm(ctx context.Context, config database.Configuration, orm *database.Orm) (*database.Orm, error) {
	var err error
	log := pixiecontext.GetCtxLogger(ctx)
	if orm != nil {
		err = orm.Ping()
		if err != nil {
			log.With("error", err).
				Warn("shared orm is dead. creating a new instance")

			orm, err = database.FactoryInstance.Create(ctx, &config)
			if err != nil {
				log.With("error", err).
					Error("unable to create new shared orm instance")
				return nil, err
			}

			if orm.Error != nil {
				log.With("error", orm.Error).
					Error("unable to create new shared orm instance with orm.Error")
				return nil, orm.Error
			}
		}
	}

	if orm == nil {
		orm, err = database.FactoryInstance.Create(ctx, &config)
		if err != nil {
			log.With("error", err).
				Error("unable to create new shared orm instance")

			return nil, err
		}

		if orm.Error != nil {
			log.With("error", orm.Error).
				Error("unable to create new shared orm instance with orm.Error")
			return nil, orm.Error
		}
	}

	return orm, nil
}

func Orm(ctx context.Context, config database.Configuration) (*database.Orm, error) {
	o, err := orm(ctx, config, lambdaSpecificOrm)
	if err != nil {
		logger.WithCtx(ctx).With("error", err).Error("unable to create orm")
		return nil, errors.New("high network traffic").WithErrorCode(errors.TooManyAttemptsErrorCode)
	}

	return o, nil
}

func ChatOrm(ctx context.Context, config database.Configuration) (*database.Orm, error) {
	o, err := orm(ctx, config, chatSpecificOrm)
	if err != nil {
		logger.WithCtx(ctx).With("error", err).Error("unable to create chat orm")
		return nil, errors.New("chat high network traffic").WithErrorCode(errors.TooManyAttemptsErrorCode)
	}

	return o, nil
}

func APIPanicHandler(response *events.APIGatewayProxyResponse) {
	if r := recover(); r != nil {
		logger.With("stack_trace", debug.Stack()).Error("recovered from panic: %v", r)

		var err error
		switch v := r.(type) {
		case error:
			err = v
		default:
			err = errors.New("unknown internal error; %v", v, errors.LambdaPanicErrorCode)
		}

		if response != nil {
			safeResponse, respErr := lambda_api.Response(err)
			if respErr != nil {
				logger.Error("failed to create response: %v", respErr)
			} else {
				*response = safeResponse
			}
		}
	}
}

func AuthorizerPanicHandler(response *events.APIGatewayCustomAuthorizerResponse) {
	if r := recover(); r != nil {
		logger.With("stack_trace", debug.Stack()).Error("recovered from panic: %v", r)
		*response = events.APIGatewayCustomAuthorizerResponse{
			PrincipalID:    "panicked",
			PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{},
			Context: map[string]interface{}{
				"error": errors.New("internal error").WithErrorCode(errors.LambdaPanicErrorCode),
			},
			UsageIdentifierKey: fmt.Sprintf("%s-%s", version.Version, version.Commit),
		}
	}
}

func SQSPanicHandler(e *error) {
	if r := recover(); r != nil {
		logger.With("stack_trace", debug.Stack()).Error("recovered from panic: %v", r)

		switch v := r.(type) {
		case error:
			*e = v
		default:
			*e = errors.New("unknown internal error; %v", v).WithErrorCode(errors.LambdaPanicErrorCode)
		}
	}
}
