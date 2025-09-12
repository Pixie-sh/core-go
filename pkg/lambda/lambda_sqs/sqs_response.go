package lambda_sqs

import (
	"errors"

	"github.com/aws/aws-lambda-go/events"
)

// Response -
// If we pass an error, it will fail the whole batch straight away, this could be usefull
// in case we don't want to provide a specific messageID, or if there's none.
func Response(args ...any) (events.SQSEventResponse, error) {
	var err error
	var resp events.SQSEventResponse

	for _, arg := range args {

		switch arg := arg.(type) {
		case error:
			// If we pass an error we might want to fail the whole batch
			if err == nil {
				err = arg
				resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{
					ItemIdentifier: "", // Batch will fail if empty item identifier
				})
			}
		case events.SQSEventResponse:
			resp = arg
		case events.SQSBatchItemFailure:
			resp.BatchItemFailures = append(resp.BatchItemFailures, arg)
		case []events.SQSBatchItemFailure:
			resp.BatchItemFailures = append(resp.BatchItemFailures, arg...)
		}

	}

	if len(resp.BatchItemFailures) > 0 && err != nil {
		err = errors.New("SQS event could not be processed correctly")

	}

	return resp, err
}
