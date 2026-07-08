package job

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Eroniction14/distributed-job-scheduler-go/internal/db"
	"github.com/Eroniction14/distributed-job-scheduler-go/internal/kafka"
	"github.com/Eroniction14/distributed-job-scheduler-go/internal/types"
	"github.com/robfig/cron/v3"
)

// CreateJobHandler handles POST /api/jobs.
// It validates the request, inserts the job into PostgreSQL,
// and publishes it to the Kafka "jobs.pending" topic for immediate execution.
func CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	// Decode the JSON request body into a Job struct.
	var job types.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate the cron schedule expression before inserting.
	// e.g. "* * * * *" is valid, "invalid" is not.
	if _, err := cron.ParseStandard(job.Schedule); err != nil {
		http.Error(w, "Invalid cron expression: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Insert the job into PostgreSQL and retrieve the auto-generated ID.
	// RETURNING id avoids a separate SELECT query to get the new row's ID.
	query := `
		INSERT INTO jobs (name, command, schedule, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var jobID int
	err := db.DB.QueryRow(query, job.Name, job.Command, job.Schedule, job.Status).Scan(&jobID)
	if err != nil {
		http.Error(w, "Failed to insert job: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Set the real DB-generated ID on the struct before publishing to Kafka.
	// Without this, the published message would have ID=0 (Go's zero value for int).
	job.ID = jobID

	// Serialize the job to JSON bytes for the Kafka message payload.
	jobBytes, err := json.Marshal(job)
	if err != nil {
		log.Println("Failed to marshal job:", err)
	}

	// Publish the job to Kafka. Non-fatal — the job is already saved in PostgreSQL,
	// so we log the error but still return success to the client.
	err = kafka.PublishJob(jobID, jobBytes)
	if err != nil {
		log.Println("Failed to publish to Kafka:", err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Job created successfully"})
}

// GetJobByIDHandler handles GET /api/jobs/:id.
// Returns a single job by its database ID.
func GetJobByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the job ID from the URL path.
	// e.g. /api/jobs/42 → "42"
	idStr := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	if idStr == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	query := `SELECT id, name, command, schedule, status, last_run FROM jobs WHERE id = $1`
	row := db.DB.QueryRow(query, idStr)

	var job types.Job
	// sql.NullString handles the nullable last_run column.
	// A plain string would panic when PostgreSQL returns NULL.
	var lastRun sql.NullString

	err := row.Scan(&job.ID, &job.Name, &job.Command, &job.Schedule, &job.Status, &lastRun)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Only set LastRun if the value is not NULL.
	if lastRun.Valid {
		job.LastRun = &lastRun.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// UpdateJobStatusHandler handles PUT /api/jobs/:id.
// Allows toggling a job between "active" and "paused" status.
// Only these two values are accepted — other status values (running, done, failed)
// are managed internally by the scheduler and consumer.
func UpdateJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	if idStr == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Restrict status updates to only user-controlled values.
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

	// Check that the job actually existed — if 0 rows were affected, the ID was not found.
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

// GetAllJobsHandler handles GET /api/jobs/all.
// Returns all jobs in the system regardless of status.
func GetAllJobsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, name, command, schedule, status, last_run FROM jobs")
	if err != nil {
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		return
	}
	// Ensure rows are closed after we're done reading,
	// releasing the database connection back to the pool.
	defer rows.Close()

	var jobs []types.Job
	for rows.Next() {
		var job types.Job
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
