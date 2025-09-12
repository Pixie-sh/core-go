package validators

import (
	"context"
	goErrors "errors"

	"github.com/go-playground/validator/v10"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"
)

type Field = validator.FieldLevel
type Validator = validator.Validate

type ValidatorPair struct {
	ValidatorTag string
	ValidatorFn  func(fl Field) bool
}

type ValidatorPairWithCheckNil struct {
	ValidatorTag string
	ValidatorFn  func(fl Field) bool
	CheckNil     bool
}

func New(pairs ...ValidatorPair) (*Validator, error) {
	v := validator.New()

	for _, pair := range pairs {
		err := v.RegisterValidation(pair.ValidatorTag, pair.ValidatorFn)
		if err != nil {
			logger.Error("unable to register validator: %s", err)
			return nil, err
		}
	}

	return v, nil
}

// For now don't want to mess with existing code, so I'm duplicating the New function
// the callValidationsEvenIfNull = true is required by the Conditional() validator in order to work with struct with pointers
// For now it's only used on deal domain
func NewWithValidateNil(checkNilMap map[string]bool, pairs ...ValidatorPair) (*Validator, error) {
	v := validator.New()

	for _, pair := range pairs {
		err := v.RegisterValidation(pair.ValidatorTag, pair.ValidatorFn, checkNilMap[pair.ValidatorTag])
		if err != nil {
			logger.Error("unable to register validator: %s", err)
			return nil, err
		}
	}

	return v, nil
}

func Validate(st interface{}) error {
	return V.StructCtx(context.Background(), st)
}

var V, _ = New() //default validator instance; override it to specify validators

func HandleError(validationErr error) error {
	if validationErr == nil {
		return nil
	}

	var casted validator.ValidationErrors

	switch {
	case goErrors.As(validationErr, &casted):
		errorResult := new(errors.Error)
		errorResult.Code = errors.InvalidFormDataCode
		errorResult.Message = "Validation failed"

		for _, fieldError := range casted {
			failure := &errors.FieldError{
				Field:   fieldError.Field(),
				Rule:    fieldError.Tag(),
				Param:   fieldError.Param(),
				Message: fieldError.Error(),
			}

			errorResult.FieldErrors = append(errorResult.FieldErrors, failure)
		}

		return errorResult
	default:
		return errors.NewWithError(validationErr, "%s", validationErr.Error()).WithErrorCode(errors.InvalidFormDataCode)
	}
}
