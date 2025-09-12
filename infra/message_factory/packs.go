package message_factory

import (
	"context"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
)

// RegisterPacks registers the provided message packs into the message factory.
// It takes a factory pointer and a variable number of message packs as parameters.
// Each pack is registered in the factory, and the registered event models are logged
// for debugging purposes.
func RegisterPacks(ctx context.Context, packs []Pack, factories ...*Factory) {
	factory := Singleton
	if len(factories) > 0 {
		factory = factories[0]
	}

	log := pixiecontext.GetCtxLogger(ctx)
	log.Debug("Registering message_packs...")

	for _, pack := range packs {
		RegisterPack(pack, factory)
	}

	log.
		With("registered_events", Singleton.GetRegisteredTypes()).
		Debug("Registered event models")
}
