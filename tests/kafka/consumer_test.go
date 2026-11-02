package kafka_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	kafkaLib "github.com/spotlibs/go-lib/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConsumerClient is a mock implementation of KafkaConsumerClient for testing.
type mockConsumerClient struct {
	subscribeTopicsFn func(topics []string, rebalanceCb kafka.RebalanceCb) error
	pollFn            func(timeoutMs int) kafka.Event
	commitMessageFn   func(m *kafka.Message) ([]kafka.TopicPartition, error)
	closeFn           func() error
}

func newMockConsumerClient() *mockConsumerClient {
	return &mockConsumerClient{}
}

func (m *mockConsumerClient) SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) error {
	if m.subscribeTopicsFn != nil {
		return m.subscribeTopicsFn(topics, rebalanceCb)
	}
	return nil
}

func (m *mockConsumerClient) Poll(timeoutMs int) kafka.Event {
	if m.pollFn != nil {
		return m.pollFn(timeoutMs)
	}
	return nil
}

func (m *mockConsumerClient) CommitMessage(m2 *kafka.Message) ([]kafka.TopicPartition, error) {
	if m.commitMessageFn != nil {
		return m.commitMessageFn(m2)
	}
	return nil, nil
}

func (m *mockConsumerClient) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func testConsumerConfig() *kafkaLib.ConsumerConfig {
	return &kafkaLib.ConsumerConfig{
		Brokers:     "localhost:9092",
		Username:    "testuser",
		Password:    "testpass",
		GroupID:     "test-group",
		Topics:      []string{"test-topic"},
		PollTimeout: 10 * time.Millisecond,
	}
}

func TestNewConsumerWithClient(t *testing.T) {
	client := newMockConsumerClient()
	cfg := testConsumerConfig()

	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})
	assert.NotNil(t, c)
}

func TestConsumer_Subscribe_Success(t *testing.T) {
	client := newMockConsumerClient()
	var subscribedTopics []string
	client.subscribeTopicsFn = func(topics []string, rebalanceCb kafka.RebalanceCb) error {
		subscribedTopics = topics
		return nil
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	err := c.Subscribe()
	assert.NoError(t, err)
	assert.Equal(t, []string{"test-topic"}, subscribedTopics)
}

func TestConsumer_Subscribe_Error(t *testing.T) {
	client := newMockConsumerClient()
	client.subscribeTopicsFn = func(topics []string, rebalanceCb kafka.RebalanceCb) error {
		return errors.New("subscribe failed")
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	err := c.Subscribe()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscribe failed")
}

func TestConsumer_Poll_Message(t *testing.T) {
	topic := "test-topic"
	client := newMockConsumerClient()
	client.pollFn = func(timeoutMs int) kafka.Event {
		return &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: 0,
				Offset:    kafka.Offset(100),
			},
			Key:       []byte("key1"),
			Value:     []byte(`{"data":"test"}`),
			Timestamp: time.Now(),
			Headers: []kafka.Header{
				{Key: "trace-id", Value: []byte("xyz")},
			},
		}
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	msg, err := c.Poll(1000)
	assert.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "test-topic", msg.Topic)
	assert.Equal(t, int32(0), msg.Partition)
	assert.Equal(t, int64(100), msg.Offset)
	assert.Equal(t, []byte("key1"), msg.Key)
	assert.Equal(t, []byte(`{"data":"test"}`), msg.Value)
	assert.Len(t, msg.Headers, 1)
	assert.Equal(t, "trace-id", msg.Headers[0].Key)
}

