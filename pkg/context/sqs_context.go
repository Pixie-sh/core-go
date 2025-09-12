package context

import (
	aws "github.com/aws/aws-lambda-go/events"
)

type SQSContext struct {
	*GenericContext

	Event aws.SQSEvent
}
