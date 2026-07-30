package kafka

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	kafkaLib "github.com/spotlibs/go-lib/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProducerClient is a mock implementation of KafkaProducerClient for testing.
type mockProducerClient struct {
	produceFn func(msg *kafka.Message, deliveryChan chan kafka.Event) error
	flushFn   func(timeoutMs int) int
	closeFn   func()
	eventsCh  chan kafka.Event
	closeOnce sync.Once
}

func newMockProducerClient() *mockProducerClient {
	return &mockProducerClient{
		eventsCh: make(chan kafka.Event, 100),
	}
}

func (m *mockProducerClient) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	if m.produceFn != nil {
		return m.produceFn(msg, deliveryChan)
	}
	// Default: simulate successful delivery
	if deliveryChan != nil {
		go func() {
			deliveryChan <- &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     msg.TopicPartition.Topic,
					Partition: 0,
					Offset:    kafka.Offset(42),
				},
				Key:   msg.Key,
				Value: msg.Value,
			}
		}()
	}
	return nil
}

func (m *mockProducerClient) Flush(timeoutMs int) int {
	if m.flushFn != nil {
		return m.flushFn(timeoutMs)
	}
	return 0
}

func (m *mockProducerClient) Close() {
	if m.closeFn != nil {
		m.closeFn()
	}
	m.closeOnce.Do(func() {
		close(m.eventsCh)
	})
}

func (m *mockProducerClient) Events() chan kafka.Event {
	return m.eventsCh
}

func testProducerConfig() *kafkaLib.ProducerConfig {
	return &kafkaLib.ProducerConfig{
		Brokers:  "localhost:9092",
		Username: "testuser",
		Password: "testpass",
	}
}

func TestNewProducerWithClient(t *testing.T) {
	client := newMockProducerClient()
	cfg := testProducerConfig()

	p := kafkaLib.NewProducerWithClient(client, cfg, &kafkaLib.NopLogger{})
	assert.NotNil(t, p)
}

func TestProducer_Publish_Success(t *testing.T) {
	client := newMockProducerClient()
	cfg := testProducerConfig()
	p := kafkaLib.NewProducerWithClient(client, cfg, &kafkaLib.NopLogger{})

	ctx := context.Background()
	err := p.Publish(ctx, "test-topic", []byte("key1"), []byte(`{"msg":"hello"}`), nil)

	assert.NoError(t, err)
}

func TestProducer_Publish_WithHeaders(t *testing.T) {
	var capturedMsg *kafka.Message
	client := newMockProducerClient()
	client.produceFn = func(msg *kafka.Message, deliveryChan chan kafka.Event) error {
		capturedMsg = msg
		// simulate success
		if deliveryChan != nil {
			go func() {
				deliveryChan <- &kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     msg.TopicPartition.Topic,
						Partition: 0,
						Offset:    kafka.Offset(1),
					},
				}
			}()
		}
		return nil
	}

	cfg := testProducerConfig()
	p := kafkaLib.NewProducerWithClient(client, cfg, &kafkaLib.NopLogger{})

	headers := []kafkaLib.Header{
		{Key: "trace-id", Value: []byte("abc123")},
		{Key: "source", Value: []byte("unit-test")},
	}

	ctx := context.Background()
	err := p.Publish(ctx, "test-topic", []byte("key1"), []byte("value1"), headers)

	assert.NoError(t, err)
	require.NotNil(t, capturedMsg)
	assert.Len(t, capturedMsg.Headers, 2)
	assert.Equal(t, "trace-id", capturedMsg.Headers[0].Key)
	assert.Equal(t, []byte("abc123"), capturedMsg.Headers[0].Value)
}

func TestProducer_Publish_ProduceError(t *testing.T) {
	client := newMockProducerClient()
	client.produceFn = func(msg *kafka.Message, deliveryChan chan kafka.Event) error {
		return errors.New("queue full")
	}

	cfg := testProducerConfig()
	p := kafkaLib.NewProducerWithClient(client, cfg, &kafkaLib.NopLogger{})

	ctx := context.Background()
	err := p.Publish(ctx, "test-topic", []byte("key1"), []byte("value1"), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue full")
}

func TestProducer_Publish_DeliveryError(t *testing.T) {
	client := newMockProducerClient()
	client.produceFn = func(msg *kafka.Message, deliveryChan chan kafka.Event) error {
		if deliveryChan != nil {
			go func() {
				deliveryChan <- &kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     msg.TopicPartition.Topic,
						Partition: 0,
						Offset:    kafka.OffsetInvalid,
						Error:     errors.New("broker unavailable"),
					},
				}
			}()
		}
		return nil
	}

	cfg := testProducerConfig()
	p := kafkaLib.NewProducerWithClient(client, cfg, &kafkaLib.NopLogger{})

	ctx := context.Background()
	err := p.Publish(ctx, "test-topic", []byte("key1"), []byte("value1"), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delivery failed")
	assert.Contains(t, err.Error(), "broker unavailable")
}

