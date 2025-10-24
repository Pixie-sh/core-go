package kafka

// Exported Kafka header keys used by producer and consumer.
const (
	XRetryUntilHeader  = "x-retry-until"
	XPayloadTypeHeader = "x-payload-type"
	XEventIDHeader     = "x-event-id"
	XRetryCountHeader  = "x-retry-count"
)
