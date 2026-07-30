package kafka

import (
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Logger defines callback/logging interface for Kafka events.
// Implement this interface to integrate with your application's logging system.
// This replaces the static KafkaCallable pattern from PHP with a testable interface.
type Logger interface {
	// OnDeliverySuccess is called when a message is successfully delivered to Kafka.
	OnDeliverySuccess(topic string, partition int32, offset kafka.Offset, key []byte)
	// OnDeliveryFailed is called when a message delivery fails.
	OnDeliveryFailed(topic string, key []byte, err error)
	// OnProduceError is called when the producer fails to enqueue a message.
	OnProduceError(topic string, key []byte, err error)
	// OnError is called for general Kafka errors (producer or consumer).
	OnError(component string, err kafka.Error)
	// OnRebalance is called when consumer group rebalancing occurs.
	OnRebalance(partitions []kafka.TopicPartition, assigned bool)
	// OnConsume is called when a message is consumed.
	OnConsume(topic string, partition int32, offset int64, key []byte)
	// OnConsumeError is called when the message handler returns an error.
	OnConsumeError(topic string, key []byte, err error)
	// OnCommitSuccess is called when an offset commit succeeds.
	OnCommitSuccess(topic string, partition int32, offset int64)
	// OnCommitError is called when an offset commit fails.
	OnCommitError(topic string, partition int32, offset int64, err error)
	// OnLog is called for general Kafka log messages.
	OnLog(level int, facility string, message string)
}

// defaultLogger is the default implementation of Logger that uses Go's standard log package.
type defaultLogger struct{}

func (l *defaultLogger) OnDeliverySuccess(topic string, partition int32, offset kafka.Offset, key []byte) {
	log.Printf("[kafka] delivery success. topic=%s partition=%d offset=%v key=%s",
		topic, partition, offset, string(key))
}

func (l *defaultLogger) OnDeliveryFailed(topic string, key []byte, err error) {
	log.Printf("[kafka] delivery failed. topic=%s key=%s error=%v",
		topic, string(key), err)
}

func (l *defaultLogger) OnProduceError(topic string, key []byte, err error) {
	log.Printf("[kafka] produce error. topic=%s key=%s error=%v",
		topic, string(key), err)
}

func (l *defaultLogger) OnError(component string, err kafka.Error) {
	log.Printf("[kafka] %s error. code=%v message=%s",
		component, err.Code(), err.String())
}

func (l *defaultLogger) OnRebalance(partitions []kafka.TopicPartition, assigned bool) {
	status := "revoked"
	if assigned {
		status = "assigned"
	}
	for _, p := range partitions {
		topic := ""
		if p.Topic != nil {
			topic = *p.Topic
		}
		log.Printf("[kafka] rebalance %s. topic=%s partition=%d offset=%v",
			status, topic, p.Partition, p.Offset)
	}
}

func (l *defaultLogger) OnConsume(topic string, partition int32, offset int64, key []byte) {
	log.Printf("[kafka] consumed. topic=%s partition=%d offset=%d key=%s",
		topic, partition, offset, string(key))
}

func (l *defaultLogger) OnConsumeError(topic string, key []byte, err error) {
	log.Printf("[kafka] consume handler error. topic=%s key=%s error=%v",
		topic, string(key), err)
}

func (l *defaultLogger) OnCommitSuccess(topic string, partition int32, offset int64) {
	log.Printf("[kafka] offset committed. topic=%s partition=%d offset=%d",
		topic, partition, offset)
}

func (l *defaultLogger) OnCommitError(topic string, partition int32, offset int64, err error) {
	log.Printf("[kafka] offset commit failed. topic=%s partition=%d offset=%d error=%v",
		topic, partition, offset, err)
}

func (l *defaultLogger) OnLog(level int, facility string, message string) {
	var lvl string
	switch {
	case level <= 3: // LOG_ERR, LOG_CRIT, LOG_EMERG
		lvl = "ERROR"
	case level <= 5: // LOG_WARNING, LOG_NOTICE
		lvl = "WARN"
	default: // LOG_INFO, LOG_DEBUG
		lvl = "INFO"
	}
	log.Printf("[kafka] %s [%s] %s", lvl, facility, message)
}

// NopLogger is a no-op logger implementation that discards all log events.
// Useful for testing when you don't want log output.
type NopLogger struct{}

func (l *NopLogger) OnDeliverySuccess(topic string, partition int32, offset kafka.Offset, key []byte) {
}
func (l *NopLogger) OnDeliveryFailed(topic string, key []byte, err error) {}
func (l *NopLogger) OnProduceError(topic string, key []byte, err error)   {}
func (l *NopLogger) OnError(component string, err kafka.Error)            {}
func (l *NopLogger) OnRebalance(partitions []kafka.TopicPartition, assigned bool) {
}
func (l *NopLogger) OnConsume(topic string, partition int32, offset int64, key []byte) {
}
func (l *NopLogger) OnConsumeError(topic string, key []byte, err error)          {}
func (l *NopLogger) OnCommitSuccess(topic string, partition int32, offset int64) {}
func (l *NopLogger) OnCommitError(topic string, partition int32, offset int64, err error) {
}
func (l *NopLogger) OnLog(level int, facility string, message string) {}

// Ensure interface compliance at compile time.
var _ Logger = (*defaultLogger)(nil)
var _ Logger = (*NopLogger)(nil)

// formatPartitions formats TopicPartition slice into a readable string.
func formatPartitions(partitions []kafka.TopicPartition) string {
	result := ""
	for i, p := range partitions {
		topic := ""
		if p.Topic != nil {
			topic = *p.Topic
		}
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s[%d]@%v", topic, p.Partition, p.Offset)
	}
	return result
}
