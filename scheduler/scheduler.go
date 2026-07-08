package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/Eroniction14/distributed-job-scheduler-go/db"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
)

type Job struct {
	ID       int
	Name     string
	Schedule string
	Command  string
	Status   string
}

func StartScheduler(db *sql.DB) {
	c := cron.New()

	jobs, err := fetchActiveJobs(db)
	if err != nil {
		log.Fatalf("Failed to fetch jobs: %v", err)
	}

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

	c.Start()
	select {} // Block forever
}

func fetchActiveJobs(db *sql.DB) ([]Job, error) {
	rows, err := db.Query(`
	SELECT id, name, command, schedule, status
	FROM jobs
	WHERE status = 'active'
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.ID, &job.Name, &job.Command, &job.Schedule, &job.Status); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func runJob(db *sql.DB, job Job) {
	log.Printf("Running job ID %d: %s", job.ID, job.Command)
	result := fmt.Sprintf("Executed at %s", time.Now().Format(time.RFC3339))

	_, err := db.Exec(
		"INSERT INTO job_logs (job_id, run_time, result, status) VALUES ($1, $2, $3, $4)",
		job.ID, time.Now(), result, "success",
	)
	if err != nil {
		log.Printf("Failed to log job %d: %v", job.ID, err)
	}

	_, err = db.Exec("UPDATE jobs SET last_run = $1 WHERE id = $2", time.Now(), job.ID)
	if err != nil {
		log.Printf("Failed to update last_run for job %d: %v", job.ID, err)
	}
}

// ExecuteJob manually runs a job command and logs the result using raw SQL
func ExecuteJob(job Job) {
	cmdParts := strings.Fields(job.Command)
	if len(cmdParts) == 0 {
		log.Printf("Empty command for job ID %d", job.ID)
		return
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.CombinedOutput()

	status := "success"
	if err != nil {
		status = "failed"
	}

	_, logErr := db.DB.Exec(`
		INSERT INTO job_logs (job_id, run_time, result, status)
		VALUES ($1, $2, $3, $4)
	`, job.ID, time.Now(), string(output), status)
	if logErr != nil {
		log.Printf("Failed to log job execution: %v", logErr)
	}

	_, updErr := db.DB.Exec("UPDATE jobs SET last_run = $1 WHERE id = $2", time.Now(), job.ID)
	if updErr != nil {
		log.Printf("Failed to update last_run: %v", updErr)
	}
}
