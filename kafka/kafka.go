// Package kafka provides a testable Kafka producer and consumer library
// built on top of confluent-kafka-go/v2.
//
// Unlike the PHP static-method approach, this package uses interfaces for both
// the Producer and Consumer, as well as for the underlying Kafka client and Logger.
// This makes it straightforward to mock in unit tests without needing a real Kafka broker.
//
// Usage example (Producer):
//
//	cfg := &kafka.ProducerConfig{
//	    Brokers:  "broker1:9092,broker2:9092",
//	    Username: "user",
//	    Password: "pass",
//	}
//	p, err := kafka.NewProducer(cfg, nil) // nil uses default logger
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer p.Close()
//
//	err = p.Publish(ctx, "my-topic", []byte("key"), []byte(`{"data":"value"}`), nil)
//
// Usage example (Consumer):
//
//	cfg := &kafka.ConsumerConfig{
//	    Brokers:  "broker1:9092,broker2:9092",
//	    Username: "user",
//	    Password: "pass",
//	    GroupID:  "my-consumer-group",
//	    Topics:   []string{"my-topic"},
//	}
//	c, err := kafka.NewConsumer(cfg, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer c.Close()
//
//	err = c.Subscribe()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	err = c.Consume(ctx, func(ctx context.Context, msg *kafka.Message) error {
//	    fmt.Printf("Received: %s\n", string(msg.Value))
//	    return nil
//	})
//
// Testing example:
//
//	// Use NewProducerWithClient or NewConsumerWithClient with a mock client
//	mockClient := &MockKafkaProducerClient{}
//	p := kafka.NewProducerWithClient(mockClient, cfg, &kafka.NopLogger{})
package kafka

// Version is the version of this kafka package.
const Version = "1.0.0"
