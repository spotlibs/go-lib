package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Consumer defines the interface for consuming messages from Kafka.
// This interface enables mocking in unit tests.
type Consumer interface {
	// Subscribe subscribes the consumer to the configured topics.
	Subscribe() error
	// Poll polls for a single message. Returns nil if timeout is reached without a message.
	Poll(timeoutMs int) (*Message, error)
	// Consume starts consuming messages and invokes the handler for each message.
	// This is a blocking call that runs until the context is cancelled.
	Consume(ctx context.Context, handler MessageHandler) error
	// CommitMessage commits the offset of the given message.
	CommitMessage(msg *Message) error
	// Close gracefully shuts down the consumer.
	Close() error
}

// MessageHandler is a function that processes a consumed Kafka message.
// Return nil to acknowledge the message, or return an error to signal failure.
type MessageHandler func(ctx context.Context, msg *Message) error

// Message represents a consumed Kafka message.
type Message struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   []Header
	Timestamp time.Time
}

// KafkaConsumerClient is the interface that wraps the confluent kafka consumer methods we use.
// This enables injecting a mock kafka consumer for testing.
type KafkaConsumerClient interface {
	SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) error
	Poll(timeoutMs int) kafka.Event
	CommitMessage(m *kafka.Message) ([]kafka.TopicPartition, error)
	Close() error
}

// consumer is the concrete implementation of Consumer.
type consumer struct {
	client KafkaConsumerClient
	logger Logger
	config *ConsumerConfig
}

// NewConsumer creates a new Kafka consumer instance.
func NewConsumer(config *ConsumerConfig, logger Logger) (Consumer, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = &defaultLogger{}
	}

	kafkaConsumer, err := kafka.NewConsumer(config.ToConfigMap())
	if err != nil {
		return nil, fmt.Errorf("kafka: failed to create consumer: %w", err)
	}

	c := &consumer{
		client: kafkaConsumer,
		logger: logger,
		config: config,
	}

	return c, nil
}

// NewConsumerWithClient creates a new Kafka consumer with an injected client.
// This is primarily used for testing.
func NewConsumerWithClient(client KafkaConsumerClient, config *ConsumerConfig, logger Logger) Consumer {
	if logger == nil {
		logger = &defaultLogger{}
	}
	return &consumer{
		client: client,
		logger: logger,
		config: config,
	}
}

// Subscribe subscribes the consumer to the configured topics.
func (c *consumer) Subscribe() error {
	rebalanceCb := func(consumer *kafka.Consumer, event kafka.AssignedPartitions) error {
		c.logger.OnRebalance(event.Partitions, true)
		return nil
	}
	_ = rebalanceCb // suppress unused warning

	err := c.client.SubscribeTopics(c.config.Topics, func(consumer *kafka.Consumer, event kafka.Event) error {
		switch e := event.(type) {
		case kafka.AssignedPartitions:
			c.logger.OnRebalance(e.Partitions, true)
		case kafka.RevokedPartitions:
			c.logger.OnRebalance(e.Partitions, false)
		case kafka.Error:
			c.logger.OnError("consumer", e)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("kafka: failed to subscribe to topics %v: %w", c.config.Topics, err)
	}
	return nil
}

// Poll polls for a single message.
func (c *consumer) Poll(timeoutMs int) (*Message, error) {
	ev := c.client.Poll(timeoutMs)
	if ev == nil {
		return nil, nil
	}

	switch e := ev.(type) {
	case *kafka.Message:
		if e.TopicPartition.Error != nil {
			return nil, fmt.Errorf("kafka: message error: %w", e.TopicPartition.Error)
		}
		msg := c.toMessage(e)
		c.logger.OnConsume(msg.Topic, msg.Partition, msg.Offset, msg.Key)
		return msg, nil
	case kafka.Error:
		c.logger.OnError("consumer", e)
		// Non-fatal errors (e.g., broker disconnects) should not stop the consumer
		if e.IsFatal() {
			return nil, fmt.Errorf("kafka: fatal consumer error: %w", e)
		}
		return nil, nil
	default:
		// Other events (e.g., OffsetsCommitted, Stats) are logged but not returned
		return nil, nil
	}
}

// Consume starts consuming messages and invokes the handler for each message.
func (c *consumer) Consume(ctx context.Context, handler MessageHandler) error {
	pollTimeout := c.config.PollTimeout
	if pollTimeout == 0 {
		pollTimeout = 100 * time.Millisecond
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.Poll(int(pollTimeout.Milliseconds()))
			if err != nil {
				return err
			}
			if msg == nil {
				continue
			}

			if err := handler(ctx, msg); err != nil {
				c.logger.OnConsumeError(msg.Topic, msg.Key, err)
				// Continue consuming even on handler errors.
				// The caller can decide to cancel via context if they want to stop.
				continue
			}

			// Auto-commit after successful processing
			if err := c.CommitMessage(msg); err != nil {
				c.logger.OnCommitError(msg.Topic, msg.Partition, msg.Offset, err)
			} else {
				c.logger.OnCommitSuccess(msg.Topic, msg.Partition, msg.Offset)
			}
		}
	}
}

// CommitMessage commits the offset of the given message.
func (c *consumer) CommitMessage(msg *Message) error {
	topic := msg.Topic
	_, err := c.client.CommitMessage(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: msg.Partition,
			Offset:    kafka.Offset(msg.Offset + 1),
		},
	})
	if err != nil {
		return fmt.Errorf("kafka: failed to commit offset for topic %s partition %d offset %d: %w",
			msg.Topic, msg.Partition, msg.Offset, err)
	}
	return nil
}

// Close gracefully shuts down the consumer.
func (c *consumer) Close() error {
	err := c.client.Close()
	if err != nil {
		return fmt.Errorf("kafka: failed to close consumer: %w", err)
	}
	return nil
}

// toMessage converts a confluent kafka message to our Message type.
func (c *consumer) toMessage(m *kafka.Message) *Message {
	headers := make([]Header, 0, len(m.Headers))
	for _, h := range m.Headers {
		headers = append(headers, Header{
			Key:   h.Key,
			Value: h.Value,
		})
	}

	topic := ""
	if m.TopicPartition.Topic != nil {
		topic = *m.TopicPartition.Topic
	}

	return &Message{
		Topic:     topic,
		Partition: m.TopicPartition.Partition,
		Offset:    int64(m.TopicPartition.Offset),
		Key:       m.Key,
		Value:     m.Value,
		Headers:   headers,
		Timestamp: m.Timestamp,
	}
}
