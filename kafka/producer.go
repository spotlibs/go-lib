package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Producer defines the interface for publishing messages to Kafka.
// This interface enables mocking in unit tests.
type Producer interface {
	// Publish sends a message to the specified topic.
	Publish(ctx context.Context, topic string, key []byte, value []byte, headers []Header) error
	// PublishWithPartition sends a message to a specific partition of the topic.
	PublishWithPartition(ctx context.Context, topic string, partition int32, key []byte, value []byte, headers []Header) error
	// Flush waits for all outstanding produce requests to complete.
	Flush(timeoutMs int) int
	// Close gracefully shuts down the producer.
	Close()
}

// Header represents a Kafka message header.
type Header struct {
	Key   string
	Value []byte
}

// KafkaProducerClient is the interface that wraps the confluent kafka producer methods we use.
// This enables injecting a mock kafka producer for testing.
type KafkaProducerClient interface {
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
	Flush(timeoutMs int) int
	Close()
	Events() chan kafka.Event
}

// producer is the concrete implementation of Producer.
type producer struct {
	client KafkaProducerClient
	logger Logger
	config *ProducerConfig
}

// NewProducer creates a new Kafka producer instance.
func NewProducer(config *ProducerConfig, logger Logger) (Producer, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = &defaultLogger{}
	}

	kafkaProducer, err := kafka.NewProducer(config.ToConfigMap())
	if err != nil {
		return nil, fmt.Errorf("kafka: failed to create producer: %w", err)
	}

	p := &producer{
		client: kafkaProducer,
		logger: logger,
		config: config,
	}

	// Start delivery report handler goroutine
	go p.handleEvents()

	return p, nil
}

// NewProducerWithClient creates a new Kafka producer with an injected client.
// This is primarily used for testing.
func NewProducerWithClient(client KafkaProducerClient, config *ProducerConfig, logger Logger) Producer {
	if logger == nil {
		logger = &defaultLogger{}
	}
	p := &producer{
		client: client,
		logger: logger,
		config: config,
	}
	go p.handleEvents()
	return p
}

// Publish sends a message to the specified topic.
func (p *producer) Publish(ctx context.Context, topic string, key []byte, value []byte, headers []Header) error {
	return p.publish(ctx, topic, kafka.PartitionAny, key, value, headers)
}

// PublishWithPartition sends a message to a specific partition of the topic.
func (p *producer) PublishWithPartition(ctx context.Context, topic string, partition int32, key []byte, value []byte, headers []Header) error {
	return p.publish(ctx, topic, partition, key, value, headers)
}

func (p *producer) publish(ctx context.Context, topic string, partition int32, key []byte, value []byte, headers []Header) error {
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for _, h := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{
			Key:   h.Key,
			Value: h.Value,
		})
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: partition,
		},
		Key:       key,
		Value:     value,
		Headers:   kafkaHeaders,
		Timestamp: time.Now(),
	}

	deliveryChan := make(chan kafka.Event, 1)
	err := p.client.Produce(msg, deliveryChan)
	if err != nil {
		p.logger.OnProduceError(topic, key, err)
		return fmt.Errorf("kafka: failed to produce message to topic %s: %w", topic, err)
	}

	// Wait for delivery report or context cancellation
	select {
	case e := <-deliveryChan:
		m, ok := e.(*kafka.Message)
		if !ok {
			return fmt.Errorf("kafka: unexpected event type from delivery channel")
		}
		if m.TopicPartition.Error != nil {
			p.logger.OnDeliveryFailed(topic, key, m.TopicPartition.Error)
			return fmt.Errorf("kafka: delivery failed for topic %s: %w", topic, m.TopicPartition.Error)
		}
		p.logger.OnDeliverySuccess(topic, m.TopicPartition.Partition, m.TopicPartition.Offset, key)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("kafka: produce to topic %s cancelled: %w", topic, ctx.Err())
	}
}

// Flush waits for all outstanding produce requests to complete.
func (p *producer) Flush(timeoutMs int) int {
	return p.client.Flush(timeoutMs)
}

// Close gracefully shuts down the producer.
func (p *producer) Close() {
	p.client.Flush(5000)
	p.client.Close()
}

// handleEvents processes events from the producer's event channel.
func (p *producer) handleEvents() {
	for e := range p.client.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				topic := ""
				if ev.TopicPartition.Topic != nil {
					topic = *ev.TopicPartition.Topic
				}
				p.logger.OnDeliveryFailed(topic, ev.Key, ev.TopicPartition.Error)
			}
		case kafka.Error:
			p.logger.OnError("producer", ev)
		}
	}
}
