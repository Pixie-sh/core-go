package context

import (
	"context"
	"time"
)

// GenericContext complies with the context.Context interface
type GenericContext struct {
	goContext context.Context

	Locals map[string]any `json:"locals,omitempty"` //Note: map content is mutable
}

func NewGenericContext(ctx context.Context) *GenericContext {
	return &GenericContext{
		goContext: ctx,
		Locals:    make(map[string]any),
	}
}

func (c *GenericContext) Deadline() (deadline time.Time, ok bool) {
	return c.goContext.Deadline()
}

func (c *GenericContext) Done() <-chan struct{} {
	return c.goContext.Done()
}

func (c *GenericContext) Err() error {
	return c.goContext.Err()
}

func (c *GenericContext) Value(key any) any {
	return c.goContext.Value(key)
}

// SetUserContext is not thread safe. meant to be used on lambdas, per one execution context per thread
func (c *GenericContext) SetUserContext(ctx context.Context) *GenericContext {
	c.goContext = ctx
	return c
}

func (c *GenericContext) SetLocals(value map[string]any) *GenericContext {
	c.Locals = value
	return c
}

func (c *GenericContext) SetLocal(key string, value any) *GenericContext {
	c.Locals[key] = value
	return c
}

func (c *GenericContext) GetLocal(key string) any {
	return c.Locals[key]
}

func (c *GenericContext) Context() context.Context {
	return c.goContext
}
