package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

func PublishJob(jobID int, jobBytes []byte) error {

	writer := kafkago.NewWriter(kafkago.WriterConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "jobs.pending",
	})
	err := writer.WriteMessages(context.Background(),
		kafkago.Message{
			Key:   []byte(fmt.Sprintf("%d", jobID)),
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

func CreateTopic() error {
	var conn *kafkago.Conn
	var err error

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
	defer conn.Close()

	err = conn.CreateTopics(kafkago.TopicConfig{
		Topic:             "jobs.pending",
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	return err
}
