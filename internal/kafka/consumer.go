package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/Eroniction14/distributed-job-scheduler-go/internal/types"
	kafkago "github.com/segmentio/kafka-go"
)

// StartConsumer connects to the Kafka "jobs.pending" topic and processes
// incoming job messages in an infinite loop.
// It runs as part of the "job-workers" consumer group — if multiple instances
// of this service are running, Kafka automatically distributes messages between them
// so each job is processed by exactly one worker.
// db is passed in for job status updates and execution logging.
func StartConsumer(db *sql.DB) {
	// Create a Kafka reader (consumer) connected to the jobs.pending topic.
	// GroupID "job-workers" ensures messages are load-balanced across all
	// running consumer instances without duplication.
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{"kafka:9092"},
		Topic:   "jobs.pending",
		GroupID: "job-workers",
	})
	// Ensure the reader connection is closed when this function returns.
	defer reader.Close()

	log.Println("🎧 Kafka consumer started")

	// Poll for messages indefinitely.
	// ReadMessage blocks until a message is available, then returns it.
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		// Deserialize the JSON message payload back into a Job struct.
		// The producer serialized the job as JSON before publishing.
		var job types.Job
		if err := json.Unmarshal(msg.Value, &job); err != nil {
			log.Printf("Error parsing job: %v", err)
			continue
		}

		log.Printf("📥 Received job %d: %s", job.ID, job.Command)

		// Execute the job in a separate goroutine so the consumer loop
		// can immediately pick up the next message without waiting for
		// the current job to finish.
		go executeJob(db, job)
	}
}

// executeJob runs a job's shell command and records the result.
// It updates the job status through its lifecycle: running → done/failed.
// Execution output is captured and stored in job_logs for auditing.
func executeJob(db *sql.DB, job types.Job) {
	// Mark the job as running before execution begins.
	_, updErr := db.Exec("UPDATE jobs SET status = 'running' WHERE id = $1", job.ID)
	if updErr != nil {
		log.Printf("Failed to update status to running: %v", updErr)
	}

	// Split the command string into executable + arguments.
	cmdParts := strings.Fields(job.Command)
	if len(cmdParts) == 0 {
		log.Printf("Empty command for job ID %d", job.ID)
		return
	}

	// Execute the command and capture both stdout and stderr.
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.CombinedOutput()

	// Two separate status values for two different tables.
	// jobs table allows: pending|active|paused|running|done|failed
	// job_logs table allows: success|failed
	jobStatus := "done"
	logStatus := "success"
	if err != nil {
		jobStatus = "failed"
		logStatus = "failed"
	}

	// Update the job's final status in PostgreSQL.
	_, updErr = db.Exec("UPDATE jobs SET status = $1 WHERE id = $2", jobStatus, job.ID)
	if updErr != nil {
		log.Printf("Failed to update status: %v", updErr)
	}

	// Insert an execution log entry with the command output and result status.
	_, err = db.Exec(
		"INSERT INTO job_logs (job_id, run_time, result, status) VALUES ($1, $2, $3, $4)",
		job.ID, time.Now(), string(output), logStatus,
	)
	if err != nil {
		log.Printf("Failed to log job %d: %v", job.ID, err)
	}
}
