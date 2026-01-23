package rest

import goHttp "net/http"

type Response[T any] struct {
	Data        T                `json:"data"`
	Error       error            `json:"error,omitempty"`
	RawBody     []byte           `json:"raw_body,omitempty"`
	RawResponse *goHttp.Response `json:"-"`
}
