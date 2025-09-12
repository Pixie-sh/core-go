package lambda

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/lambda/lambda_api"
	"github.com/pixie-sh/core-go/pkg/lambda/lambda_shared"
	"github.com/pixie-sh/core-go/pkg/lambda/router"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type AbstractAPILambdaConfiguration struct {
	AbstractLambdaConfiguration

	RouterRoutePrefix string `json:"router_prefix"`
}

func (c AbstractAPILambdaConfiguration) Load(ctx context.Context) error {
	return c.AbstractLambdaConfiguration.Load(ctx)
}

type AbstractAPILambda struct {
	AbstractLambda
	router.APIRouter

	config AbstractAPILambdaConfiguration
}

func NewAbstractAPILambda(ctx context.Context, config AbstractAPILambdaConfiguration) AbstractAPILambda {
	return AbstractAPILambda{
		APIRouter: *router.NewAPIRouter(ctx, config.RouterRoutePrefix),
		config:    config,
	}
}

func (gl *AbstractAPILambda) Init(ctx context.Context) error {
	return gl.AbstractLambda.Init(ctx)
}

func (gl *AbstractAPILambda) HandleAPI(
	_ context.Context,
	_ events.APIGatewayProxyRequest,
) (resp events.APIGatewayProxyResponse, err error) {
	//defer lambda_shared.LoggingV1(ctx, request)(&resp, err)
	//return gl.APIRouter.Handle(ctx, request.HTTPMethod, request.Resource, request)
	return lambda_api.APIErrorResponse(errors.New("format not supported. upgrade to API GW v2").WithErrorCode(errors.ForbiddenErrorCode))
}

func (gl *AbstractAPILambda) HandleAPIV2(
	ctx context.Context,
	request events.APIGatewayV2HTTPRequest,
) (resp events.APIGatewayProxyResponse, err error) {
	traceID := uid.NewUUID()
	ctx = context.WithValue(ctx, logger.TraceID, traceID)

	defer lambda_shared.LoggingAPIWithTrace(ctx, traceID, request)(&resp, err)

	keys := strings.Split(request.RouteKey, " ")
	resp, err = gl.HandleV2(ctx, keys[0], keys[1], request)

	return
}

func (gl *AbstractAPILambda) Defer() {}
