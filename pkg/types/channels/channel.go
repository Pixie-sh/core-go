package channels

import (
	"context"
	"time"

	"github.com/pixie-sh/logger-go/logger"
)

// ConsumeChannel blocking for loop to consume provided channel
// includes draining once context is canceled as default
func ConsumeChannel[T any](ctxWithCancel context.Context, chnToDrain chan T, handler func(T), withDrain ...bool) {
loop:
	for {
		select {
		case <-ctxWithCancel.Done():
			break loop
		case src := <-chnToDrain:
			handler(src)
		}
	}

	if len(withDrain) == 0 || len(withDrain) > 0 && withDrain[0] {
		for src := range chnToDrain {
			logger.Logger.With("drained_from_channel", src).Debug("context canceled; drain channel")
		}
	}
}

func ChanPublisherPanicHandler() {
	if r := recover(); r != nil {
		logger.Logger.Error("chan producer recovered from panic: %+v", r)
	}
}

func PublishToChannelWithTimeout[T any](chnToPub chan<- T, payload T, timeoutDuration time.Duration) {
	timeout := time.After(timeoutDuration)
	defer ChanPublisherPanicHandler()

waitLoop:
	select {
	case <-timeout:
		logger.Logger.Error("timeout reach publishing to chan")
		break waitLoop
	default:
		chnToPub <- payload
		break waitLoop
	}
}
