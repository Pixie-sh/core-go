package events

import (
	"context"
	"time"

	messagewrapper "github.com/pixie-sh/core-go/infra/message_wrapper"
	"github.com/pixie-sh/core-go/infra/uidgen"
)

// Event generic comm event data
// the event payload is meant to be immutable. keep in mind it's not using pointers
type Event[T any] struct {
	UntypedEventWrapper

	Payload messagewrapper.Message[T] `json:"-"`
}

// NewEventWrapper create event wrapper that holds a UntypedMessage
// the event payload is meant to be immutable. keep in mind it's not using pointers
func NewEventWrapper[T any](ID string, payloadType string, payload T) Event[T] {
	um := messagewrapper.NewUntypedMessage(ID, payloadType, payload)
	e := Event[T]{
		UntypedEventWrapper{um, make([]Producer, 0), true},
		messagewrapper.MessageOf[T](context.Background(), um),
	}

	e.Timestamp = time.Now().UTC()
	e.WithSender(uidgen.SystemUUID)
	return e
}

func (e *Event[T]) SetHeader(key string, val string) *Event[T] {
	e.UntypedEventWrapper.SetHeader(key, val)
	return e
}

func (e *Event[T]) WithProducer(p ...Producer) Emitter {
	e.UntypedEventWrapper.WithProducer(p...)
	return e
}

func (e Event[T]) Emit(ctx context.Context, useDefault ...bool) error {
	return e.UntypedEventWrapper.Emit(ctx, useDefault...)
}

func (e Event[T]) EmitAsync(ctx context.Context, useDefault ...bool) error {
	return e.UntypedEventWrapper.EmitAsync(ctx, useDefault...)
}

func (e *Event[T]) WithTo(to ...string) *Event[T] {
	e.UntypedEventWrapper.WithTo(to...)
	return e
}

func (e *Event[T]) WithSender(senderID string) *Event[T] {
	e.FromSenderID = senderID
	return e
}
