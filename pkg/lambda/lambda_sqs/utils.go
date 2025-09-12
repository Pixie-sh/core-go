package lambda_sqs

import (
	"github.com/aws/aws-lambda-go/events"
)

func GetMessageAttributeString(message events.SQSMessage, attributeKey string) string {
	reqScope, ok := message.MessageAttributes[attributeKey]
	if !ok {
		return ""
	}
	return *reqScope.StringValue
}
