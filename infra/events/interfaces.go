package events

import (
	"context"

	"github.com/pixie-sh/logger-go/logger"
)

// Emitter uses the Producer to emit itself using producer
type Emitter interface {
	WithProducer(p ...Producer) Emitter
	Emit(ctx context.Context, useDefault ...bool) error
	EmitAsync(ctx context.Context, useDefault ...bool) error
}

// Producer used by Emitter.Emit to produce itself
type Producer interface {
	ID() string
	ProduceBatch(ctx context.Context, wrapper ...UntypedEventWrapper) error
	Produce(ctx context.Context, wrapper UntypedEventWrapper) error
}

// Consumer must be implemented by consumers of Events
type Consumer interface {
	ConsumeBatch(ctx context.Context, handler func(context.Context, UntypedEventWrapper) error) error
	Consume(ctx context.Context, handler func(context.Context, UntypedEventWrapper) error) error
}

func ConsumerPanicHandler() {
	if r := recover(); r != nil {
		logger.Logger.Error("consumer recovered from panic: %+v", r)
	}
}
