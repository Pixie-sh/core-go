package lambda

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/lambda/lambda_shared"
	"github.com/pixie-sh/core-go/pkg/lambda/lambda_sqs"
	pixierouter "github.com/pixie-sh/core-go/pkg/lambda/router"
	"github.com/pixie-sh/core-go/pkg/uid"
)

type AbstractSQSLambdaConfiguration struct {
	AbstractLambdaConfiguration
	RouterRoutePrefix string `json:"router_prefix"`
}

func (c AbstractSQSLambdaConfiguration) Load(ctx context.Context) error {
	return c.AbstractLambdaConfiguration.Load(ctx)
}

type AbstractSQSLambda struct {
	AbstractLambda
	*pixierouter.SQSRouter

	config AbstractSQSLambdaConfiguration
}

func NewAbstractSQSLambda(ctx context.Context, config AbstractSQSLambdaConfiguration) AbstractSQSLambda {
	return AbstractSQSLambda{
		SQSRouter: pixierouter.NewSQSRouter(ctx, config.RouterRoutePrefix),
		config:    config,
	}
}

func (gl *AbstractSQSLambda) Init(ctx context.Context) error {
	return gl.AbstractLambda.Init(ctx)
}

func (gl *AbstractSQSLambda) HandleSQS(ctx context.Context, evs events.SQSEvent) (resp events.SQSEventResponse, err error) {
	traceID := uid.NewUUID()
	ctx = context.WithValue(ctx, logger.TraceID, traceID)
	ctxLog := pixiecontext.GetCtxLogger(ctx)

	defer lambda_shared.LoggingSQSWithTrace(ctx, traceID, evs)(&resp, err)

	if len(evs.Records) == 0 {
		// TODO: Maybe we just want to skip this, this might be retried on fail
		// However i'm not sure so we def have to test this
		return lambda_sqs.Response(errors.New("invalid events.records; empty list.", errors.InvalidRecordsListErrorCode))
	}

	var evsByQueue = map[string][]events.SQSMessage{}
	for _, sqsMsg := range evs.Records {
		arnSplit := strings.Split(sqsMsg.EventSourceARN, ":")
		queueName := arnSplit[len(arnSplit)-1]

		ctxLog.Debug("Got SQS queue name: %s", queueName)

		evsByQueue[queueName] = append(evsByQueue[queueName], sqsMsg)
	}

	var respAcc events.SQSEventResponse
	var errAcc []error
	for queue, sqsMsgs := range evsByQueue {
		// TODO: Change this shit after we have the error aggregator
		resp, err = gl.HandleSQSMessage(ctx, queue, sqsMsgs)
		respAcc.BatchItemFailures = append(respAcc.BatchItemFailures, resp.BatchItemFailures...)
		errAcc = append(errAcc, err) // ! Maybe this is not needed, since the batchItemFailures contain the potential errors
	}

	return respAcc, nil
}

func (gl *AbstractSQSLambda) HandleSQSBatch(ctx context.Context, evs []events.SQSEvent) (resp events.SQSEventResponse, err error) {
	var respAcc []events.SQSEventResponse
	var errAcc []error
	for _, ev := range evs {
		// TODO: Change this shit after we have the error aggregator
		resp, err = gl.HandleSQS(ctx, ev)
		respAcc = append(respAcc, resp)
		errAcc = append(errAcc, err)
	}

	return
}

func (gl *AbstractSQSLambda) Defer() {}
