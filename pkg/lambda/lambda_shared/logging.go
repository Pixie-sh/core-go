package lambda_shared

import (
	"context"
	"time"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/types/slices"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pixie-sh/logger-go/logger"
)

type LambdaRequestLog struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"queryParams"`
	Body        string            `json:"body"`
	RequestTime string            `json:"requestTime"`
	TraceID     string            `json:"traceID"`
}

type LambdaResponseLog struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Error      error             `json:"error,omitempty"`
}

type logging struct{}

func LoggingAPIWithTrace(ctx context.Context, traceID string, request events.APIGatewayV2HTTPRequest) func(resp *events.APIGatewayProxyResponse, err error) {
	return logging{}.HandleAPI(ctx, traceID, request)
}

func LoggingSQSWithTrace(ctx context.Context, traceID string, request events.SQSEvent) func(resp *events.SQSEventResponse, err error) {
	return logging{}.HandleSQS(ctx, traceID, request)
}

func (l logging) HandleAPI(
	ctx context.Context,
	traceID string,
	request events.APIGatewayV2HTTPRequest,
) func(resp *events.APIGatewayProxyResponse, err error) {
	start := time.Now()

	reqLog := LambdaRequestLog{
		Method:      request.RequestContext.HTTP.Method,
		Path:        request.RequestContext.HTTP.Path,
		Headers:     request.Headers,
		QueryParams: request.QueryStringParameters,
		Body:        request.Body,
		RequestTime: request.RequestContext.Time,
		TraceID:     traceID,
	}

	log := logger.
		WithCtx(ctx).
		With(logger.TraceID, traceID).
		With("request", reqLog)

	log.Log("request %s %s", reqLog.Method, reqLog.Path)

	return func(response *events.APIGatewayProxyResponse, err error) {
		if types.Nil(response.Headers) {
			response.Headers = make(map[string]string)
		}

		response.Headers[XRequestIDKey] = traceID
		response.Headers["Access-Control-Expose-Headers"] = XRequestIDKey

		resLog := LambdaResponseLog{
			StatusCode: response.StatusCode,
			Headers:    response.Headers,
			Body:       response.Body,
		}

		if err != nil {
			resLog.Error = err
		}

		log.
			With("response", resLog).
			Log("request finished: %s; took %s", reqLog.Path, time.Since(start).String())
	}
}

func (l logging) HandleSQS(ctx context.Context, traceID string, request events.SQSEvent) func(resp *events.SQSEventResponse, err error) {
	start := time.Now()

	log := pixiecontext.GetCtxLogger(ctx).
		With(logger.TraceID, traceID).
		With("records", len(request.Records))

	msgIds := slices.Map(request.Records, func(msg events.SQSMessage) string {
		return msg.MessageId
	})

	log.
		With("MessageIDs", msgIds).
		Log("Event contains %s event records", len(request.Records))

	return func(response *events.SQSEventResponse, err error) {

		log.
			With("response", response).
			Log("queue finished: took %s", time.Since(start).String())
	}
}
