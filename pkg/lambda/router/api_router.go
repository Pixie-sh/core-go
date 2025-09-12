package router

import (
	"context"
	"fmt"

	"github.com/pixie-sh/errors-go"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"

	"github.com/aws/aws-lambda-go/events"

	"github.com/pixie-sh/core-go/pkg/comm/http"
	"github.com/pixie-sh/core-go/pkg/lambda/lambda_api"
)

// APIRouteKey is a struct to hold the combination of HTTP method and path.
type APIRouteKey struct {
	Method string
	Path   string
}

func (k APIRouteKey) Key() string {
	return fmt.Sprintf("%s:%s", k.Method, k.Path)
}

// APIHandler is a type for handling API requests.
type APIHandler func(*pixiecontext.LambdaAPIContext) (events.APIGatewayProxyResponse, error)

// APIRouter is a concrete router for API routes.
type APIRouter struct {
	*GenericRouter[APIRouteKey, APIHandler]
}

func NewAPIRouter(ctx context.Context, routePrefix string) *APIRouter {
	return &APIRouter{
		GenericRouter: NewGenericRouter[APIRouteKey, APIHandler](ctx, routePrefix),
	}
}

type APIGroup struct {
	gates       []APIHandler
	routePrefix string
	apiRouter   *APIRouter
}

func (r *APIRouter) Group(
	resource string,
	handler ...APIHandler) *APIGroup {
	return &APIGroup{
		gates:       append([]APIHandler{}, handler...),
		routePrefix: resource,
		apiRouter:   r,
	}
}

func (r *APIGroup) Group(
	resource string,
	handler ...APIHandler) *APIGroup {
	return &APIGroup{
		gates:       append(r.gates, handler...),
		routePrefix: fmt.Sprintf("%s%s", r.routePrefix, resource),
		apiRouter:   r.apiRouter,
	}
}

// RegisterHandler registers a handler for a specific method and path.
func (r *APIGroup) RegisterHandler(
	_ context.Context,
	HTTPMethod http.Method,
	resource string,
	handler ...APIHandler,
) *APIGroup {
	key := APIRouteKey{Method: HTTPMethod, Path: fmt.Sprintf("%s%s%s", r.apiRouter.routePrefix, r.routePrefix, resource)}

	if r.apiRouter.routes == nil {
		r.apiRouter.routes = make(map[string][]APIHandler)
	}

	r.apiRouter.routes[key.Key()] = append(r.gates, handler...)
	return r
}

// RegisterHandler registers a handler for a specific method and path.
func (r *APIRouter) RegisterHandler(
	_ context.Context,
	HTTPMethod http.Method,
	resource string,
	handler ...APIHandler,
) *APIRouter {
	key := APIRouteKey{Method: HTTPMethod, Path: fmt.Sprintf("%s%s", r.routePrefix, resource)}

	if r.routes == nil {
		r.routes = make(map[string][]APIHandler)
	}

	r.routes[key.Key()] = handler
	return r
}

// Handle finds and executes the appropriate handler based on the method and path.
func (r *APIRouter) Handle(ctx context.Context, method http.Method, path string, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	apiCtx := &pixiecontext.LambdaAPIContext{Request: &request}
	apiCtx.SetUserContext(ctx)
	apiCtx.SetLocals(map[string]any{})

	for _, g := range r.gates {
		resp, err := g(apiCtx)
		if err != nil {
			return lambda_api.APIErrorResponse(err)
		}

		if resp.StatusCode >= 400 {
			return resp, nil
		}
	}

	key := APIRouteKey{Method: method, Path: path}
	if handler, ok := r.routes[key.Key()]; ok {
		for i := 0; i <= len(handler)-1; i++ {
			resp, err := handler[i](apiCtx)

			switch i {
			case len(handler) - 1: //last iteration return either result
				return resp, err
			default:
				if err != nil {
					return lambda_api.APIErrorResponse(err)
				}

				if resp.StatusCode >= 400 {
					return resp, nil
				}
			}
		}
	}

	return lambda_api.APIErrorResponse(errors.New("route not found %s %s", method, path).WithErrorCode(errors.NotFoundErrorCode))
}

func (r *APIRouter) HandleV2(ctx context.Context, method string, path string, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	apiCtx := &pixiecontext.LambdaAPIContext{RequestV2: &request}
	apiCtx.SetUserContext(ctx)
	apiCtx.SetLocals(map[string]any{})

	for _, g := range r.gates {
		resp, err := g(apiCtx)
		if err != nil {
			return lambda_api.APIErrorResponse(err)
		}

		if resp.StatusCode >= 400 {
			return resp, nil
		}
	}

	key := APIRouteKey{Method: method, Path: path}
	if handler, ok := r.routes[key.Key()]; ok {
		for i := 0; i <= len(handler)-1; i++ {
			resp, err := handler[i](apiCtx)

			switch i {
			case len(handler) - 1: //last iteration return either result
				return resp, err
			default:
				if err != nil {
					return lambda_api.APIErrorResponse(err)
				}

				if resp.StatusCode >= 400 {
					return resp, nil
				}
			}
		}
	}

	return lambda_api.APIErrorResponse(errors.New("route not found %s %s", method, path).WithErrorCode(errors.NotFoundErrorCode))
}
