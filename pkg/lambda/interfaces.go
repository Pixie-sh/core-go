package lambda

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type Authorizer interface {
	Lambda
	HandleAuthorizer(context.Context, events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error)
}

type SQS interface {
	Lambda
	HandleSQS(context.Context, events.SQSEvent) (events.SQSEventResponse, error)
	HandleSQSBatch(context.Context, []events.SQSEvent) (events.SQSEventResponse, error)
}

type Kafka interface {
	Lambda
	HandleKafka(context.Context, events.KafkaEvent) error
	HandleKafkaBatch(context.Context, []events.KafkaEvent) error
}

// API just interface to make all lambdas to comply with with API GW v1
// Deprecated: Use APIv2 which is meant to be used for API GW v2
type API interface {
	Lambda
	HandleAPI(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}

type APIv2 interface {
	Lambda
	HandleAPIV2(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error)
}

type Lambda interface {
	Init(ctx context.Context) error
	Defer()
}

type LambdaConfig interface {
	Load(ctx context.Context) error
}
