package context

import (
	aws "github.com/aws/aws-lambda-go/events"
)

type KafkaContext struct {
	GenericContext

	Request *aws.KafkaEvent
}
