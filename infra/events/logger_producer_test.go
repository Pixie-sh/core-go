package events

import (
	"testing"

	"github.com/pixie-sh/errors-go"
	"github.com/stretchr/testify/assert"

	"github.com/pixie-sh/core-go/infra/message_wrapper"
)

func TestNilness(t *testing.T) {
	lp := NewLoggerProducer(nil, "test")

	var um UntypedEventWrapper
	assert.NoError(t, lp.Produce(nil, um))

	um.Error = nil
	assert.NoError(t, lp.Produce(nil, um))

	um.Error = errors.New("test error")
	assert.NoError(t, lp.Produce(nil, um))
}

func TestNilnessBatch(t *testing.T) {
	lp := NewLoggerProducer(nil, "test")

	var um []UntypedEventWrapper
	for i := 0; i < 10; i++ {
		um = append(um, UntypedEventWrapper{UntypedMessage: message_wrapper.UntypedMessage{
			Error: errors.New("test error"),
		}})
		um = append(um, UntypedEventWrapper{})
	}

	assert.NoError(t, lp.ProduceBatch(nil, um...))
}
