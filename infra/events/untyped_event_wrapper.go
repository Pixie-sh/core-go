package events

import (
	"context"
	"time"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/errors-go/utils"

	"github.com/pixie-sh/core-go/infra/message_wrapper"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
)

type UntypedEventWrapper struct {
	message_wrapper.UntypedMessage

	producers                  []Producer //producers are not appended, the list is replaced with provided one
	clearProducerAfterProduced bool
}

// NewUntypedEventWrapper create event wrapper that holds a UntypedMessage
// the event payload is meant to be immutable. keep in mind it's no using pointers
func NewUntypedEventWrapper(ID string, fromSenderID string, timestamp time.Time, payloadType string, payload any) UntypedEventWrapper {
	return UntypedEventWrapper{
		message_wrapper.NewFullUntypedMessage(ID, fromSenderID, timestamp, payloadType, payload),
		make([]Producer, 0),
		true,
	}
}

// NewUntypedEventWrapperFromMessage create an event wrapper from message wrapper
// the event payload is meant to be immutable. keep in mind it's no using pointers
func NewUntypedEventWrapperFromMessage(message message_wrapper.UntypedMessage) UntypedEventWrapper {
	return UntypedEventWrapper{
		message,
		make([]Producer, 0),
		true,
	}
}

func (ue *UntypedEventWrapper) EmitAsync(ctx context.Context, useDefault ...bool) error {
	if utils.Nil(ue) {
		return errors.New("invalid event pointer, unable to emit asynchronously")
	}

	go func() {
		err := ue.Emit(ctx, useDefault...)
		if err != nil {
			pixiecontext.GetCtxLogger(ctx).With("error", err).Error("failed to emit event asynchronously")
		}
	}()

	return nil
}

func (ue *UntypedEventWrapper) Emit(ctx context.Context, useDefault ...bool) error {
	if utils.Nil(ue) {
		return errors.New("invalid event pointer, unable to emit").WithErrorCode(errors.ProducerErrorCode)
	}

	log := pixiecontext.GetCtxLogger(ctx).With("event", ue)
	log.Debug("emitting event %s(%s)", ue.PayloadType, ue.ID)
	defer log.Debug("finished emitting event %s(%s)", ue.PayloadType, ue.ID)

	err := ue.Validate() // TODO: shall we?
	if err != nil {
		return errors.NewWithError(err, "validation failed at emit").WithErrorCode(errors.ProducerErrorCode)
	}

	if len(ue.producers) == 0 && len(defaultEmitters) == 0 {
		log.Warn("no default producer nor customs. unable to emit.")
		return errors.New("producers not defined. unable to emit.").WithErrorCode(errors.ProducerErrorCode)
	}

	var errorList []error
	if len(defaultEmitters) > 0 &&
		(len(ue.producers) == 0 ||
			(len(useDefault) > 0 && useDefault[0])) {
		for _, producer := range defaultEmitters {
			if utils.Nil(producer) {
				log.Warn("default emitter is nil. unable to emit.")
				continue
			}

			err = producer.Produce(ctx, *ue)
			if err != nil {
				errorList = append(errorList, err)
			}
		}
	}

	if len(ue.producers) > 0 {
		for _, producer := range ue.producers {
			if utils.Nil(producer) {
				log.Warn("provided emitter is nil. unable to emit.")
				continue
			}

			err = producer.Produce(ctx, *ue)
			if err != nil {
				errorList = append(errorList, err)
			}
		}

		if ue.clearProducerAfterProduced {
			ue.producers = nil
		}
	}

	if len(errorList) > 0 {
		return errors.New("errors producing events").
			WithErrorCode(errors.ProducerErrorCode).
			WithNestedError(errorList...)
	}
	return nil
}

// WithProducer the producers are not appended, the list is replaced with provided one
// default behaviour is to clear the list after emit the event
func (ue *UntypedEventWrapper) WithProducer(p ...Producer) *UntypedEventWrapper {
	ue.producers = p
	return ue
}

// WithClearProducers set to false to override default behaviour of clearing producers after emit
func (ue *UntypedEventWrapper) WithClearProducers(resetProducers bool) *UntypedEventWrapper {
	ue.clearProducerAfterProduced = resetProducers
	return ue
}

func (ue *UntypedEventWrapper) SetHeader(key string, val string) *UntypedEventWrapper {
	ue.UntypedMessage.SetHeader(key, val)
	return ue
}

func (ue *UntypedEventWrapper) WithTo(to ...string) *UntypedEventWrapper {
	ue.UntypedMessage.To = append(ue.UntypedMessage.To, to...)
	return ue
}

func (ue *UntypedEventWrapper) ClearTo() *UntypedEventWrapper {
	ue.UntypedMessage.To = []string{}
	return ue
}
