package notification_models

import (
	"github.com/pixie-sh/core-go/infra/message_wrapper"
)

type DangerProduceSQSModel struct {
	EventPayload message_wrapper.UntypedMessage `json:"event_payload"`
	SQSUrl       string                         `json:"sqs_url"`
	IsFIFO       bool                           `json:"is_fifo"`
}
