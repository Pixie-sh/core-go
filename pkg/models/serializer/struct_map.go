package serializer

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/errors-go/utils"

	"github.com/pixie-sh/core-go/pkg/models/validators"
	"github.com/pixie-sh/core-go/pkg/uid"
)

func StructToMap[T map[string]interface{}](fromStruct any, withFromValidations ...bool) (T, error) {
	if len(withFromValidations) > 0 && withFromValidations[0] {
		err := Validate(fromStruct)
		if err != nil {
			return nil, err
		}
	}

	var result = make(T)
	var decoder, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			timeToStringHook,
		),
		TagName: "json",
		Result:  &result,
	})
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(fromStruct)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ToJSONB(fromStruct any, withFromValidations ...bool) (map[string]interface{}, error) {
	return StructToMap(fromStruct, withFromValidations...)
}

func FromJSONB[T any](fromStruct map[string]interface{}, to T, withToValidations ...bool) error {
	return ToStruct(fromStruct, to, withToValidations...)
}

func ToStruct[T any](from any, to T, withStructValidations ...bool) error {
	if !utils.IsPointer(to) {
		return errors.New("to %s must be pointer", reflect.TypeOf(to))
	}

	if utils.Nil(from) {
		return errors.New("from struct must not be nil")
	}

	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				stringToTimeHook,
				stringToUIDHookFunc,
			),
			TagName: "json",
			Result:  to,
			Squash:  true,
		})
	if err != nil {
		return err
	}

	err = decoder.Decode(from)
	if err != nil {
		return err
	}

	if len(withStructValidations) > 0 && withStructValidations[0] {
		return Validate(to)
	}

	return nil
}

func FromAny[T any](data any, withDataValidations ...bool) (T, error) {
	var m T
	return m, ToStruct(data, &m, withDataValidations...)
}

func stringToUIDHookFunc(from reflect.Type, to reflect.Type, data any) (any, error) {
	// Check if the types match the criteria: string -> uid.UID
	if from.Kind() == reflect.String && to == reflect.TypeOf(uid.UID{}) {
		return uid.FromString(data.(string))
	}
	return data, nil
}

func stringToTimeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f == reflect.TypeOf("") && t == reflect.TypeOf(time.Time{}) {
		parsedTime, err := time.Parse(time.RFC3339, data.(string))
		if err != nil {
			return nil, err
		}
		return parsedTime, nil
	}

	if f == reflect.TypeOf(map[string]interface{}{}) && t == reflect.TypeOf(time.Time{}) {
		dataCasted, ok := data.(map[string]interface{})
		if !ok {
			return nil, validators.HandleError(errors.New("data is not a map"))
		}

		parsedTime, err := DeserializeMapToTime(dataCasted)
		if err != nil {
			return nil, err
		}
		return *parsedTime, nil
	}

	return data, nil
}

func timeToStringHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f == reflect.TypeOf(&time.Time{}) {
		return SerializeTimeToMap(data.(*time.Time)), nil
	}

	return data, nil
}

func SerializeTimeToMap(t *time.Time) map[string]string {
	return map[string]string{
		"RFC3339": t.UTC().Format(time.RFC3339),
	}
}

func DeserializeMapToTime(data map[string]interface{}) (*time.Time, error) {
	timeStr, ok := data["RFC3339"].(string)
	if !ok {
		return nil, validators.HandleError(errors.New("RFC3339 key not found or not a string"))
	}

	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, err
	}
	return &parsedTime, nil
}
