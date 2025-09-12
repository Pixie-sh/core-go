package serializer

import (
	"reflect"

	gojson "github.com/goccy/go-json"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/errors-go/utils"

	"github.com/pixie-sh/core-go/pkg/models/validators"
	"github.com/pixie-sh/core-go/pkg/types"
)

type Raw = map[string]interface{}

// Serialize model to payload
func Serialize(event interface{}) ([]byte, error) {
	if utils.Nil(event) {
		return nil, errors.New("event must not be nil").WithErrorCode(errors.InvalidFormDataCode)
	}

	switch expr := event.(type) {
	case string:
		return types.UnsafeBytes(expr), nil
	case []byte:
		return expr, nil
	default:
		return gojson.Marshal(event)
	}
}

// Deserialize a model into struct
func Deserialize(blob []byte, dest interface{}, withValidations ...bool) error {
	if blob == nil || len(blob) == 0 || !types.IsJSON(blob) {
		return errors.NewValidationError("provided payload is invalid, must be json")
	}

	if !utils.IsPointer(dest) {
		return errors.New("dest %s must be pointer", reflect.TypeOf(dest)).WithErrorCode(errors.InvalidFormDataCode)
	}

	err := gojson.Unmarshal(blob, dest)
	if err != nil {
		return errors.New("error processing deserialization: %+v", err).WithErrorCode(errors.InvalidFormDataCode)
	}

	if len(withValidations) == 0 || (len(withValidations) > 0 && withValidations[0]) {
		return Validate(dest)
	}
	return nil
}

// DeserializeFromStr a model into struct
func DeserializeFromStr(blob string, dest interface{}, withValidations ...bool) error {
	return Deserialize(types.UnsafeBytes(blob), dest, withValidations...)
}

func DeserializeFromFn(parser func(out interface{}) error, dest interface{}, withValidations ...bool) error {
	err := parser(dest)
	if err != nil {
		return errors.New("error processing deserialization: %+v", err).WithErrorCode(errors.InvalidFormDataCode)
	}

	if len(withValidations) == 0 || (len(withValidations) > 0 && withValidations[0]) {
		return Validate(dest)
	}

	return nil
}

func Validate(dest interface{}) error {
	err := validators.Validate(dest)
	if err != nil {
		return validators.HandleError(err)
	}

	return nil
}