func TestProducer_Publish_ContextCancelled(t *testing.T) {
	client := newMockProducerClient()
	client.produceFn = func(msg *kafka.Message, deliveryChan chan kafka.Event) error {
		// Don't send anything to deliveryChan - simulating a slow broker
		return nil
	}

	cfg := testProducerConfig()
	p := kafkaLib.NewProducerWithClient(client, cfg, &kafkaLib.NopLogger{})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := p.Publish(ctx, "test-topic", []byte("key1"), []byte("value1"), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}

func TestProducer_PublishWithPartition(t *testing.T) {
	var capturedMsg *kafka.Message
	client := newMockProducerClient()
	client.produceFn = func(msg *kafka.Message, deliveryChan chan kafka.Event) error {
		capturedMsg = msg
		if deliveryChan != nil {
			go func() {
				deliveryChan <- &kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     msg.TopicPartition.Topic,
						Partition: msg.TopicPartition.Partition,
						Offset:    kafka.Offset(10),
					},
				}
			}()
		}
		return nil
	}

	cfg := testProducerConfig()
	p := kafkaLib.NewProducerWithClient(client, cfg, &kafkaLib.NopLogger{})

	ctx := context.Background()
	err := p.PublishWithPartition(ctx, "test-topic", 3, []byte("key1"), []byte("value1"), nil)

	assert.NoError(t, err)
	require.NotNil(t, capturedMsg)
	assert.Equal(t, int32(3), capturedMsg.TopicPartition.Partition)
}

func TestProducer_Flush(t *testing.T) {
	client := newMockProducerClient()
	flushCalled := false
	client.flushFn = func(timeoutMs int) int {
		flushCalled = true
		assert.Equal(t, 5000, timeoutMs)
		return 0
	}

	cfg := testProducerConfig()
	p := kafkaLib.NewProducerWithClient(client, cfg, &kafkaLib.NopLogger{})

	remaining := p.Flush(5000)
	assert.Equal(t, 0, remaining)
	assert.True(t, flushCalled)
}

func TestProducer_Close(t *testing.T) {
	client := newMockProducerClient()
	closeCalled := false
	client.closeFn = func() {
		closeCalled = true
	}
	client.flushFn = func(timeoutMs int) int {
		return 0
	}

	cfg := testProducerConfig()
	p := kafkaLib.NewProducerWithClient(client, cfg, &kafkaLib.NopLogger{})

	p.Close()
	assert.True(t, closeCalled)
}

func TestProducerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  kafkaLib.ProducerConfig
		wantErr string
	}{
		{
			name:    "missing brokers",
			config:  kafkaLib.ProducerConfig{Username: "u", Password: "p"},
			wantErr: "brokers is required",
		},
		{
			name:    "missing username",
			config:  kafkaLib.ProducerConfig{Brokers: "b", Password: "p"},
			wantErr: "username is required",
		},
		{
			name:    "missing password",
			config:  kafkaLib.ProducerConfig{Brokers: "b", Username: "u"},
			wantErr: "password is required",
		},
		{
			name:    "valid config",
			config:  kafkaLib.ProducerConfig{Brokers: "b", Username: "u", Password: "p"},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestProducerConfig_ToConfigMap(t *testing.T) {
	cfg := &kafkaLib.ProducerConfig{
		Brokers:          "broker1:9092,broker2:9092",
		Username:         "myuser",
		Password:         "mypass",
		SecurityProtocol: "SASL_PLAINTEXT",
		SASLMechanism:    "SCRAM-SHA-256",
		CompressionCodec: "snappy",
		MessageTimeoutMs: 10000,
		SocketTimeoutMs:  5000,
		AdditionalConfig: map[string]string{
			"acks": "all",
		},
	}

	cm := cfg.ToConfigMap()

	v, _ := cm.Get("bootstrap.servers", "")
	assert.Equal(t, "broker1:9092,broker2:9092", v)
	v, _ = cm.Get("security.protocol", "")
	assert.Equal(t, "SASL_PLAINTEXT", v)
	v, _ = cm.Get("sasl.mechanism", "")
	assert.Equal(t, "SCRAM-SHA-256", v)
	v, _ = cm.Get("compression.codec", "")
	assert.Equal(t, "snappy", v)
	v, _ = cm.Get("acks", "")
	assert.Equal(t, "all", v)
}
