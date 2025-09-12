package errors

import (
	"fmt"

	"github.com/pixie-sh/errors-go"
)

var (
	BaseErrorCodeValue   = 90000
	GenericErrorCodeWith = func(reason string, httpErrorCode int) errors.ErrorCode {
		return errors.NewErrorCode(fmt.Sprintf("%sErrorCode", reason), BaseErrorCodeValue+httpErrorCode)
	}

	MessageFactoryDuplicateRegistrationErrorCode = errors.NewErrorCode("MessageFactoryDuplicateRegistrationErrorCode", BaseErrorCodeValue+errors.HTTPConflict)
	TemplatePanicErrorCode                       = errors.NewErrorCode("TemplatePanicErrorCode", BaseErrorCodeValue+errors.HTTPServerError)
	TemplateDoNotTriggerErrorCode                = errors.NewErrorCode("TemplateDoNotTriggerErrorCode", BaseErrorCodeValue+errors.HTTPServerError)
	TagsNotFoundErrorCode                        = errors.NewErrorCode("TagsNotFoundErrorCode", BaseErrorCodeValue+errors.HTTPNotFound)
	TagsInvalidFormatErrorCode                   = errors.NewErrorCode("TagsInvalidFormatErrorCode", BaseErrorCodeValue+errors.HTTPInvalidData)
	TagsInvalidScopeErrorCode                    = errors.NewErrorCode("TagsInvalidScopeErrorCode", BaseErrorCodeValue+errors.HTTPInvalidData)
	TagsDependencyErrorCode                      = errors.NewErrorCode("TagsDependencyErrorCode", BaseErrorCodeValue+errors.HTTPServerError)

	ConfigurationLookupErrorCode = errors.NewErrorCode("ConfigurationLoadErrorCode", BaseErrorCodeValue+503)
)
