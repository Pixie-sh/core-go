package utils

import (
	"reflect"
	"strings"
)

// SchemaJSON generates a JSON schema for a given struct type
func SchemaJSON(v interface{}) map[string]interface{} {
	t := reflect.TypeOf(v)
	schema := map[string]interface{}{
		"$schema":    "http://json-schema.org/draft-07/schema#",
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}

	properties := schema["properties"].(map[string]interface{})
	required := schema["required"].([]string)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		jsonParts := strings.Split(jsonTag, ",")
		jsonName := jsonParts[0]
		if jsonName == "" {
			jsonName = field.Name
		}

		fieldSchema := map[string]interface{}{}
		switch field.Type.Kind() {
		case reflect.String:
			fieldSchema["type"] = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldSchema["type"] = "integer"
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldSchema["type"] = "integer"
		case reflect.Float32, reflect.Float64:
			fieldSchema["type"] = "number"
		case reflect.Bool:
			fieldSchema["type"] = "boolean"
		case reflect.Slice:
			elemType := field.Type.Elem().Kind()
			elemSchema := map[string]interface{}{"type": "string"} // Default to string for simplicity
			switch elemType {
			case reflect.String:
				elemSchema["type"] = "string"
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				elemSchema["type"] = "integer"
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				elemSchema["type"] = "integer"
			case reflect.Float32, reflect.Float64:
				elemSchema["type"] = "number"
			case reflect.Bool:
				elemSchema["type"] = "boolean"
			case reflect.Struct:
				elemSchema = SchemaJSON(reflect.New(field.Type.Elem()).Elem().Interface())
			}
			fieldSchema["type"] = "array"
			fieldSchema["items"] = elemSchema
		case reflect.Struct:
			fieldSchema = SchemaJSON(reflect.New(field.Type).Elem().Interface())
		default:
			fieldSchema["type"] = "string" // Default to string for simplicity
		}

		properties[jsonName] = fieldSchema

		// Check for omitempty and validate:"required" tags
		if len(jsonParts) > 1 && jsonParts[1] == "omitempty" {
			continue
		}
		validateTag := field.Tag.Get("validate")
		if validateTag == "required" || (len(jsonParts) <= 1 || jsonParts[1] != "omitempty") {
			required = append(required, jsonName)
		}
	}

	if len(required) == 0 {
		delete(schema, "required")
	} else {
		schema["required"] = required
	}

	return schema
}

type SchemaDescriptionsModel map[string]string

func SchemaDescriptions(t interface{}) SchemaDescriptionsModel {
	descriptions := make(map[string]string)
	val := reflect.ValueOf(t)
	typ := reflect.TypeOf(t)

	// Get the event_description tag from the embedded Event field
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous {
			eventDescription := field.Tag.Get("event_description")
			if eventDescription != "" {
				descriptions["event_description"] = eventDescription
			}
		}
	}

	// Get the field_description tags from the fields of the struct
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldDescription := field.Tag.Get("field_description")
		if fieldDescription != "" {
			descriptions[field.Tag.Get("json")] = fieldDescription
		}
	}

	return descriptions
}
