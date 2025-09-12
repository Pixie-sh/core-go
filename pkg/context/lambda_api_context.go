package context

import (
	aws "github.com/aws/aws-lambda-go/events"
)

type LambdaAPIContext struct {
	GenericContext

	// Deprecated
	// Either Request or RequestV2 maybe affected it depends on the deployment
	// default since 11st April 2024 is RequestV2
	// Request is deprecated, do not use it
	Request *aws.APIGatewayProxyRequest

	RequestV2 *aws.APIGatewayV2HTTPRequest
}
