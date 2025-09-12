package db_errors

import (
	"strings"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/types"
)

func Handle(e error) error {
	if e == nil {
		return nil
	}

	if types.InstanceOf[errors.E](e) {
		return e
	}

	//custom handlers
	errStr := e.Error()
	if strings.Contains(errStr, "idx_user_email_address") || strings.Contains(e.Error(), "user_email_address_key") {
		return errors.New("Validation failed", &errors.FieldError{
			Field:   "email",
			Rule:    "duplicate",
			Message: "Email already taken",
		})
	}

	if strings.Contains(errStr, "user_phone_number_key") {
		return errors.New("Validation failed", &errors.FieldError{
			Field:   "phone_number",
			Rule:    "duplicate",
			Message: "Phone Number already taken",
		})
	}

	if strings.Contains(errStr, "duplicate key value violates unique constraint") {
		return errors.New("Validation failed", &errors.FieldError{
			Rule:    "duplicateKeyConstraint",
			Message: errStr,
		})
	}

	handledErr, handled := handleGormError(e)
	if handled {
		return handledErr
	}

	return errors.NewWithError(e, "db execution aborted").WithErrorCode(errors.DBErrorCode)
}

func Has(e error, code errors.ErrorCode) bool {
	_, tru := errors.Has(Handle(e), code)
	return tru
}
