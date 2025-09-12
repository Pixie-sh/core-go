package events

import (
	"context"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
)

// LoggerProducer is a default implementation of a producer that uses the LoggerInterface to print the event
type LoggerProducer struct {
	id string
}

func (lp *LoggerProducer) ID() string {
	return lp.id
}

func NewLoggerProducer(_ context.Context, id string) *LoggerProducer {
	return &LoggerProducer{
		id: id,
	}
}

func (lp *LoggerProducer) ProduceBatch(ctx context.Context, wrappers ...UntypedEventWrapper) error {
	tlog := pixiecontext.GetCtxLogger(ctx).With("producer_id", lp.id)
	for _, wrapper := range wrappers {
		tlog.With("event", wrapper.UntypedMessage).Log("LoggerProducer with event %s", wrapper.ID)
	}

	return nil
}

func (lp *LoggerProducer) Produce(ctx context.Context, wrapper UntypedEventWrapper) error {
	pixiecontext.GetCtxLogger(ctx).
		With("producer_id", lp.id).
		With("event", wrapper.UntypedMessage).
		Log("LoggerProducer with event %s", wrapper.ID)
	return nil
}
