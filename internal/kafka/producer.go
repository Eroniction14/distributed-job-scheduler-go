package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

// PublishJob sends a job message to the "jobs.pending" Kafka topic.
// jobID is used as the message key for Kafka partitioning.
// jobBytes is the JSON-encoded job payload that the consumer will decode.
// Returns an error if the message could not be published.
func PublishJob(jobID int, jobBytes []byte) error {
	// Create a new Kafka writer (producer) targeting the jobs.pending topic.
	// A new writer is created per publish call — for high-throughput production use,
	// consider creating a shared writer at startup instead.
	writer := kafkago.NewWriter(kafkago.WriterConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "jobs.pending",
	})

	err := writer.WriteMessages(context.Background(),
		kafkago.Message{
			// Key is the job ID as bytes — used by Kafka to route messages
			// to consistent partitions (all messages for the same job go to the same partition).
			Key: []byte(fmt.Sprintf("%d", jobID)),
			// Value is the full JSON-encoded job struct.
			// The consumer will unmarshal this back into a types.Job.
			Value: jobBytes,
		},
	)
	if err != nil {
		log.Printf("❌ Failed to publish job %d to Kafka: %v", jobID, err)
	} else {
		log.Printf("✅ Published job %d to Kafka", jobID)
	}
	return err
}

// CreateTopic ensures the "jobs.pending" Kafka topic exists.
// Called once at application startup before the consumer or producer are used.
// Uses retry logic (up to 10 attempts, 3 seconds apart) to wait for Kafka
// to be fully ready — necessary because the scheduler container starts before
// Kafka finishes initializing in Docker Compose.
// Non-fatal if topic creation fails (topic may already exist from a previous run).
func CreateTopic() error {
	var conn *kafkago.Conn
	var err error

	// Retry connecting to Kafka up to 10 times.
	// Kafka takes a few seconds after the container starts to accept connections.
	for i := 0; i < 10; i++ {
		conn, err = kafkago.Dial("tcp", "kafka:9092")
		if err == nil {
			break
		}
		log.Printf("Waiting for Kafka... attempt %d/10", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return err
	}
	// Close the admin connection when done — this is separate from the producer/consumer connections.
	defer conn.Close()

	// Create the topic with a single partition and replication factor of 1.
	// In production with multiple Kafka brokers, increase ReplicationFactor
	// to match the number of brokers for fault tolerance.
	err = conn.CreateTopics(kafkago.TopicConfig{
		Topic:             "jobs.pending",
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	return err
}