func TestConsumer_Poll_NoMessage(t *testing.T) {
	client := newMockConsumerClient()
	client.pollFn = func(timeoutMs int) kafka.Event {
		return nil
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	msg, err := c.Poll(100)
	assert.NoError(t, err)
	assert.Nil(t, msg)
}

func TestConsumer_Poll_NonFatalError(t *testing.T) {
	client := newMockConsumerClient()
	client.pollFn = func(timeoutMs int) kafka.Event {
		return kafka.NewError(kafka.ErrTransport, "connection reset", false)
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	msg, err := c.Poll(100)
	assert.NoError(t, err) // non-fatal errors return nil, nil
	assert.Nil(t, msg)
}

func TestConsumer_Poll_FatalError(t *testing.T) {
	client := newMockConsumerClient()
	client.pollFn = func(timeoutMs int) kafka.Event {
		return kafka.NewError(kafka.ErrFatal, "fatal error", true)
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	msg, err := c.Poll(100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fatal")
	assert.Nil(t, msg)
}

func TestConsumer_Poll_MessageError(t *testing.T) {
	topic := "test-topic"
	client := newMockConsumerClient()
	client.pollFn = func(timeoutMs int) kafka.Event {
		return &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: 0,
				Offset:    kafka.OffsetInvalid,
				Error:     errors.New("partition error"),
			},
		}
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	msg, err := c.Poll(100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "partition error")
	assert.Nil(t, msg)
}

func TestConsumer_Consume_Success(t *testing.T) {
	topic := "test-topic"
	var callCount int32
	client := newMockConsumerClient()
	client.pollFn = func(timeoutMs int) kafka.Event {
		count := atomic.AddInt32(&callCount, 1)
		if count <= 3 {
			return &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: 0,
					Offset:    kafka.Offset(count),
				},
				Key:   []byte("key"),
				Value: []byte("value"),
			}
		}
		// After 3 messages, return nil to simulate no more messages
		time.Sleep(20 * time.Millisecond)
		return nil
	}
	client.commitMessageFn = func(m *kafka.Message) ([]kafka.TopicPartition, error) {
		return nil, nil
	}

	cfg := testConsumerConfig()
	cfg.PollTimeout = 10 * time.Millisecond
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	var processedCount int32
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := c.Consume(ctx, func(ctx context.Context, msg *kafkaLib.Message) error {
		atomic.AddInt32(&processedCount, 1)
		return nil
	})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&processedCount), int32(3))
}

func TestConsumer_Consume_HandlerError(t *testing.T) {
	topic := "test-topic"
	var callCount int32
	client := newMockConsumerClient()
	client.pollFn = func(timeoutMs int) kafka.Event {
		count := atomic.AddInt32(&callCount, 1)
		if count <= 2 {
			return &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: 0,
					Offset:    kafka.Offset(count),
				},
				Key:   []byte("key"),
				Value: []byte("value"),
			}
		}
		time.Sleep(20 * time.Millisecond)
		return nil
	}

	cfg := testConsumerConfig()
	cfg.PollTimeout = 10 * time.Millisecond
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	var processedCount int32
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	// Handler returns error - consumer should continue
	err := c.Consume(ctx, func(ctx context.Context, msg *kafkaLib.Message) error {
		atomic.AddInt32(&processedCount, 1)
		return errors.New("processing failed")
	})

	// Consumer should have continued despite handler errors
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&processedCount), int32(2))
}

