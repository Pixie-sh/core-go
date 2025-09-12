package events

import (
	"context"
	"testing"
	"time"

	"github.com/pixie-sh/errors-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/pixie-sh/core-go/infra/message_factory"
	"github.com/pixie-sh/core-go/pkg/types"
	"github.com/pixie-sh/core-go/pkg/types/slices"
)

// MockProducer is a mock implementation of the Producer interface
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) ProduceBatch(ctx context.Context, wrapper ...UntypedEventWrapper) error {
	errrs := slices.Map(wrapper, func(item UntypedEventWrapper) error {
		return m.Produce(ctx, item)
	})

	return errors.Join(errrs...)
}

func (m *MockProducer) ID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProducer) Produce(ctx context.Context, wrapper UntypedEventWrapper) error {
	args := m.Called(ctx, wrapper)
	return args.Error(0)
}

func TestNewProducersPool(t *testing.T) {
	ctx := context.Background()
	config := ProducerPoolConfiguration{
		ProducerPoolID: "test-pool",
		SupportedPayloadTypesByProducerID: map[string][]string{
			"producer1": {"type1", "type2"},
			"producer2": {"*", "type3"},
		},
	}

	producer1 := new(MockProducer)
	producer1.On("ID").Return("producer1")

	producer2 := new(MockProducer)
	producer2.On("ID").Return("producer2")

	pool, err := NewProducersPool(ctx, config, producer1, producer2)

	assert.NoError(t, err)
	assert.Equal(t, "test-pool", pool.ID())
	assert.Len(t, pool.producersList, 2)
	assert.Len(t, pool.producersMapped, 3) // type1, type2, and EventTypesWildcard
	assert.Len(t, pool.producersMapped["type1"], 1)
	assert.Len(t, pool.producersMapped["type2"], 1)
	assert.Len(t, pool.producersMapped["type3"], 0) // Because wildcard already exists
	assert.Len(t, pool.producersMapped["*"], 1)
}

func TestNewProducersPoll_WithPacks(t *testing.T) {
	ctx := context.Background()
	config := ProducerPoolConfiguration{
		ProducerPoolID: "test-pool",
		SupportedPayloadTypesByProducerID: map[string][]string{
			"producer1": {"type1", "type2"},
			"producer2": {"*"},
			"producer3": {"type1"},
		},
		SupportedPacksByProducerID: map[string][]string{
			"producer1": {"TestMessagePack"},
			"producer2": {"TestToSeeIfThisIsIgnored"}, // Should be ignored in favour of wildcard
			"producer3": {"AnotherTestMessagePack"},
			"producer4": {"TestMessagePack"},
		},
	}
	testFactory := message_factory.Singleton
	testPack := message_factory.NewPack(
		"TestMessagePack",
		[]message_factory.UntypedPackEntry{
			{
				MessageType: types.PayloadType("first_payload_type"),
			},
			{
				MessageType: types.PayloadType("second_payload_type"),
			},
		}...,
	)

	anotherTestPack := message_factory.NewPack(
		"AnotherTestMessagePack",
		[]message_factory.UntypedPackEntry{
			{
				MessageType: types.PayloadType("another_payload_type"),
			},
		}...,
	)

	message_factory.RegisterPack(testPack, testFactory)
	message_factory.RegisterPack(anotherTestPack, testFactory)

	packs := testFactory.GetRegisteredPacks()
	dealMessagePack := packs["TestMessagePack"]
	partnersMessagePack := packs["AnotherTestMessagePack"]

	producer1 := new(MockProducer)
	producer1.On("ID").Return("producer1")

	producer2 := new(MockProducer)
	producer2.On("ID").Return("producer2")

	producer3 := new(MockProducer)
	producer3.On("ID").Return("producer3")

	producer4 := new(MockProducer)
	producer4.On("ID").Return("producer4")

	pool, err := NewProducersPool(ctx, config, producer1, producer2, producer3, producer4)

	assert.NoError(t, err)
	assert.Equal(t, "test-pool", pool.ID())
	assert.Len(t, pool.producersList, 4)
	// == Number of  producers Mapped is the number of events of each pack
	assert.Len(t, pool.producersMapped, len(dealMessagePack.Entries)+len(partnersMessagePack.Entries)+3)

	//  == The packs itself shouldn't be stored ==
	assert.Len(t, pool.producersMapped["TestMessagePack"], 0)
	assert.Len(t, pool.producersMapped["AnotherTestMessagePack"], 0)

	// == The events of the pack should be mapped ==
	assert.Len(t, pool.producersMapped["first_payload_type"], 2)
	assert.Len(t, pool.producersMapped["second_payload_type"], 2)
	assert.Len(t, pool.producersMapped["another_payload_type"], 1)

	// == Checks for other events, outside packs ==
	assert.Len(t, pool.producersMapped["type1"], 2)
	assert.Len(t, pool.producersMapped["type2"], 1)

	assert.Len(t, pool.producersMapped["*"], 1)
	assert.Len(t, pool.producersMapped["TestToSeeIfThisIsIgnored"], 0)
}

