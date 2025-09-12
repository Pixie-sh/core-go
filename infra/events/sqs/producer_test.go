package sqs

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pixie-sh/core-go/infra/events"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pixie-sh/core-go/infra/message_wrapper"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
)

// MockClient is a mock implementation of the SQS client
type MockClient struct {
	mock.Mock
}

func (m *MockClient) SendMessage(ctx context.Context, input *sqs.SendMessageInput, opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

func (m *MockClient) SendMessageBatch(ctx context.Context, input *sqs.SendMessageBatchInput, opts ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*sqs.SendMessageBatchOutput), args.Error(1)
}

func TestNewProducer(t *testing.T) {
	tests := []struct {
		name   string
		config ProducerConfiguration
		want   *Producer
	}{
		{
			name: "with custom CheckSize function",
			config: ProducerConfiguration{
				ProducerID: "test-producer",
				QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				IsFIFO:     false,
				CheckSize: func(b []byte) error {
					if len(b) > 100 {
						return fmt.Errorf("message too large")
					}
					return nil
				},
			},
		},
		{
			name: "with nil CheckSize function - should use default",
			config: ProducerConfiguration{
				ProducerID: "test-producer",
				QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				IsFIFO:     false,
				CheckSize:  nil,
			},
		},
		{
			name: "FIFO queue configuration",
			config: ProducerConfiguration{
				ProducerID: "test-producer-fifo",
				QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue.fifo",
				IsFIFO:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{}
			producer, err := NewProducer(context.Background(), mockClient, tt.config)

			require.NoError(t, err)
			assert.NotNil(t, producer)
			assert.Equal(t, tt.config.ProducerID, producer.cfg.ProducerID)
			assert.Equal(t, tt.config.QueueURL, producer.cfg.QueueURL)
			assert.Equal(t, tt.config.IsFIFO, producer.cfg.IsFIFO)
			assert.NotNil(t, producer.cfg.CheckSize) // Should always be set
		})
	}
}

func TestProducer_DefaultCheckSize(t *testing.T) {
	tests := []struct {
		name        string
		messageSize int
		expectError bool
	}{
		{
			name:        "message within limit",
			messageSize: 100,
			expectError: false,
		},
		{
			name:        "message at limit",
			messageSize: 262143,
			expectError: false,
		},
		{
			name:        "message exceeds limit",
			messageSize: 262144,
			expectError: true,
		},
		{
			name:        "large message",
			messageSize: 500000,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{}
			config := ProducerConfiguration{
				ProducerID: "test-producer",
				QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				IsFIFO:     false,
			}

			producer, err := NewProducer(context.Background(), mockClient, config)
			require.NoError(t, err)

			// Create a message of the specified size
			message := make([]byte, tt.messageSize)
			for i := range message {
				message[i] = 'a'
			}

			err = producer.cfg.CheckSize(message)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "exceeds maximum allowed size")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProducer_ID(t *testing.T) {
	mockClient := &MockClient{}
	config := ProducerConfiguration{
		ProducerID: "test-producer-123",
		QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
	}

	producer, err := NewProducer(context.Background(), mockClient, config)
	require.NoError(t, err)

	assert.Equal(t, "test-producer-123", producer.ID())
}

func TestProducer_createMessageAttributes(t *testing.T) {
	mockClient := &MockClient{}
	config := ProducerConfiguration{
		ProducerID: "test-producer",
		QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
	}

	producer, err := NewProducer(context.Background(), mockClient, config)
	require.NoError(t, err)

	os.Setenv("SCOPE", "testscope")
	ctx := context.Background()
	ctx = pixiecontext.SetCtxTraceID(ctx, "testtraceid")
	attrs := producer.createMessageAttributes(ctx)

	// Check that basic attributes are present
	assert.NotNil(t, attrs)
	assert.Contains(t, attrs, "SCOPE")    // env.Scope
	assert.Contains(t, attrs, "trace_id") // logger.TraceID

	// Check attribute structure
	for _, attr := range attrs {
		assert.NotNil(t, attr.DataType)
		assert.Equal(t, "String", *attr.DataType)
		assert.NotNil(t, attr.StringValue)
	}
}

func TestProducer_appendPayloadType(t *testing.T) {
	mockClient := &MockClient{}
	config := ProducerConfiguration{
		ProducerID: "test-producer",
		QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
	}

	producer, err := NewProducer(context.Background(), mockClient, config)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name        string
		payloadType string
		eventID     string
		attributes  map[string]types.MessageAttributeValue
	}{
		{
			name:        "with existing attributes",
			payloadType: "test.event",
			eventID:     "event-123",
			attributes: map[string]types.MessageAttributeValue{
				"custom": {
					DataType:    aws.String("String"),
					StringValue: aws.String("value"),
				},
			},
		},
		{
			name:        "with nil attributes",
			payloadType: "another.event",
			eventID:     "event-456",
			attributes:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := producer.appendPayloadType(ctx, tt.payloadType, tt.eventID, tt.attributes)

			assert.NotNil(t, result)
			assert.Contains(t, result, "x-payload-type")
			assert.Contains(t, result, "x-event-id")

			// Check payload type attribute
			payloadTypeAttr := result["x-payload-type"]
			assert.Equal(t, "String", *payloadTypeAttr.DataType)
			assert.Equal(t, tt.payloadType, *payloadTypeAttr.StringValue)

			// Check event ID attribute
			eventIDAttr := result["x-event-id"]
			assert.Equal(t, "String", *eventIDAttr.DataType)
			assert.Equal(t, tt.eventID, *eventIDAttr.StringValue)

			// If original attributes were provided, they should still be there
			if tt.attributes != nil {
				for key, value := range tt.attributes {
					assert.Contains(t, result, key)
					assert.Equal(t, value, result[key])
				}
			}
		})
	}
}

