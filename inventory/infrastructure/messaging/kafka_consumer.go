package messaging

import (
	"context"
	"fmt"
	"log"

	"github.com/Bass-Peerapon/eventsource-demo/inventory/helper"
	"github.com/IBM/sarama"
)

type Consumer interface {
	StartConsumer(ctx context.Context) error
}

type kafkaConsumer struct {
	brokers []string
	group   string
	topics  []string
	ready   chan bool
}

func NewKafkaConsumer(brokers []string, group string, topics []string) Consumer {
	// NewKafkaConsumer initializes a new Kafka consumer for the specified brokers, group, and topics
	return &kafkaConsumer{
		brokers: brokers,
		group:   group,
		topics:  topics,
		ready:   make(chan bool),
	}
}

// StartConsumer starts the Kafka consumer to consume messages from the given topics
func (kc *kafkaConsumer) StartConsumer(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0 // Make sure the Kafka version matches your Kafka cluster version
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(kc.brokers, kc.group, config)
	if err != nil {
		return fmt.Errorf("error creating consumer group client: %w", err)
	}

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, kc.topics, kc); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			// Check if context was canceled, signaling the consumer to stop
			if ctx.Err() != nil {
				return
			}
			kc.ready = make(chan bool)
		}
	}()

	<-kc.ready // Wait till the consumer has been set up
	log.Println("Sarama consumer up and running!...")
	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (kc *kafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(kc.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (kc *kafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (kc *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Note: Do not use defer here as it will slow down the processing
	for message := range claim.Messages() {
		helper.Println(fmt.Sprintf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic))
		session.MarkMessage(message, "")
	}
	return nil
}
