package pubsub

import (
	"context"
	"sync"

	"github.com/pixie-sh/errors-go/utils"
)

type GenericPublisher[T any] struct {
	mu            sync.RWMutex
	subscriptions map[string]Subscriber[T]
}

// NewGenericPublisher creates a new GenericPublisher
func NewGenericPublisher[T any](_ context.Context) *GenericPublisher[T] {
	as := &GenericPublisher[T]{subscriptions: make(map[string]Subscriber[T])}
	return as
}

// Subscribe method implementation
func (s *GenericPublisher[T]) Subscribe(subscriber Subscriber[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscriptions[subscriber.ID()] = subscriber
}

func (s *GenericPublisher[T]) CountSubscriptions() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subscriptions)
}

func (s *GenericPublisher[T]) NotifySubscriptions(msg T) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sub := range s.subscriptions {
		if !utils.Nil(sub) {
			sub.Publish(msg)
		}
	}
}

func (s *GenericPublisher[T]) NotifySubscriptionsWithFrom(fromID string, msg T) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id, sub := range s.subscriptions {
		if id == fromID {
			continue
		}

		if !utils.Nil(sub) {
			sub.Publish(msg)
		}
	}
}

func (s *GenericPublisher[T]) Unsubscribe(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, found := s.subscriptions[id]
	if found {
		delete(s.subscriptions, id)
	}
}

func (s *GenericPublisher[T]) UnsubscribeAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id := range s.subscriptions {
		delete(s.subscriptions, id)
	}
}
