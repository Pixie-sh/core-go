package events

import (
	"context"

	"github.com/pixie-sh/errors-go"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
)

type singletonEmitter []Producer

var defaultEmitters singletonEmitter

func AddDefaultProducer(producer Producer) {
	defaultEmitters = append(defaultEmitters, producer)
}

// EmitBatch emits a batch of events to all producers
func EmitBatch(ctx context.Context, evs ...UntypedEventWrapper) error {
	if len(evs) == 0 {
		return errors.New("no events provided").WithErrorCode(errors.ProducerErrorCode)
	}

	if len(defaultEmitters) == 0 {
		return errors.New("no producers provided").WithErrorCode(errors.ProducerErrorCode)
	}

	var evsByType = map[string][]UntypedEventWrapper{}
	for _, ev := range evs {
		evsByType[ev.PayloadType] = append(evsByType[ev.PayloadType], ev)
	}

	log := pixiecontext.GetCtxLogger(ctx)
	for _, producer := range defaultEmitters {
		for _, evss := range evsByType {
			log.Debug("emitting batch... len(%d)", len(evss))
			err := producer.ProduceBatch(ctx, evss...)
			if err != nil {
				log.With("error", err).Error("error on event type '%s' batch emitting", evss[0].PayloadType)
				return err
			}
		}
	}

	return nil
}

// Emit emits a single event to all producers
func Emit(ctx context.Context, evs ...UntypedEventWrapper) error {
	if len(evs) == 0 {
		return errors.New("no events provided").WithErrorCode(errors.ProducerErrorCode)
	}

	if len(defaultEmitters) == 0 {
		return errors.New("no producers provided").WithErrorCode(errors.ProducerErrorCode)
	}

	log := pixiecontext.GetCtxLogger(ctx)
	for _, producer := range defaultEmitters {
		for _, ev := range evs {
			log.Debug("emitting event '%s'", ev.ID)
			err := producer.Produce(ctx, ev)
			if err != nil {
				log.With("error", err).Error("error on event '%s'(%s) emitting", ev.ID, ev.PayloadType)
				return err
			}
		}
	}

	return nil
}
