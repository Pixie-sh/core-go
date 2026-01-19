package http

import (
	goHttp "net/http"
	"strings"

	fiberAdaptor "github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"
	"github.com/pixie-sh/logger-go/structs"

	"github.com/pixie-sh/core-go/pkg/models/response_models"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/uid"
)

// CtxKeyType key type to be used when adding custom structs to context
type CtxKeyType string

const (
	XRequestIDKey        = "X-Request-ID"
	LocalsRequestLogger  = "request_logger"
	LocalsTraceID        = logger.TraceID
	LocalsMetricsTraceID = "request_metrics_trace_id"
)

type RequestLog struct {
	Method  string
	URL     string
	Headers []byte
	Body    []byte
	IP      string
}

type ResponseLog struct {
	Status  int
	Headers []byte
	Body    []byte
	Error   error
}

func GoHandlerAdaptor(goHandler goHttp.HandlerFunc) ServerHandler {
	return fiberAdaptor.HTTPHandler(goHandler)
}

var errorHandlerSkipPaths = map[string]bool{
	"/health":    true,
	"/ws-health": true,
	"/metrics":   true,
}

func errorHandlerWithCustomProcessor(ctx ServerCtx, errInput error) error {
	// assuming all the controllers are using the http.APIResponse and http.APIError wrappers, this should be true
	if !types.Nil(ctx) && ctx.Response() != nil && len(ctx.Response().Body()) > 0 {
		return nil
	}

	if !env.IsDebugActive() && errorHandlerSkipPaths[ctx.Path()] {
		return nil
	}

	log := GetCtxLogger(ctx).
		With("at_error.response", ResponseLog{
			Status:  ctx.Response().StatusCode(),
			Headers: ctx.Response().Header.Header(),
			Body:    ctx.Response().Body(),
			Error:   errInput,
		}).
		With("at_error.request", RequestLog{
			Method:  ctx.Method(),
			URL:     ctx.OriginalURL(),
			Headers: ctx.Request().Header.Header(),
			Body:    ctx.Body(),
			IP:      ctx.IP(),
		})

	log.Debug("error at errorHandlerWithCustomProcessor")
	castedError, ok := errInput.(errors.E)
	if !ok {
		castedError = errors.New("%s", errInput.Error()).WithErrorCode(errors.UnknownErrorCode)
	}

	log.Error("http error handler %d - %s", castedError.Code.HTTPError, ctx.OriginalURL())
	return ctx.Status(castedError.Code.HTTPError).JSON(response_models.ResponseErrorBody{Error: castedError})
}

// ParseQueryParameters map the query parameters to map
// query is the map key
// string array are the list of comma separated values of that key
func ParseQueryParameters(ctx ServerCtx, withSplit ...bool) map[string][]string {
	queryParams := map[string][]string{}
	ctx.Context().QueryArgs().VisitAll(func(key []byte, val []byte) {
		k := *structs.UnsafeString(key)
		v := *structs.UnsafeString(val)
		if strings.Contains(v, ",") && (len(withSplit) > 0 && withSplit[0]) {
			values := strings.Split(v, ",")
			for i := 0; i < len(values); i++ {
				queryParams[k] = append(queryParams[k], values[i])
			}
		} else {
			queryParams[k] = append(queryParams[k], v)
		}
	})

	return queryParams
}

func ParamsUint64(ctx ServerCtx, key string) (uint64, error) {
	val, err := ctx.ParamsInt(key)
	if err != nil {
		return 0, err
	}

	return uint64(val), nil
}

func ParamsUID(ctx ServerCtx, key string) (uid.UID, error) {
	val, err := uid.FromString(ctx.Params(key))
	if err != nil {
		return uid.Nil, err
	}

	return val, nil
}