func TestNewProducersPool_WithNonExistentPack(t *testing.T) {

	ctx := context.Background()
	config := ProducerPoolConfiguration{
		ProducerPoolID: "test-pool",
		SupportedPayloadTypesByProducerID: map[string][]string{
			"producer1": {"type1", "type2"},
		},
		SupportedPacksByProducerID: map[string][]string{
			"producer1": {"NonExistentPack"},
		},
	}

	producer1 := new(MockProducer)
	producer1.On("ID").Return("producer1")

	_, err := NewProducersPool(ctx, config, producer1)

	assert.Error(t, err)

}

func TestProducersPool_Produce(t *testing.T) {
	ctx := context.Background()
	config := ProducerPoolConfiguration{
		ProducerPoolID: "test-pool",
		SupportedPayloadTypesByProducerID: map[string][]string{
			"producer1": {"type1"},
			"producer2": {"*"},
			"producer3": {"type1", "type2"},
			"producer4": {"type2"},
		},
	}

	producer1 := new(MockProducer)
	producer1.On("ID").Return("producer1")
	producer1.On("Produce", mock.Anything, mock.Anything).Return(nil)

	producer2 := new(MockProducer)
	producer2.On("ID").Return("producer2")
	producer2.On("Produce", mock.Anything, mock.Anything).Return(nil)

	producer3 := new(MockProducer)
	producer3.On("ID").Return("producer3")
	producer3.On("Produce", mock.Anything, mock.Anything).Return(nil)

	producer4 := new(MockProducer)
	producer4.On("ID").Return("producer4")
	producer4.On("Produce", mock.Anything, mock.Anything).Return(nil)

	pool, _ := NewProducersPool(ctx, config, producer1, producer2, producer3, producer4)

	wrapper := NewUntypedEventWrapper("id-type1", "sender-type1", time.Now().UTC(), "type1", []byte("test payload"))
	err := pool.Produce(ctx, wrapper)

	assert.NoError(t, err)
	producer1.AssertCalled(t, "Produce", ctx, wrapper)
	producer2.AssertCalled(t, "Produce", ctx, wrapper)
	producer3.AssertCalled(t, "Produce", ctx, wrapper)
	producer4.AssertNotCalled(t, "Produce", ctx, wrapper)

}

