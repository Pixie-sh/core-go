package maps

import (
	"reflect"

	"github.com/pixie-sh/core-go/pkg/types"
)

func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func MapSliceValues[K comparable, V any](m map[K][]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v...)
	}
	return values
}

func MapStructValue[K any, V any](strukts []K, f func(strukt K) V, withEmpty ...bool) []V {
	vals := make([]V, len(strukts))
	for i, v := range strukts {
		if !(len(withEmpty) > 0 && withEmpty[0]) && (types.IsEmpty(v) || types.Nil(v)) {
			continue
		}

		vals[i] = f(v)
	}

	return vals
}

func MapKeysFilteringOnValueField[K comparable, V any](data map[K]V, field string, value interface{}) []K {
	var keys []K

	for key, obj := range data {
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Struct {
			fieldVal := val.FieldByName(field)
			if fieldVal.IsValid() && fieldVal.Interface() == value {
				keys = append(keys, key)
			}
		}
	}

	return keys
}
