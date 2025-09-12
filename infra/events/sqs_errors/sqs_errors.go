package sqs_errors

import (
	"fmt"

	"github.com/pixie-sh/errors-go"
)

var (
	SQSErrorCode         = 91000
	GenericErrorCodeWith = func(reason string, inputErrorCode int) errors.ErrorCode {
		return errors.NewErrorCode(fmt.Sprintf("%sErrorCode", reason), SQSErrorCode+inputErrorCode)
	}
)
