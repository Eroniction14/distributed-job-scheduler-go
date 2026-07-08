package job

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Eroniction14/distributed-job-scheduler-go/internal/db"
	"github.com/Eroniction14/distributed-job-scheduler-go/internal/kafka"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type Job struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Command  string  `json:"command"`
	Schedule string  `json:"schedule"`
	Status   string  `json:"status"`
	LastRun  *string `json:"last_run"` // nullable
}

func CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Cron validation
	if _, err := cron.ParseStandard(job.Schedule); err != nil {
		http.Error(w, "Invalid cron expression: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := `
    INSERT INTO jobs (name, command, schedule, status)
    VALUES ($1, $2, $3, $4)
    RETURNING id
	`

	var jobID int
	err := db.DB.QueryRow(query, job.Name, job.Command, job.Schedule, job.Status).Scan(&jobID)
	job.ID = jobID
	if err != nil {
		http.Error(w, "Failed to insert job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jobBytes, err := json.Marshal(job)
	if err != nil {
		log.Println("Failed to marshal job:", err)
	}

	err = kafka.PublishJob(jobID, jobBytes)
	if err != nil {
		log.Println("Failed to publish to Kafka:", err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Job created successfully"})
}

func GetJobByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	if idStr == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	query := `SELECT id, name, command, schedule, status, last_run FROM jobs WHERE id = $1`
	row := db.DB.QueryRow(query, idStr)

	var job Job
	var lastRun sql.NullString

	err := row.Scan(&job.ID, &job.Name, &job.Command, &job.Schedule, &job.Status, &lastRun)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	if lastRun.Valid {
		job.LastRun = &lastRun.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func UpdateJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	if idStr == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Status string `json:"status"` // expected: "active" or "paused"
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if body.Status != "active" && body.Status != "paused" {
		http.Error(w, "Invalid status: must be 'active' or 'paused'", http.StatusBadRequest)
		return
	}

	query := `UPDATE jobs SET status = $1 WHERE id = $2`
	result, err := db.DB.Exec(query, body.Status, idStr)
	if err != nil {
		http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Job status updated successfully",
	})
}

func GetAllJobsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, name, command, schedule, status, last_run FROM jobs")
	if err != nil {
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		var lastRun sql.NullString
		if err := rows.Scan(&job.ID, &job.Name, &job.Command, &job.Schedule, &job.Status, &lastRun); err != nil {
			http.Error(w, "Failed to parse job", http.StatusInternalServerError)
			return
		}
		if lastRun.Valid {
			job.LastRun = &lastRun.String
		}
		jobs = append(jobs, job)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func RunJobNow(c *gin.Context) {
	id := c.Param("id")
	jobID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	// Fetch job from DB
	var job Job
	var lastRun sql.NullString
	query := `SELECT id, name, command, schedule, status, last_run FROM jobs WHERE id = $1`
	row := db.DB.QueryRow(query, jobID)
	err = row.Scan(&job.ID, &job.Name, &job.Command, &job.Schedule, &job.Status, &lastRun)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	// Run the job command
	go func(job Job) {
		cmdParts := strings.Fields(job.Command)
		if len(cmdParts) == 0 {
			return
		}
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		output, err := cmd.CombinedOutput()

		logQuery := `INSERT INTO job_logs (job_id, output, success, run_at) VALUES ($1, $2, $3, $4)`
		_, _ = db.DB.Exec(logQuery, job.ID, string(output), err == nil, time.Now())
	}(job)

	c.JSON(http.StatusOK, gin.H{"message": "job triggered"})
}
