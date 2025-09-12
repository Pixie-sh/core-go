package router

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockRouteKey struct {
	key string
}

func (m mockRouteKey) Key() string {
	return m.key
}

func TestNewGenericRouter(t *testing.T) {
	router := NewGenericRouter[mockRouteKey, SQSHandler](context.Background(), "prefix")
	assert.NotNil(t, router)
	assert.Empty(t, router.routes)
	assert.Nil(t, router.gates)
	assert.Equal(t, "prefix", router.routePrefix)
}

func TestPreAllocate(t *testing.T) {
	router := NewGenericRouter[mockRouteKey, SQSHandler](context.Background(), "prefix")
	router.PreAllocate(10)
	assert.NotNil(t, router.routes)
	assert.Len(t, router.routes, 0)
}