func TestProducersPool_ProduceBatch(t *testing.T) {
	ctx := context.Background()
	config := ProducerPoolConfiguration{
		ProducerPoolID: "test-pool",
		SupportedPayloadTypesByProducerID: map[string][]string{
			"producer1": {"type1", "type2"},
			"producer2": {"*"},
			"producer3": {"type2"},
		},
	}

	producer1 := new(MockProducer)
	producer1.On("ID").Return("producer1")
	producer1.On("Produce", mock.Anything, mock.Anything).Return(nil)

	producer2 := new(MockProducer)
	producer2.On("ID").Return("producer2")
	producer2.On("Produce", mock.Anything, mock.Anything).Return(nil)

	producer3 := new(MockProducer)
	producer3.On("ID").Return("producer3")
	producer3.On("Produce", mock.Anything, mock.Anything).Return(nil)

	pool, _ := NewProducersPool(ctx, config, producer1, producer2, producer3)

	wrappers := []UntypedEventWrapper{
		NewUntypedEventWrapper("id-type1", "sender-type1", time.Now().UTC(), "type1", []byte("test payload")),
		NewUntypedEventWrapper("id-type2", "sender-type2", time.Now().UTC(), "type2", []byte("test payload 2")),
		NewUntypedEventWrapper("id-type2-2", "sender-type2-2", time.Now().UTC(), "type2", []byte("another test payload 2")),
	}

	err := pool.ProduceBatch(ctx, wrappers...)

	assert.NoError(t, err)
	producer1.AssertNumberOfCalls(t, "Produce", 3)
	producer2.AssertNumberOfCalls(t, "Produce", 3)
	producer3.AssertNumberOfCalls(t, "Produce", 2)
}

func TestProducersPool_ProductBatch_WithPacks(t *testing.T) {
	ctx := context.Background()
	config := ProducerPoolConfiguration{
		ProducerPoolID: "test-pool",
		SupportedPayloadTypesByProducerID: map[string][]string{
			"producer1": {"type1"},
			"producer2": {"*"},
		},
		SupportedPacksByProducerID: map[string][]string{
			"producer1": {"TestMessagePack"},
		},
	}
	f := message_factory.Singleton
	testPack := message_factory.NewPack(
		"TestMessagePack",
		[]message_factory.UntypedPackEntry{
			{
				MessageType: types.PayloadType("first_payload_type"),
			},
			{
				MessageType: types.PayloadType("second_payload_type"),
			},
			{
				MessageType: types.PayloadType("type1"),
			},
		}...,
	)

	message_factory.RegisterPack(testPack, f)

	producer1 := new(MockProducer)
	producer1.On("ID").Return("producer1")
	producer1.On("Produce", mock.Anything, mock.Anything).Return(nil)

	producer2 := new(MockProducer)
	producer2.On("ID").Return("producer2")
	producer2.On("Produce", mock.Anything, mock.Anything).Return(nil)

	pool, _ := NewProducersPool(ctx, config, producer1, producer2)

	wrappers := []UntypedEventWrapper{
		NewUntypedEventWrapper("id-type1", "sender-type1", time.Now().UTC(), "type1", []byte("test payload")),
		NewUntypedEventWrapper("id-type2", "sender-type2", time.Now().UTC(), "type2", []byte("test payload 2")),
		NewUntypedEventWrapper("id-second_payload_type", "sender-second_payload_type", time.Now().UTC(), "second_payload_type", []byte("another test payload 2")),
	}

	err := pool.ProduceBatch(ctx, wrappers...)

	assert.NoError(t, err)
	producer1.AssertNumberOfCalls(t, "Produce", 2)
	producer2.AssertNumberOfCalls(t, "Produce", 3)
}

func TestProducersPool_ProduceWithUnsupportedType(t *testing.T) {
	ctx := context.Background()
	config := ProducerPoolConfiguration{
		ProducerPoolID: "test-pool",
		SupportedPayloadTypesByProducerID: map[string][]string{
			"producer1": {"type1"},
		},
	}

	producer1 := new(MockProducer)
	producer1.On("ID").Return("producer1")

	pool, _ := NewProducersPool(ctx, config, producer1)

	wrapper := NewUntypedEventWrapper("id-type1", "sender-type1", time.Now().UTC(), "type-unsupported", []byte("test payload"))
	err := pool.Produce(ctx, wrapper)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no producers found for payload type 'type-unsupported' nor for '*'")
}
