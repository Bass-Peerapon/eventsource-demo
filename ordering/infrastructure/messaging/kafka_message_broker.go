package messaging

import (
	"fmt"

	"github.com/IBM/sarama"
)

const TOPIC_ORDER_EVENT = "ORDER_EVENT"

type MessageBroker interface {
	Publish(topic string, key string, value []byte) error
}

type kafkaMessageBroker struct {
	producer sarama.SyncProducer
}

// Publish implements MessageBroker.
func (k *kafkaMessageBroker) Publish(topic string, key string, value []byte) error {
	_, _, err := k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	})
	return err
}

func NewKafaMessageBroker(brokers []string) MessageBroker {
	// Create admin client
	admin, err := newKafkaAdmin(brokers)
	if err != nil {
		panic(err)
	}
	defer admin.Close()

	// Create topic if it does not exist
	err = createTopicIfNotExists(admin, TOPIC_ORDER_EVENT)
	if err != nil {
		panic(err)
	}

	// Create producer
	producer, err := newKafkaProducer(brokers)
	if err != nil {
		panic(err)
	}

	return &kafkaMessageBroker{
		producer: producer,
	}
}

func newKafkaProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func newKafkaAdmin(brokers []string) (sarama.ClusterAdmin, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0 // Ensure the Kafka version is compatible

	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func createTopicIfNotExists(admin sarama.ClusterAdmin, topic string) error {
	topics, err := admin.ListTopics()
	if err != nil {
		return fmt.Errorf("failed to list topics: %w", err)
	}

	// Check if the topic already exists
	if _, ok := topics[topic]; ok {
		fmt.Printf("Topic %s already exists\n", topic)
		return nil
	}

	// Define topic details
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
		ConfigEntries:     make(map[string]*string),
	}

	// Create the topic
	err = admin.CreateTopic(topic, topicDetail, false)
	if err != nil {
		return fmt.Errorf("failed to create topic %s: %w", topic, err)
	}

	fmt.Printf("Topic %s created successfully\n", topic)
	return nil
}
