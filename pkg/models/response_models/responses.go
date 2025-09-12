package response_models

import "github.com/pixie-sh/errors-go"

type ResponseErrorBody struct {
	Error errors.E `json:"error,omitempty"`
}

type ResponseBody[T any] struct {
	Data T `json:"data,omitempty"`
}

type Response[T any] struct {
	ResponseBody[T]
	ResponseErrorBody
}
