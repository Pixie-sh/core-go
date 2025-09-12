package lambda_api

import (
	goHttp "net/http"

	aws "github.com/aws/aws-lambda-go/events"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/errors-go/utils"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/comm/http"
	"github.com/pixie-sh/core-go/pkg/models/response_models"
	"github.com/pixie-sh/core-go/pkg/models/serializer"
)

func APIErrorResponse(respErr error, codeOptional ...int) (aws.APIGatewayProxyResponse, error) {
	if respErr == nil {
		logger.Error("unable to parse error properly")
		return aws.APIGatewayProxyResponse{
			StatusCode:      errors.UnknownErrorCode.HTTPError,
			Body:            errors.UnknownErrorCode.String(),
			IsBase64Encoded: false,
		}, nil
	}
	casted, ok := errors.As(respErr)
	if !ok {
		return newApiErrorFromE(errors.NewWithError(respErr, "Internal Server Error").WithErrorCode(errors.UnknownErrorCode))
	}

	code := casted.Code.HTTPError
	if len(codeOptional) > 0 {
		code = codeOptional[0]
	}

	return newApiErrorFromE(casted, code)
}

func newApiErrorFromE(casted errors.E, codeOptional ...int) (aws.APIGatewayProxyResponse, error) {
	blob, _ := serializer.Serialize(response_models.ResponseErrorBody{
		Error: casted,
	})

	code := casted.Code.HTTPError
	if len(codeOptional) > 0 {
		code = codeOptional[0]
	}

	return aws.APIGatewayProxyResponse{
		StatusCode:      code,
		Body:            string(blob),
		IsBase64Encoded: false,
		Headers: map[string]string{
			"Content-PayloadType": "Application/json",
		},
	}, nil
}

var OkBody = goHttp.StatusText(http.StatusOK)

// APIResponse wrapper for HTTP Lambda API response.
// for codeOptional only the first code is used. Default is 200
func APIResponse(body interface{}, codeOptional ...int) (aws.APIGatewayProxyResponse, error) {
	var code = http.StatusOK

	if len(codeOptional) > 0 {
		codeText := goHttp.StatusText(codeOptional[0])
		if len(codeText) > 0 {
			code = codeOptional[0]
		}
	}

	blob, _ := serializer.Serialize(response_models.ResponseBody[any]{
		Data: body,
	})

	return aws.APIGatewayProxyResponse{
		StatusCode:      code,
		Body:            string(blob),
		IsBase64Encoded: false,
		Headers: map[string]string{
			"Content-PayloadType": "Application/json",
		},
	}, nil
}

// Response wrapper to remove this always the same logic from controllers
// examples:
//
//	return lambda_api.Response("Ok", 200)
//	return lambda_api.Response(data, err)
//	return lambda_api.Response(err, data, 200)
//	return lambda_api.Response(200, err, data)
func Response(args ...interface{}) (aws.APIGatewayProxyResponse, error) {
	var err error
	var code = -1
	var body interface{}

	for _, arg := range args {
		if utils.Nil(arg) {
			continue
		}

		switch arg.(type) {
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

	var codeOptional []int
	if code != -1 {
		codeOptional = append(codeOptional, code)
	}

	if err != nil && code == -1 {
		return APIErrorResponse(err)
	} else if err != nil {
		return APIErrorResponse(err, []int{code}...)
	}

	if body == nil {
		body = "Ok"
	}

	return APIResponse(body, []int{code}...)
}