func TestProducer_dedupID(t *testing.T) {
	mockClient := &MockClient{}
	config := ProducerConfiguration{
		ProducerID: "test-producer",
		QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
	}

	producer, err := NewProducer(context.Background(), mockClient, config)
	require.NoError(t, err)

	timestamp := time.Now()
	wrapper := &message_wrapper.UntypedMessage{
		ID:        "test-message-123",
		Timestamp: timestamp,
	}

	ctx := context.Background()
	result := producer.dedupID(ctx, wrapper)

	assert.NotNil(t, result)
	expected := fmt.Sprintf("%s:%d", wrapper.ID, timestamp.UnixMilli())
	assert.Equal(t, expected, *result)
}

func TestProducer_CustomCheckSizeFunction(t *testing.T) {
	mockClient := &MockClient{}
	customCheckCalled := false
	config := ProducerConfiguration{
		ProducerID: "test-producer",
		QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		CheckSize: func(b []byte) error {
			customCheckCalled = true
			if len(b) > 50 {
				return fmt.Errorf("custom size limit exceeded")
			}
			return nil
		},
	}

	producer, err := NewProducer(context.Background(), mockClient, config)
	require.NoError(t, err)

	// Test that custom function is used
	smallMessage := make([]byte, 30)
	err = producer.cfg.CheckSize(smallMessage)
	assert.NoError(t, err)
	assert.True(t, customCheckCalled)

	// Reset flag
	customCheckCalled = false

	// Test that custom function rejects larger messages
	largeMessage := make([]byte, 100)
	err = producer.cfg.CheckSize(largeMessage)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "custom size limit exceeded")
	assert.True(t, customCheckCalled)
}

func TestProducer_ConfigurationValidation(t *testing.T) {
	mockClient := &MockClient{}
	tests := []struct {
		name   string
		config ProducerConfiguration
	}{
		{
			name: "empty producer ID",
			config: ProducerConfiguration{
				ProducerID: "",
				QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			},
		},
		{
			name: "empty queue URL",
			config: ProducerConfiguration{
				ProducerID: "test-producer",
				QueueURL:   "",
			},
		},
		{
			name: "valid minimal configuration",
			config: ProducerConfiguration{
				ProducerID: "test-producer",
				QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer, err := NewProducer(context.Background(), mockClient, tt.config)

			// The constructor should not fail for any configuration
			// as validation is typically done at usage time
			assert.NoError(t, err)
			assert.NotNil(t, producer)
			assert.Equal(t, tt.config.ProducerID, producer.cfg.ProducerID)
			assert.Equal(t, tt.config.QueueURL, producer.cfg.QueueURL)
		})
	}
}

func TestProducer_MessageSizeErrorFormat(t *testing.T) {
	mockClient := &MockClient{}
	config := ProducerConfiguration{
		ProducerID: "test-producer",
		QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
	}

	producer, err := NewProducer(context.Background(), mockClient, config)
	require.NoError(t, err)

	// Create a message that exceeds the limit
	largeMessage := make([]byte, 300000)
	err = producer.cfg.CheckSize(largeMessage)

	require.Error(t, err)

	// Check error message format
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "message size")
	assert.Contains(t, errorMsg, "exceeds maximum allowed size")
	assert.Contains(t, errorMsg, "300000") // actual size
	assert.Contains(t, errorMsg, "262143") // max size
}

// Benchmark tests
func BenchmarkProducer_CheckSize(b *testing.B) {
	mockClient := &MockClient{}
	config := ProducerConfiguration{
		ProducerID: "test-producer",
		QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
	}

	producer, err := NewProducer(context.Background(), mockClient, config)
	require.NoError(b, err)

	message := make([]byte, 100000) // 100KB message

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = producer.cfg.CheckSize(message)
	}
}

