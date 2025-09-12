package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewForwarder(t *testing.T) {

	forwarder, err := NewForwarder(context.Background(), ForwarderConfiguration{})

	assert.NoError(t, err)
	assert.Len(t, forwarder.rules, 0)

}

func TestRegisterRules(t *testing.T) {

	ctx := context.Background()

	testHandler := func(ctx context.Context, wrappers ...UntypedEventWrapper) error {
		return nil
	}

	testHandler2 := func(ctx context.Context, wrappers ...UntypedEventWrapper) error {
		return errors.New("test")
	}

	exampleEvent := NewUntypedEventWrapper(
		"ExampleID",
		"test",
		time.Now(),
		"test-payload",
		"Test payload",
	)

	forwarder, _ := NewForwarder(ctx, ForwarderConfiguration{})
	forwarder.RegisterRule(ctx, "test-payload", testHandler)
	forwarder.RegisterRule(ctx, "another-payload", testHandler2)

	assert.Len(t, forwarder.rules, 2)
	assert.NoError(t, forwarder.rules["test-payload"](ctx, exampleEvent))
	assert.Error(t, forwarder.rules["another-payload"](ctx, exampleEvent))

}

func TestForward(t *testing.T) {

	ctx := context.Background()

	testHandler := func(ctx context.Context, wrappers ...UntypedEventWrapper) error {
		return nil
	}

	testHandler2 := func(ctx context.Context, wrappers ...UntypedEventWrapper) error {
		return errors.New("test")
	}

	forwarder, _ := NewForwarder(ctx, ForwarderConfiguration{})
	forwarder.RegisterRule(ctx, "test-payload", testHandler)
	forwarder.RegisterRule(ctx, "test-payload-2", testHandler)
	forwarder.RegisterRule(ctx, "test-error-payload", testHandler2)

	exampleEvent := NewUntypedEventWrapper(
		"ExampleID",
		"test",
		time.Now(),
		"test-payload",
		"Test payload",
	)

	exampleEvent2 := NewUntypedEventWrapper(
		"ExampleID2",
		"test",
		time.Now(),
		"test-payload-2",
		"Test payload 2",
	)

	errorEvent2 := NewUntypedEventWrapper(
		"ExampleIDError",
		"test",
		time.Now(),
		"test-error-payload",
		"Test payload with error",
	)

	unregisteredEvent := NewUntypedEventWrapper(
		"UnregistedID",
		"test",
		time.Now(),
		"test-payload-3",
		"Test Unregistered payload",
	)

	assert.Error(t, forwarder.Forward(ctx))
	assert.NoError(t, forwarder.Forward(ctx, exampleEvent))
	assert.NoError(t, forwarder.Forward(ctx, exampleEvent, exampleEvent))
	assert.Error(t, forwarder.Forward(ctx, exampleEvent, exampleEvent2))
	assert.Error(t, forwarder.Forward(ctx, errorEvent2))
	assert.Error(t, forwarder.Forward(ctx, unregisteredEvent))

}

func TestForwardWithWildcard(t *testing.T) {

	ctx := context.Background()

	testHandler := func(ctx context.Context, wrappers ...UntypedEventWrapper) error {
		return nil
	}

	forwarder, _ := NewForwarder(ctx, ForwarderConfiguration{})
	forwarder.RegisterRule(ctx, "test-payload", testHandler)
	forwarder.RegisterRule(ctx, "*", testHandler)

	exampleEvent := NewUntypedEventWrapper(
		"ExampleID",
		"test",
		time.Now(),
		"test-payload",
		"Test payload",
	)

	unregisteredEvent := NewUntypedEventWrapper(
		"UnregistedID",
		"test",
		time.Now(),
		"test-payload-3",
		"Test Unregistered payload",
	)

	assert.NoError(t, forwarder.Forward(ctx, exampleEvent))
	assert.NoError(t, forwarder.Forward(ctx, unregisteredEvent))

}