func TestConsumer_Consume_ContextCancelled(t *testing.T) {
	client := newMockConsumerClient()
	client.pollFn = func(timeoutMs int) kafka.Event {
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	cfg := testConsumerConfig()
	cfg.PollTimeout = 10 * time.Millisecond
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a brief delay
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := c.Consume(ctx, func(ctx context.Context, msg *kafkaLib.Message) error {
		return nil
	})

	assert.ErrorIs(t, err, context.Canceled)
}

func TestConsumer_CommitMessage(t *testing.T) {
	var committedMsg *kafka.Message
	client := newMockConsumerClient()
	client.commitMessageFn = func(m *kafka.Message) ([]kafka.TopicPartition, error) {
		committedMsg = m
		return []kafka.TopicPartition{m.TopicPartition}, nil
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	msg := &kafkaLib.Message{
		Topic:     "test-topic",
		Partition: 2,
		Offset:    50,
	}

	err := c.CommitMessage(msg)
	assert.NoError(t, err)
	require.NotNil(t, committedMsg)
	assert.Equal(t, int32(2), committedMsg.TopicPartition.Partition)
	// Should commit offset+1 (next offset to read)
	assert.Equal(t, kafka.Offset(51), committedMsg.TopicPartition.Offset)
}

func TestConsumer_CommitMessage_Error(t *testing.T) {
	client := newMockConsumerClient()
	client.commitMessageFn = func(m *kafka.Message) ([]kafka.TopicPartition, error) {
		return nil, errors.New("commit failed")
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	msg := &kafkaLib.Message{
		Topic:     "test-topic",
		Partition: 0,
		Offset:    10,
	}

	err := c.CommitMessage(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit failed")
}

func TestConsumer_Close(t *testing.T) {
	client := newMockConsumerClient()
	closeCalled := false
	client.closeFn = func() error {
		closeCalled = true
		return nil
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	err := c.Close()
	assert.NoError(t, err)
	assert.True(t, closeCalled)
}

func TestConsumer_Close_Error(t *testing.T) {
	client := newMockConsumerClient()
	client.closeFn = func() error {
		return errors.New("close failed")
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	err := c.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "close failed")
}

func TestConsumerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  kafkaLib.ConsumerConfig
		wantErr string
	}{
		{
			name:    "missing brokers",
			config:  kafkaLib.ConsumerConfig{Username: "u", Password: "p", GroupID: "g", Topics: []string{"t"}},
			wantErr: "brokers is required",
		},
		{
			name:    "missing username",
			config:  kafkaLib.ConsumerConfig{Brokers: "b", Password: "p", GroupID: "g", Topics: []string{"t"}},
			wantErr: "username is required",
		},
		{
			name:    "missing password",
			config:  kafkaLib.ConsumerConfig{Brokers: "b", Username: "u", GroupID: "g", Topics: []string{"t"}},
			wantErr: "password is required",
		},
		{
			name:    "missing group_id",
			config:  kafkaLib.ConsumerConfig{Brokers: "b", Username: "u", Password: "p", Topics: []string{"t"}},
			wantErr: "group_id is required",
		},
		{
			name:    "missing topics",
			config:  kafkaLib.ConsumerConfig{Brokers: "b", Username: "u", Password: "p", GroupID: "g"},
			wantErr: "at least one topic is required",
		},
		{
			name:    "valid config",
			config:  kafkaLib.ConsumerConfig{Brokers: "b", Username: "u", Password: "p", GroupID: "g", Topics: []string{"t"}},
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

func TestConsumerConfig_ToConfigMap(t *testing.T) {
	cfg := &kafkaLib.ConsumerConfig{
		Brokers:          "broker1:9092",
		Username:         "myuser",
		Password:         "mypass",
		ClientID:         "my-client",
		GroupID:          "my-group",
		AutoOffsetReset:  "latest",
		SecurityProtocol: "SASL_PLAINTEXT",
		Topics:           []string{"topic1"},
	}

	cm := cfg.ToConfigMap()

	v, _ := cm.Get("bootstrap.servers", "")
	assert.Equal(t, "broker1:9092", v)
	v, _ = cm.Get("group.id", "")
	assert.Equal(t, "my-group", v)
	v, _ = cm.Get("auto.offset.reset", "")
	assert.Equal(t, "latest", v)
	v, _ = cm.Get("client.id", "")
	assert.Equal(t, "my-client", v)
	v, _ = cm.Get("security.protocol", "")
	assert.Equal(t, "SASL_PLAINTEXT", v)
}

func TestConsumer_Subscribe_RebalanceCallback(t *testing.T) {
	client := newMockConsumerClient()
	var rebalanceCb kafka.RebalanceCb
	client.subscribeTopicsFn = func(topics []string, cb kafka.RebalanceCb) error {
		rebalanceCb = cb
		return nil
	}

	cfg := testConsumerConfig()
	c := kafkaLib.NewConsumerWithClient(client, cfg, &kafkaLib.NopLogger{})

	err := c.Subscribe()
	assert.NoError(t, err)
	require.NotNil(t, rebalanceCb)

	// Test assigned partitions
	topic := "test-topic"
	assignedEvent := kafka.AssignedPartitions{
		Partitions: []kafka.TopicPartition{
			{Topic: &topic, Partition: 0, Offset: 0},
			{Topic: &topic, Partition: 1, Offset: 0},
		},
	}
	err = rebalanceCb(nil, assignedEvent)
	assert.NoError(t, err)

	// Test revoked partitions
	revokedEvent := kafka.RevokedPartitions{
		Partitions: []kafka.TopicPartition{
			{Topic: &topic, Partition: 0, Offset: 100},
		},
	}
	err = rebalanceCb(nil, revokedEvent)
	assert.NoError(t, err)
}
