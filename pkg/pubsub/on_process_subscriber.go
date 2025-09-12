package pubsub

import (
	"context"
	"time"

	"github.com/pixie-sh/core-go/pkg/types/channels"

	"github.com/pixie-sh/core-go/pkg/types"

	"github.com/pixie-sh/logger-go/logger"
)

type OnProcessSubscriber[T any] struct {
	id          string
	queue       chan T
	processFunc func(T)
}

// NewOnProcessSubscriber creates a new OnProcessSubscriber
func NewOnProcessSubscriber[T any](ctx context.Context, id string, processFunc func(T), queueSize int) *OnProcessSubscriber[T] {
	var f = processFunc
	if types.Nil(processFunc) {
		f = func(t T) {
			logger.With("value", t).Debug("empty process func.")
		}
	}

	onProcessSubscriber := &OnProcessSubscriber[T]{id: id, queue: make(chan T, queueSize), processFunc: f}
	go channels.ConsumeChannel[T](ctx, onProcessSubscriber.queue, onProcessSubscriber.processFunc, true)
	return onProcessSubscriber
}

// ID return subscriber ID
func (s *OnProcessSubscriber[T]) ID() string {
	return s.id
}

// Publish method implementation, not blocking
func (s *OnProcessSubscriber[T]) Publish(msg T) {
	channels.PublishToChannelWithTimeout[T](s.queue, msg, 5*time.Second)
}
