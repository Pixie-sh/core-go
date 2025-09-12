package lambda_api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pixie-sh/errors-go"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/uid"
)

func GetQueryParametersForPipeline(ctx *pixiecontext.LambdaAPIContext) map[string][]string {
	queryParams := make(map[string][]string)
	for k, v := range ctx.RequestV2.QueryStringParameters {
		queryParams[k] = append(queryParams[k], v)
	}

	return queryParams
}

func GetQueryParameters(ctx *pixiecontext.LambdaAPIContext) map[string]string {
	return ctx.RequestV2.QueryStringParameters
}

func GetQueryParameter(ctx *pixiecontext.LambdaAPIContext, key string) string {
	return ctx.RequestV2.QueryStringParameters[key]
}

func GetBoolQueryParameter(ctx *pixiecontext.LambdaAPIContext, key string, defaultValue bool) bool {
	stringValue, exists := ctx.RequestV2.QueryStringParameters[key]
	if !exists {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(stringValue)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

func GetPathUint64(ctx *pixiecontext.LambdaAPIContext, key string) uint64 {
	var valStr = ctx.RequestV2.PathParameters[key]
	val, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		panic(
			errors.NewValidationError(fmt.Sprintf("unable to parse value %s to uin64", key), &errors.FieldError{
				Field:   key,
				Rule:    "invalidFormat",
				Param:   valStr,
				Message: fmt.Sprintf("unable to parse value '%s' from key '%s'", valStr, key),
			}))
	}

	return val
}

func GetPathUID(ctx *pixiecontext.LambdaAPIContext, key string) uid.UID {
	var valStr = ctx.RequestV2.PathParameters[key]
	val, err := uid.FromString(valStr)
	if err != nil {
		panic(
			errors.NewValidationError(fmt.Sprintf("unable to parse value %s to uid", key), &errors.FieldError{
				Field:   key,
				Rule:    "invalidFormat",
				Param:   valStr,
				Message: fmt.Sprintf("unable to parse value '%s' from key '%s'", valStr, key),
			}))
	}

	return val
}

func GetPathString(ctx *pixiecontext.LambdaAPIContext, key string) string {
	var valStr = ctx.RequestV2.PathParameters[key]
	return valStr
}

func GetHeader(ctx *pixiecontext.LambdaAPIContext, key string) string {
	ua, ok := ctx.RequestV2.Headers[key]
	if !ok {
		ua, ok = ctx.RequestV2.Headers[strings.ToLower(key)]
		if !ok {
			return ""
		}
	}

	return ua
}
