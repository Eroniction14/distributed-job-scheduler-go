package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type JobMessage struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Command  string `json:"command"`
	Schedule string `json:"schedule"`
	Status   string `json:"status"`
}

func StartConsumer(db *sql.DB) {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "jobs.pending",
		GroupID: "job-workers",
	})
	defer reader.Close()

	log.Println("🎧 Kafka consumer started")

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var job JobMessage
		if err := json.Unmarshal(msg.Value, &job); err != nil {
			log.Printf("Error parsing job: %v", err)
			continue
		}

		log.Printf("📥 Received job %d: %s", job.ID, job.Command)
		go executeJob(db, job)
	}
}

func executeJob(db *sql.DB, job JobMessage) {

	_, updErr := db.Exec("UPDATE jobs SET status = 'running' WHERE id = $1", job.ID)
	if updErr != nil {
		log.Printf("Failed to update status to running: %v", updErr)
	}

	cmdParts := strings.Fields(job.Command)
	if len(cmdParts) == 0 {
		log.Printf("Empty command for job ID %d", job.ID)
		return
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.CombinedOutput()

	status := "done"
	if err != nil {
		status = "failed"
	}

	_, updErr = db.Exec("UPDATE jobs SET status = $1 WHERE id = $2", status, job.ID)

	if updErr != nil {
		log.Printf("Failed to update status: %v", updErr)
	}

	_, err = db.Exec(
		"INSERT INTO job_logs (job_id, run_time, result, status) VALUES ($1, $2, $3, $4)",
		job.ID, time.Now(), string(output), status,
	)
	if err != nil {
		log.Printf("Failed to log job %d: %v", job.ID, err)
	}
}
