package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Bass-Peerapon/eventsource-demo/inventory/infrastructure/messaging"
)

var (
	KAFKA_BROKERS      = strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	ORDER_EVENT_GROUP  = os.Getenv("ORDER_EVENT_GROUP")
	ORDER_EVENT_TOPICS = strings.Split(os.Getenv("ORDER_EVENT_TOPICS"), ",")
)

func main() {
	// Connect to Kafka
	kafkaConsumer := messaging.NewKafkaConsumer(KAFKA_BROKERS, ORDER_EVENT_GROUP, ORDER_EVENT_TOPICS)

	// Create a new context to control consumer shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Capture interrupt signals to gracefully shut down the consumer
	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		<-sigterm
		log.Println("Received shutdown signal, stopping consumer...")
		cancel()
	}()

	// Start the consumer
	if err := kafkaConsumer.StartConsumer(ctx); err != nil {
		log.Fatalf("Error starting consumer: %v", err)
	}

	// Keep the application running to allow the consumer to process messages
	select {
	case <-ctx.Done():
		log.Println("Shutting down application...")
		time.Sleep(2 * time.Second) // Give time for cleanup
	}
}
