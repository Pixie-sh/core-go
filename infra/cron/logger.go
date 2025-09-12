package cron

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	plogger "github.com/pixie-sh/logger-go/logger"
)

type cronLogger struct {
	plogger.Interface
}

func (l cronLogger) Info(msg string, keysAndValues ...interface{}) {
	l.Log(fmt.Sprintf("%s %s", msg, l.formatKeyValuePairs(keysAndValues)))
}

func (l cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.With("error", err).Error(fmt.Sprintf("%s %s", msg, l.formatKeyValuePairs(keysAndValues)))
}

func newLogger(_ context.Context, log plogger.Interface) *cronLogger {
	return &cronLogger{
		Interface: log,
	}
}

func (l cronLogger) formatKeyValuePairs(keysAndValues []interface{}) string {
	var pairs []string
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			value := l.formatValue(keysAndValues[i+1])
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
		}
	}
	return strings.Join(pairs, " ")
}

func (l cronLogger) formatValue(v interface{}) string {
	if v == nil {
		return "null"
	}

	// Use reflection to check the type
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return "null"
		}
		return l.formatValue(val.Elem().Interface())
	}

	switch value := v.(type) {
	case time.Time:
		return value.Format(time.RFC3339)
	case fmt.Stringer:
		return value.String()
	case string:
		return fmt.Sprintf("%q", value)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", value)
	case float32, float64:
		return fmt.Sprintf("%f", value)
	case bool:
		return fmt.Sprintf("%t", value)
	default:
		// For complex types, use %+v for a more detailed output
		return fmt.Sprintf("%+v", value)
	}
}
