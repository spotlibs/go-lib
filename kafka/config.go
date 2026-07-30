package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// SchemaType determines the serialization behavior of the message.
const (
	// Schemaless means no schema and no serialization mechanism.
	Schemaless = 1
	// SchemalessWithSerde means no schema but has JSON encode/decode mechanism.
	SchemalessWithSerde = 2
	// SchemafullWithSerde means has schema (Avro) and encode/decode mechanism.
	SchemafullWithSerde = 3
)

// ProducerConfig holds configuration for creating a Kafka producer.
type ProducerConfig struct {
	Brokers          string
	Username         string
	Password         string
	SecurityProtocol string
	SASLMechanism    string
	CompressionCodec string
	MessageTimeoutMs int
	SocketTimeoutMs  int
	AdditionalConfig map[string]string
}

// ConsumerConfig holds configuration for creating a Kafka consumer.
type ConsumerConfig struct {
	Brokers          string
	Username         string
	Password         string
	SecurityProtocol string
	SASLMechanism    string
	CompressionCodec string
	MessageTimeoutMs int
	SocketTimeoutMs  int
	ClientID         string
	GroupID          string
	Topics           []string
	AutoOffsetReset  string
	PollTimeout      time.Duration
	AdditionalConfig map[string]string
}

// Validate checks the required fields for ProducerConfig.
func (c *ProducerConfig) Validate() error {
	if c.Brokers == "" {
		return fmt.Errorf("kafka: brokers is required")
	}
	if c.Username == "" {
		return fmt.Errorf("kafka: username is required")
	}
	if c.Password == "" {
		return fmt.Errorf("kafka: password is required")
	}
	return nil
}

// Validate checks the required fields for ConsumerConfig.
func (c *ConsumerConfig) Validate() error {
	if c.Brokers == "" {
		return fmt.Errorf("kafka: brokers is required")
	}
	if c.Username == "" {
		return fmt.Errorf("kafka: username is required")
	}
	if c.Password == "" {
		return fmt.Errorf("kafka: password is required")
	}
	if c.GroupID == "" {
		return fmt.Errorf("kafka: group_id is required")
	}
	if len(c.Topics) == 0 {
		return fmt.Errorf("kafka: at least one topic is required")
	}
	return nil
}

// ToConfigMap converts ProducerConfig to kafka.ConfigMap.
func (c *ProducerConfig) ToConfigMap() *kafka.ConfigMap {
	securityProtocol := c.SecurityProtocol
	if securityProtocol == "" {
		securityProtocol = "SASL_SSL"
	}
	saslMechanism := c.SASLMechanism
	if saslMechanism == "" {
		saslMechanism = "PLAIN"
	}
	compressionCodec := c.CompressionCodec
	if compressionCodec == "" {
		compressionCodec = "lz4"
	}
	messageTimeoutMs := c.MessageTimeoutMs
	if messageTimeoutMs == 0 {
		messageTimeoutMs = 8000
	}
	socketTimeoutMs := c.SocketTimeoutMs
	if socketTimeoutMs == 0 {
		socketTimeoutMs = 8000
	}

	cm := &kafka.ConfigMap{
		"bootstrap.servers":  c.Brokers,
		"sasl.username":      c.Username,
		"sasl.password":      c.Password,
		"sasl.mechanism":     saslMechanism,
		"security.protocol":  securityProtocol,
		"compression.codec":  compressionCodec,
		"message.timeout.ms": messageTimeoutMs,
		"socket.timeout.ms":  socketTimeoutMs,
	}

	for k, v := range c.AdditionalConfig {
		_ = cm.SetKey(k, v)
	}

	return cm
}

// ToConfigMap converts ConsumerConfig to kafka.ConfigMap.
func (c *ConsumerConfig) ToConfigMap() *kafka.ConfigMap {
	securityProtocol := c.SecurityProtocol
	if securityProtocol == "" {
		securityProtocol = "SASL_SSL"
	}
	saslMechanism := c.SASLMechanism
	if saslMechanism == "" {
		saslMechanism = "PLAIN"
	}
	compressionCodec := c.CompressionCodec
	if compressionCodec == "" {
		compressionCodec = "lz4"
	}
	messageTimeoutMs := c.MessageTimeoutMs
	if messageTimeoutMs == 0 {
		messageTimeoutMs = 10000
	}
	socketTimeoutMs := c.SocketTimeoutMs
	if socketTimeoutMs == 0 {
		socketTimeoutMs = 10000
	}
	autoOffsetReset := c.AutoOffsetReset
	if autoOffsetReset == "" {
		autoOffsetReset = "earliest"
	}

	cm := &kafka.ConfigMap{
		"bootstrap.servers":  c.Brokers,
		"sasl.username":      c.Username,
		"sasl.password":      c.Password,
		"sasl.mechanism":     saslMechanism,
		"security.protocol":  securityProtocol,
		"compression.codec":  compressionCodec,
		"message.timeout.ms": messageTimeoutMs,
		"socket.timeout.ms":  socketTimeoutMs,
		"group.id":           c.GroupID,
		"auto.offset.reset":  autoOffsetReset,
	}

	if c.ClientID != "" {
		_ = cm.SetKey("client.id", c.ClientID)
	}

	for k, v := range c.AdditionalConfig {
		_ = cm.SetKey(k, v)
	}

	return cm
}
