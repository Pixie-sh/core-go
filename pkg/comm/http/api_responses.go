package http

import (
	"net/http"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/errors-go/utils"

	"github.com/pixie-sh/core-go/pkg/models/response_models"
)

func APIError(ctx ServerCtx, respErr error, codeOptional ...int) error {
	casted, ok := errors.As(respErr)
	if !ok {
		casted = errors.NewWithError(respErr, "%s", http.StatusText(errors.ErrorPerformingRequestErrorCode.HTTPError)).WithErrorCode(errors.ErrorPerformingRequestErrorCode)
	}

	code := casted.Code.HTTPError
	if len(codeOptional) > 0 {
		code = codeOptional[0]
	}

	return ctx.Status(code).JSON(response_models.ResponseErrorBody{Error: casted})
}

func APIResponse(ctx ServerCtx, body any, codeOptional ...int) error {
	code := http.StatusOK
	if len(codeOptional) > 0 {
		code = codeOptional[0]
	}

	return ctx.Status(code).JSON(response_models.ResponseBody[any]{Data: body})
}

// Response wrapper to remove this always the same logic from controllers
// examples:
//
//	return http.Response(ctx, "Ok", 200)
//	return http.Response(ctx, data, err)
//	return http.Response(ctx, err, data, 200)
//	return http.Response(ctx, 200, err, data)
func Response(args ...interface{}) error {
	var err error
	var code = -1
	var body interface{}
	var ctx ServerCtx

	for _, arg := range args {
		if utils.Nil(arg) {
			continue
		}

		switch arg.(type) {
		case ServerCtx:
			ctx = arg.(ServerCtx)
			break

		case error:
			if err == nil {
				err = arg.(error)
			}
			break

		case int:
			if code == -1 {
				code = arg.(int)
			}
		default:
			if body == nil {
				body = arg
			}
		}

		if err != nil && code != -1 {
			break
		}
	}

	if utils.Nil(ctx) {
		panic("http.Response requires a valid ctx instance")
	}

	var codeOptional []int
	if code != -1 {
		codeOptional = append(codeOptional, code)
	}

	if err != nil {
		return APIError(ctx, err, codeOptional...)
	}

	if body == nil {
		body = "Ok"
	}

	return APIResponse(ctx, body, codeOptional...)
}
