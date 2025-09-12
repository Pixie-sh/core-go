package validators

import (
	"reflect"
	"strconv"
	"strings"
)

const MapContainsRuleValidatorTag = "map_contains"

var MapContainsRuleValidator = ValidatorPair{
	ValidatorTag: MapContainsRuleValidatorTag,
	ValidatorFn: func(fl Field) bool {
		field := fl.Field()
		if field.Kind() != reflect.Map {
			return false
		}

		keys := strings.Split(fl.Param(), "|")
		found := make(map[string]bool, len(keys))
		mapKeys := field.MapKeys()

		// Mark all desired keys as not found initially
		for _, key := range keys {
			found[key] = false
		}

		for _, key := range mapKeys {
			var keyString string

			// Convert the key to a string based on its type
			switch key.Kind() {
			case reflect.String:
				keyString = key.String()
			case reflect.Int, reflect.Int64:
				keyString = strconv.FormatInt(key.Int(), 10)
			case reflect.Uint, reflect.Uint64:
				keyString = strconv.FormatUint(key.Uint(), 10)
			case reflect.Bool:
				keyString = strconv.FormatBool(key.Bool())
			default:
				continue // Skip keys of unsupported types
			}

			// If the converted key string is one of the desired keys, mark as found
			if _, ok := found[keyString]; ok {
				found[keyString] = true
			}
		}

		// Check if all keys were found
		for _, v := range found {
			if !v {
				return false
			}
		}

		return true
	},
}