func BenchmarkProducer_CreateMessageAttributes(b *testing.B) {
	mockClient := &MockClient{}
	config := ProducerConfiguration{
		ProducerID: "test-producer",
		QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
	}

	producer, err := NewProducer(context.Background(), mockClient, config)
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = producer.createMessageAttributes(ctx)
	}
}

func TestProducer_CheckSizeIntegration(t *testing.T) {
	mockClient := &MockClient{}

	// Track CheckSize calls
	checkSizeCalls := 0
	checkSizePayloads := [][]byte{}

	config := ProducerConfiguration{
		ProducerID: "test-producer",
		QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		IsFIFO:     false,
		CheckSize: func(b []byte) error {
			checkSizeCalls++
			checkSizePayloads = append(checkSizePayloads, b)

			// Simulate size limit of 1000 bytes for testing
			if len(b) > 1000 {
				return fmt.Errorf("message size %d bytes exceeds test limit of 1000 bytes", len(b))
			}
			return nil
		},
	}

	producer, err := NewProducer(context.Background(), mockClient, config)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Produce method calls CheckSize", func(t *testing.T) {
		// Reset counters
		checkSizeCalls = 0
		checkSizePayloads = [][]byte{}

		// Mock successful SendMessage
		mockClient.On("SendMessage", mock.Anything, mock.Anything).Return(&sqs.SendMessageOutput{}, nil).Once()

		wrapper := events.UntypedEventWrapper{
			UntypedMessage: message_wrapper.UntypedMessage{
				ID:        "test-event-1",
				Timestamp: time.Now(),
				// This will be serialized and checked
			},
		}

		err := producer.Produce(ctx, wrapper)

		assert.NoError(t, err)
		assert.Equal(t, 1, checkSizeCalls, "CheckSize should be called exactly once")
		assert.Len(t, checkSizePayloads, 1, "Should capture one payload")

		// Verify the payload is not empty (serialization worked)
		assert.NotEmpty(t, checkSizePayloads[0], "Payload should not be empty")

		mockClient.AssertExpectations(t)
	})

	t.Run("Produce method fails when CheckSize returns error", func(t *testing.T) {
		// Reset counters
		checkSizeCalls = 0
		checkSizePayloads = [][]byte{}

		// Create a wrapper that will result in a large payload
		largeData := strings.Repeat("x", 2000) // This will exceed our 1000 byte test limit
		wrapper := events.UntypedEventWrapper{
			UntypedMessage: message_wrapper.UntypedMessage{
				ID:        "test-event-large",
				Timestamp: time.Now(),
				// Add large data to make payload exceed limit
				Payload: largeData,
			},
		}

		err := producer.Produce(ctx, wrapper)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds test limit")
		assert.Equal(t, 1, checkSizeCalls, "CheckSize should be called exactly once")

		// SendMessage should not be called when CheckSize fails
		mockClient.AssertNotCalled(t, "SendMessage")
	})

	t.Run("ProduceBatch method calls CheckSize for each message", func(t *testing.T) {
		// Reset counters
		checkSizeCalls = 0
		checkSizePayloads = [][]byte{}

		// Mock successful SendMessageBatch
		mockClient.On("SendMessageBatch", mock.Anything, mock.Anything).Return(&sqs.SendMessageBatchOutput{}, nil).Once()

		wrappers := []events.UntypedEventWrapper{
			{
				UntypedMessage: message_wrapper.UntypedMessage{
					ID:        "batch-event-1",
					Timestamp: time.Now(),
				},
			},
			{
				UntypedMessage: message_wrapper.UntypedMessage{
					ID:        "batch-event-2",
					Timestamp: time.Now(),
				},
			},
			{
				UntypedMessage: message_wrapper.UntypedMessage{
					ID:        "batch-event-3",
					Timestamp: time.Now(),
				},
			},
		}

		err := producer.ProduceBatch(ctx, wrappers...)

		assert.NoError(t, err)
		assert.Equal(t, 3, checkSizeCalls, "CheckSize should be called for each message")
		assert.Len(t, checkSizePayloads, 3, "Should capture all payloads")

		// Verify all payloads are not empty
		for i, payload := range checkSizePayloads {
			assert.NotEmpty(t, payload, fmt.Sprintf("Payload %d should not be empty", i))
		}

		mockClient.AssertExpectations(t)
	})

	t.Run("ProduceBatch method fails on first oversized message", func(t *testing.T) {
		// Reset counters
		checkSizeCalls = 0
		checkSizePayloads = [][]byte{}

		largeData := strings.Repeat("y", 2000) // Exceeds 1000 byte limit
		wrappers := []events.UntypedEventWrapper{
			{
				UntypedMessage: message_wrapper.UntypedMessage{
					ID:        "batch-event-good",
					Timestamp: time.Now(),
				},
			},
			{
				UntypedMessage: message_wrapper.UntypedMessage{
					ID:        "batch-event-bad",
					Timestamp: time.Now(),
					Payload:   largeData, // This will cause CheckSize to fail
				},
			},
			{
				UntypedMessage: message_wrapper.UntypedMessage{
					ID:        "batch-event-never-reached",
					Timestamp: time.Now(),
				},
			},
		}

		err := producer.ProduceBatch(ctx, wrappers...)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds test limit")
		assert.Equal(t, 2, checkSizeCalls, "CheckSize should be called twice (stops at first error)")

		// SendMessageBatch should not be called when CheckSize fails
		mockClient.AssertNotCalled(t, "SendMessageBatch")
	})

	t.Run("ProduceWithQueue method calls CheckSize", func(t *testing.T) {
		// Reset counters
		checkSizeCalls = 0
		checkSizePayloads = [][]byte{}

		// Mock successful SendMessage
		mockClient.On("SendMessage", mock.Anything, mock.Anything).Return(&sqs.SendMessageOutput{}, nil).Once()

		wrapper := message_wrapper.UntypedMessage{
			ID:        "queue-event-1",
			Timestamp: time.Now(),
		}

		err := producer.ProduceWithQueue(ctx, wrapper, "https://sqs.us-east-1.amazonaws.com/123456789012/custom-queue", false)

		assert.NoError(t, err)
		assert.Equal(t, 1, checkSizeCalls, "CheckSize should be called exactly once")
		assert.Len(t, checkSizePayloads, 1, "Should capture one payload")
		assert.NotEmpty(t, checkSizePayloads[0], "Payload should not be empty")

		mockClient.AssertExpectations(t)
	})

	t.Run("ProduceWithQueue method fails when CheckSize returns error", func(t *testing.T) {
		// Reset counters
		checkSizeCalls = 0
		checkSizePayloads = [][]byte{}

		largeData := strings.Repeat("z", 2000) // Exceeds 1000 byte limit
		wrapper := message_wrapper.UntypedMessage{
			ID:        "queue-event-large",
			Timestamp: time.Now(),
			Payload:   largeData,
		}

		err := producer.ProduceWithQueue(ctx, wrapper, "https://sqs.us-east-1.amazonaws.com/123456789012/custom-queue", false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds test limit")
		assert.Equal(t, 1, checkSizeCalls, "CheckSize should be called exactly once")

		// SendMessage should not be called when CheckSize fails
		mockClient.AssertNotCalled(t, "SendMessage")
	})

	t.Run("CheckSize receives actual serialized payload", func(t *testing.T) {
		// Reset counters
		checkSizeCalls = 0
		checkSizePayloads = [][]byte{}

		// Mock successful SendMessage
		mockClient.On("SendMessage", mock.Anything, mock.Anything).Return(&sqs.SendMessageOutput{}, nil).Once()

		testData := "specific-test-data-12345"
		wrapper := events.UntypedEventWrapper{
			UntypedMessage: message_wrapper.UntypedMessage{
				ID:        "payload-test",
				Timestamp: time.Now(),
				Payload:   testData,
			},
		}

		err := producer.Produce(ctx, wrapper)

		assert.NoError(t, err)
		assert.Equal(t, 1, checkSizeCalls, "CheckSize should be called exactly once")

		// Verify the payload contains our test data (after serialization)
		payloadStr := string(checkSizePayloads[0])
		assert.Contains(t, payloadStr, testData, "Serialized payload should contain our test data")

		mockClient.AssertExpectations(t)
	})

	t.Run("CheckSize is called after serialization but before SQS client call", func(t *testing.T) {
		// Reset counters
		checkSizeCalls = 0
		checkSizePayloads = [][]byte{}

		// Create a custom CheckSize that will fail
		failingConfig := ProducerConfiguration{
			ProducerID: "test-producer",
			QueueURL:   "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			IsFIFO:     false,
			CheckSize: func(b []byte) error {
				checkSizeCalls++
				checkSizePayloads = append(checkSizePayloads, b)
				return fmt.Errorf("intentional test failure")
			},
		}

		failingProducer, err := NewProducer(context.Background(), mockClient, failingConfig)
		require.NoError(t, err)

		wrapper := events.UntypedEventWrapper{
			UntypedMessage: message_wrapper.UntypedMessage{
				ID:        "test-ordering",
				Timestamp: time.Now(),
			},
		}

		err = failingProducer.Produce(ctx, wrapper)

		// Should fail due to CheckSize
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "intentional test failure")
		assert.Equal(t, 1, checkSizeCalls, "CheckSize should be called")
		assert.NotEmpty(t, checkSizePayloads[0], "Should have captured serialized payload")

		// Most importantly: SendMessage should NOT be called because CheckSize failed
		mockClient.AssertNotCalled(t, "SendMessage")
	})
}
