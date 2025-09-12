package message_factory

import (
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/logger"

	messagewrapper "github.com/pixie-sh/core-go/infra/message_wrapper"
	pixieErrors "github.com/pixie-sh/core-go/pkg/errors"
	"github.com/pixie-sh/core-go/pkg/models/serializer"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/utils"
)

func RegisterMessage[T any](forceValidations bool, customFactory ...*Factory) {
	var f = Singleton

	if len(customFactory) > 0 && customFactory[0] != nil {
		f = customFactory[0]
	}

	pt := types.PayloadTypeOf[T]()
	f.knownMessages[pt] = PackEntry[T](forceValidations)
}

func RegisterMessageCustomType[T any](customType types.PayloadType, forceValidations bool, customFactory ...*Factory) {
	var f = Singleton
	if len(customFactory) > 0 && customFactory[0] != nil {
		f = customFactory[0]
	}

	up := PackEntry[T](forceValidations)
	up.MessageType = customType
	f.knownMessages[customType] = up
}

func RegisterPack(pack Pack, customFactory ...*Factory) {
	f := Singleton
	if len(customFactory) > 0 && customFactory[0] != nil {
		f = customFactory[0]
	}

	// ctx logger instead?
	log := logger.Logger

	if knownPack, ok := f.knownPacks[pack.Name]; ok {

		err := errors.
			New("Pack %s is already registered", knownPack).
			WithErrorCode(pixieErrors.MessageFactoryDuplicateRegistrationErrorCode)

		log.With("error", err).Warn("Registered pack will be ignored")
		return
	}

	for _, entry := range pack.Entries {
		_, ok := f.knownMessages[entry.MessageType]
		if ok {
			err := errors.
				New("payload type %s already registered; pack: %s", entry.MessageType, pack.Name).
				WithErrorCode(pixieErrors.MessageFactoryDuplicateRegistrationErrorCode)

			log.With("error", err).Warn("Registered payload type will be ignored")

		}

		f.knownMessages[entry.MessageType] = entry
	}

	f.knownPacks[pack.Name] = pack
}

func PackEntry[T any](forceValidations ...bool) UntypedPackEntry {
	var t T

	validate := false
	if len(forceValidations) > 0 {
		validate = forceValidations[0]
	}

	pt := types.PayloadTypeOf[T]()
	return UntypedPackEntry{
		MessageType:      pt,
		Descriptions:     utils.SchemaDescriptions(t),
		ForceValidations: validate,
		FromBlob: func(blob []byte) (messagewrapper.UntypedMessage, error) {
			var msg messagewrapper.Message[T]
			err := serializer.Deserialize(blob, &msg, validate)
			if err != nil {
				return messagewrapper.UntypedMessage{}, err
			}

			return msg.Untyped(), err
		},
		Translate: func(fromPayload any) (any, error) {
			t, err := serializer.FromAny[T](fromPayload, validate)
			return t, err
		},
	}
}
