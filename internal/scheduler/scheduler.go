package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Eroniction14/distributed-job-scheduler-go/internal/types"
	"github.com/robfig/cron/v3"
)

// StartScheduler loads all active jobs from PostgreSQL and schedules them
// using their cron expressions. It runs indefinitely as a background goroutine.
// Note: jobs created after startup are NOT automatically picked up by this scheduler —
// they are handled immediately by the Kafka consumer instead.
// This scheduler handles recurring execution of pre-existing active jobs.
func StartScheduler(db *sql.DB) {
	// Create a new cron scheduler instance.
	c := cron.New()

	// Fetch all currently active jobs from the database at startup.
	jobs, err := fetchActiveJobs(db)
	if err != nil {
		log.Fatalf("Failed to fetch jobs: %v", err)
	}

	// Register each job with the cron scheduler using its schedule expression.
	// jobCopy is used to avoid the loop variable capture issue in Go —
	// without it, all goroutines would reference the same loop variable.
	for _, job := range jobs {
		jobCopy := job
		_, err := c.AddFunc(job.Schedule, func() {
			runJob(db, jobCopy)
		})
		if err != nil {
			log.Printf("Invalid cron format for job %d: %v", job.ID, err)
		} else {
			log.Printf("Scheduled job %d (%s)", job.ID, job.Name)
		}
	}

	// Start the cron scheduler in its own goroutine (non-blocking).
	c.Start()

	// Block forever — keeps this goroutine alive so the cron scheduler
	// continues running in the background.
	select {}
}

// fetchActiveJobs queries PostgreSQL for all jobs with status = 'active'.
// These are the jobs that will be registered with the cron scheduler.
func fetchActiveJobs(db *sql.DB) ([]types.Job, error) {
	rows, err := db.Query(`
		SELECT id, name, command, schedule, status
		FROM jobs
		WHERE status = 'active'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []types.Job
	for rows.Next() {
		var job types.Job
		if err := rows.Scan(&job.ID, &job.Name, &job.Command, &job.Schedule, &job.Status); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// runJob is called by the cron scheduler each time a job's schedule fires.
// It logs the execution time to job_logs and updates the job's last_run timestamp.
// Note: this function logs a timestamp as the result rather than executing the
// actual shell command — full command execution with output capture is handled
// by the Kafka consumer's executeJob function.
func runJob(db *sql.DB, job types.Job) {
	log.Printf("Running job ID %d: %s", job.ID, job.Command)
	result := fmt.Sprintf("Executed at %s", time.Now().Format(time.RFC3339))

	// Insert an execution record into job_logs.
	_, err := db.Exec(
		"INSERT INTO job_logs (job_id, run_time, result, status) VALUES ($1, $2, $3, $4)",
		job.ID, time.Now(), result, "success",
	)
	if err != nil {
		log.Printf("Failed to log job %d: %v", job.ID, err)
	}

	// Update the job's last_run timestamp so the UI can show when it last ran.
	_, err = db.Exec("UPDATE jobs SET last_run = $1 WHERE id = $2", time.Now(), job.ID)
	if err != nil {
		log.Printf("Failed to update last_run for job %d: %v", job.ID, err)
	}
}
