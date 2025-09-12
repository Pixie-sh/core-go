package events

import (
	"context"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/infra/events/forwarder_errors"
	"github.com/pixie-sh/core-go/pkg/types"
)

type ForwardRuleHandler func(ctx context.Context, wrappers ...UntypedEventWrapper) error

type ForwarderConfiguration struct{}

type Forwarder struct {
	config ForwarderConfiguration
	rules  map[string]ForwardRuleHandler
}

// NewForwarder creates a new instance of Forwarder.
// this will execute the forwarding based on payload type and registered destinations
func NewForwarder(_ context.Context, config ForwarderConfiguration) (Forwarder, error) {
	return Forwarder{
		config: config,
		rules:  make(map[string]ForwardRuleHandler),
	}, nil
}

// Forward is a list is passed only the first payloadTypes is used for forwarding decisions
func (f *Forwarder) Forward(ctx context.Context, evs ...UntypedEventWrapper) error {
	if len(evs) == 0 {
		return errors.New("no events to forward").WithErrorCode(forwarder_errors.ForwarderEmptyListErrorCode)
	}

	e0 := evs[0]
	if types.Nil(e0) {
		return errors.New("first event is nil").WithErrorCode(forwarder_errors.ForwarderNilEventErrorCode)
	}

	//iterate over all events and return error if the payload types mismatch
	if len(evs) > 1 {
		for i, ev := range evs[1:] {
			if ev.PayloadType != e0.PayloadType {
				return errors.New("payload type mismatch at expected from index 0: %s; got %s from index %d ", e0.PayloadType, ev.PayloadType, i).
					WithErrorCode(forwarder_errors.ForwarderPayloadTypeMismatchErrorCode)
			}
		}
	}

	// Handle in case of wildcard
	h, ok := f.rules[EventTypesWildcard]
	if ok {
		logger.Log("Forwarding rule on wildcard %s", e0.PayloadType)
		return h(ctx, evs...)
	}

	// Check for specific rule
	h, ok = f.rules[e0.PayloadType]
	if ok {
		return h(ctx, evs...)
	}

	// If there's no specific rule, check if there's a fallback rule
	h, ok = f.rules[EventTypesFallback]
	if ok {
		return h(ctx, evs...)
	}

	return errors.New("rule for event type %s not registered", e0.PayloadType).WithErrorCode(forwarder_errors.ForwarderTypeNotRegisteredErrorCode)
}

// RegisterRule only the last handler by payloadType will be stored
func (f *Forwarder) RegisterRule(_ context.Context, payloadType string, handler ForwardRuleHandler) {
	f.rules[payloadType] = handler
}
