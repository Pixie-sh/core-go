package io

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeWriter is used to simulate an underlying writer.
// It includes synchronization to prevent data races during concurrent access.
type fakeWriter struct {
	mu  sync.RWMutex
	buf bytes.Buffer
}

func (fw *fakeWriter) Write(p []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.buf.Write(p)
}

func (fw *fakeWriter) String() string {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.buf.String()
}

func TestAsyncWriter_ConcurrentWrite(t *testing.T) {
	fw := &fakeWriter{}
	// Create a context that lives long enough for the test to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Increase channel buffer size if needed.
	aw := NewAsyncWriter(ctx, fw, 100)
	defer func() {
		_ = aw.close()
		aw.wg.Wait()
	}()

	const numWriters = 50
	const writesPerGoroutine = 20

	var wg sync.WaitGroup
	wg.Add(numWriters)

	for i := 0; i < numWriters; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				msg := fmt.Sprintf("goroutine-%d-write-%d;", id, j)
				_, err := aw.Write([]byte(msg))
				if err != nil {
					t.Errorf("unexpected error from write (goroutine %d, write %d): %v", id, j, err)
					// break out early to ensure wg.Done is called and test doesn't hang waiting.
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to finish writing.
	wg.Wait()

	// Cancel the context to signal shutdown
	cancel()

	// Wait for the AsyncWriter to properly close and flush all pending writes
	err := aw.close()
	if err != nil {
		t.Fatalf("unexpected error during close: %v", err)
	}

	// Now it's safe to read the final output
	output := fw.String()
	for i := 0; i < numWriters; i++ {
		for j := 0; j < writesPerGoroutine; j++ {
			expectedSubstring := fmt.Sprintf("goroutine-%d-write-%d;", i, j)
			if !strings.Contains(output, expectedSubstring) {
				t.Errorf("expected output to contain %q, but it was missing", expectedSubstring)
			}
		}
	}
}

func TestAsyncWriter_ConcurrentWriteDuringClose(t *testing.T) {
	fw := &fakeWriter{}
	ctx, cancel := context.WithCancel(context.Background())

	aw := NewAsyncWriter(ctx, fw, 10)

	const numWriters = 10
	const writesPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numWriters)

	// Start concurrent writers
	for i := 0; i < numWriters; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				msg := fmt.Sprintf("writer-%d-msg-%d;", id, j)
				_, err := aw.Write([]byte(msg))
				if err != nil {
					// It's okay to get timeout errors during shutdown
					if !strings.Contains(err.Error(), "timeout") {
						t.Errorf("unexpected error from write (writer %d, msg %d): %v", id, j, err)
					}
				}
				// Small delay to interleave writes and close
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// Let some writes happen, then close
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for all writers to finish
	wg.Wait()

	// Wait for AsyncWriter to finish
	err := aw.close()
	if err != nil {
		t.Fatalf("unexpected error during close: %v", err)
	}

	// Verify that we got some output (data race fix allows safe reading)
	output := fw.String()
	if len(output) == 0 {
		t.Error("expected some output from concurrent writes")
	}

	// Verify output contains expected patterns
	if !strings.Contains(output, "writer-") || !strings.Contains(output, "-msg-") {
		t.Errorf("output doesn't contain expected patterns: %q", output)
	}
}
