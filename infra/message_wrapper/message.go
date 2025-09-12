package message_wrapper

import (
	"context"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/models/serializer"
	"github.com/pixie-sh/core-go/pkg/types"
)

type messageData interface{}

// Message abstraction that holds a specific payload and known how to convert itself to untyped
// it's not serializable and it's meant to be immutable
type Message[T messageData] struct {
	um UntypedMessage

	data T
	err  errors.E
}

func (m *Message[T]) Data() T {
	return m.data
}

func (m *Message[T]) Untyped() UntypedMessage {
	um := m.um
	um.Payload = m.data
	um.Error = m.err
	return um
}

func (m *Message[T]) Error(set ...errors.E) errors.E {
	if len(set) > 0 {
		m.err = set[0]
		return m.err
	}

	return m.err
}

func (m *Message[T]) MarshalJSON() ([]byte, error) {
	um := m.Untyped()
	return serializer.Serialize(um)
}

func (m *Message[T]) UnmarshalJSON(blob []byte) error {
	var um UntypedMessage
	err := serializer.Deserialize(blob, &um)
	if err != nil {
		return err
	}

	t, err := serializer.FromAny[T](um.Payload, false)
	if err != nil {
		return err
	}

	um.Payload = t
	*m = MessageOf[T](context.Background(), um)
	return nil
}

func MessageOf[T any](ctx context.Context, fromUntyped UntypedMessage) Message[T] {
	castedPayload, ok := fromUntyped.Payload.(T)
	if !ok {
		err := errors.New("unable to cast %s to %s", fromUntyped.Type(), types.NameOf(fromUntyped.Payload)).WithErrorCode(errors.InvalidTypeErrorCode)
		logger.Logger.WithCtx(ctx).With("error", err).Error("error casting payload")
		panic(err)
	}

	return Message[T]{
		um:   fromUntyped,
		data: castedPayload,
		err:  fromUntyped.Error,
	}
}
