package message_factory

import (
	"context"

	"github.com/pixie-sh/core-go/infra/message_wrapper"
	"github.com/pixie-sh/core-go/pkg/types/maps"

	"github.com/pixie-sh/core-go/pkg/types"

	"github.com/pixie-sh/errors-go"

	"github.com/pixie-sh/core-go/pkg/models/serializer"
)

var Singleton = NewFactory()

type Factory struct {
	// If we want to handle payload types individualy instead
	// of packs, we can just type the specific event
	knownMessages map[types.PayloadType]UntypedPackEntry

	// All packs that are known in the factory
	// a pack always contains a payload type, but a payload type does not
	// need to belong in a pack
	knownPacks map[string]Pack
}

func NewFactory() *Factory {
	f := &Factory{
		knownMessages: make(map[types.PayloadType]UntypedPackEntry),
		knownPacks:    make(map[string]Pack),
	}

	return f
}

func (f *Factory) GetRegisteredEvents() map[types.PayloadType]UntypedPackEntry {
	return f.knownMessages
}

func (f *Factory) GetRegisteredPacks() map[string]Pack {
	return f.knownPacks
}

// Create deserializes a byte array into an UntypedMessage by extracting the payload type
// and using the corresponding registered message handler.
//
// Parameters:
//   - ctx: Context for the operation (currently unused)
//   - blob: Raw byte array containing the serialized message with a payload_type field
//
// Returns:
//   - message_wrapper.UntypedMessage: The deserialized message if successful
//   - error: Returns an error if:
//   - The blob cannot be deserialized with error {serializer.Deserialize error}
//   - The payload type is not registered in the factory with error {errors with FieldErrors and payload at 'payload_type'}
//   - The message handler fails to process the blob with error {Specific registration error}
func (f *Factory) Create(_ context.Context, blob []byte) (message_wrapper.UntypedMessage, error) {
	var payloadTypeOnly struct {
		PayloadType string `json:"payload_type" validate:"required"`
	}

	var err error
	var msg message_wrapper.UntypedMessage

	err = serializer.Deserialize(blob, &payloadTypeOnly, true)
	if err != nil {
		return message_wrapper.UntypedMessage{}, err
	}

	entry, ok := f.knownMessages[types.PayloadType(payloadTypeOnly.PayloadType)]
	if !ok {
		return message_wrapper.UntypedMessage{}, errors.NewValidationError("event type not registered", &errors.FieldError{
			Field:   "payload_type",
			Rule:    "invalidPayloadType",
			Message: "event type " + payloadTypeOnly.PayloadType + " not registered",
		}).WithErrorCode(errors.InvalidTypeErrorCode)
	}

	msg, err = entry.FromBlob(blob)
	if err != nil {
		return message_wrapper.UntypedMessage{}, err
	}

	return msg, nil
}

// CreateFromString string to byte array deserialized into an UntypedMessage by extracting the payload type
// and using the corresponding registered message handler.
//
// Parameters:
//   - ctx: Context for the operation (currently unused)
//   - blob: Raw byte array containing the serialized message with a payload_type field
//
// Returns:
//   - message_wrapper.UntypedMessage: The deserialized message if successful
//   - error: Returns an error if:
//   - The blob cannot be deserialized with error {serializer.Deserialize error}
//   - The payload type is not registered in the factory with error {errors with FieldErrors and payload at 'payload_type'}
//   - The message handler fails to process the blob with error {Specific registration error}
func (f *Factory) CreateFromString(ctx context.Context, blob string) (message_wrapper.UntypedMessage, error) {
	return f.Create(ctx, types.UnsafeBytes(blob))
}

func (f *Factory) GetRegisteredTypes() []types.PayloadType {
	evTypes := maps.MapStructValue(maps.MapValues(f.knownMessages), func(event UntypedPackEntry) types.PayloadType {
		return event.MessageType
	})

	return evTypes
}

func (f *Factory) Exists(payloadType types.PayloadType) bool {
	_, ok := f.knownMessages[payloadType]
	return ok
}

// Translate although it returns any it has the proper underlying type from registrations
func (f *Factory) Translate(payloadType string, fromPayload any) (any, error) {
	entry, ok := f.knownMessages[types.PayloadType(payloadType)]
	if !ok {
		return nil, errors.New("event type %s not registered", payloadType)
	}

	msg, err := entry.Translate(fromPayload)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// TranslateMap uses Translate function to have underlying types on the map values
func (f *Factory) TranslateMap(payloadType string, payload any) (map[string]interface{}, error) {
	typed, err := f.Translate(payloadType, payload)
	if err != nil {
		return nil, err
	}

	values, err := serializer.StructToMap(typed, false)
	if err != nil {
		return nil, err
	}

	//fucking map structure is ducking us
	values["payload_type"] = payloadType
	return values, nil
}
