package io

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pixie-sh/errors-go"

	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
)

// AsyncWriter provides an asynchronous io.Writer implementation using channels.
type AsyncWriter struct {
	writer io.Writer
	ch     chan []byte
	wg     sync.WaitGroup
	ctx    context.Context

	errMu sync.RWMutex
	err   error
}

// NewAsyncWriter creates a new AsyncWriter wrapping the provided io.Writer.
// bufferSize determines the size of the internal channel buffer.
func NewAsyncWriter(ctx context.Context, w io.Writer, bufferSize int) *AsyncWriter {
	var ch chan []byte
	if bufferSize == -1 {
		ch = make(chan []byte)
	} else {
		ch = make(chan []byte, bufferSize)
	}

	aw := &AsyncWriter{
		writer: w,
		ch:     ch,
		ctx:    ctx,
	}

	go aw.writeLoop()
	return aw
}

// writeLoop is the goroutine loop that processes writes from the channel.
func (aw *AsyncWriter) writeLoop() {
	aw.wg.Add(1)
	defer func() {
		err := aw.close()
		if err != nil {
			pixiecontext.GetCtxLogger(aw.ctx).
				With("err", err).
				Error("async writer: write loop exited")
		}
	}()
	defer aw.wg.Done()

loop:
	for {
		select {
		case <-aw.ctx.Done():
			// Context cancelled - drain any remaining messages from the channel
			for {
				select {
				case data := <-aw.ch:
					// Process remaining data even though context is cancelled
					if aw.getError() == nil {
						if _, err := aw.writer.Write(data); err != nil {
							aw.setError(err)
						}
					}
				default:
					// No more data in channel, exit
					return
				}
			}
		case data := <-aw.ch:
			if aw.getError() != nil {
				break loop
			}

			if _, err := aw.writer.Write(data); err != nil {
				aw.setError(err)
			}
		}
	}
}

// Write implements the io.Writer interface asynchronously.
// It copies the data and sends it through a channel for the writeLoop to process.
func (aw *AsyncWriter) Write(p []byte) (n int, err error) {
	if err = aw.getError(); err != nil {
		return 0, err
	}

	// Copy the input slice, as it may be modified after the call.
	data := make([]byte, len(p))
	copy(data, p)

	select {
	case aw.ch <- data:
		return len(p), nil
	case <-time.After(5 * time.Millisecond):
		return 0, errors.New("async writer: write timeout; buffer's full")
	}
}

// getError retrieves the current error in a thread-safe manner.
func (aw *AsyncWriter) getError() error {
	aw.errMu.RLock()
	defer aw.errMu.RUnlock()
	return aw.err
}

// setError sets the first error encountered during writing.
func (aw *AsyncWriter) setError(err error) {
	aw.errMu.Lock()
	defer aw.errMu.Unlock()
	if aw.err == nil {
		aw.err = err
	}
}

// Close closes the AsyncWriter, waits for the writeLoop to finish,
// and returns any error encountered during writes.
func (aw *AsyncWriter) close() error {
	aw.wg.Wait()
	return aw.getError()
}
