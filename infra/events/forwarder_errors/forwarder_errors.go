package forwarder_errors

import (
	"fmt"

	"github.com/pixie-sh/errors-go"
)

var (
	ForwarderErrorCode   = 92000
	GenericErrorCodeWith = func(reason string, httpErrorCode int) errors.ErrorCode {
		return errors.NewErrorCode(fmt.Sprintf("%sErrorCode", reason), ForwarderErrorCode+httpErrorCode)
	}

	ForwarderEmptyListErrorCode           = errors.NewErrorCode("ForwarderEmptyListErrorCode", ForwarderErrorCode+errors.HTTPInvalidData)
	ForwarderNilEventErrorCode            = errors.NewErrorCode("ForwarderNilEventErrorCode", ForwarderErrorCode+errors.HTTPInvalidData)
	ForwarderPayloadTypeMismatchErrorCode = errors.NewErrorCode("ForwarderPayloadTypeMismatchErrorCode", ForwarderErrorCode+errors.HTTPInvalidData)
	ForwarderTypeNotRegisteredErrorCode   = errors.NewErrorCode("ForwarderTypeNotRegisteredErrorCode", ForwarderErrorCode+errors.HTTPServerError)
)
