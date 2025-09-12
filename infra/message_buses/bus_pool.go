package message_buses

import (
	"context"
	"sync"
	"time"

	"github.com/pixie-sh/logger-go/logger"
)

type busPool struct {
	mu    sync.RWMutex
	buses map[string]MessageBus
}

func NewBusPool(ctx context.Context) BusPool {
	bp := &busPool{
		buses: make(map[string]MessageBus),
	}

	go bp.busesLookup(ctx, time.Second*120)
	return bp
}

func (b *busPool) Get(ctx context.Context, key string) MessageBus {
	b.mu.Lock()
	defer b.mu.Unlock()

	bus, ok := b.buses[key]
	if ok {
		return bus
	}

	//create bus
	b.buses[key] = NewOnProcessMessageBus(ctx)
	return b.buses[key]
}

func (b *busPool) busesLookup(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case tick := <-ticker.C:
			logger.Logger.Debug("busesLookup %v", tick)
			b.mu.Lock()
			for key, bus := range b.buses {
				if bus.countSubscriptions() == 0 {
					logger.Logger.Debug("deleting bus %s", key)
					delete(b.buses, key)
				}
			}
			b.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
