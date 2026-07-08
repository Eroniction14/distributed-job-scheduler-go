package job

import (
	"encoding/json"
	"net/http"

	"github.com/Eroniction14/distributed-job-scheduler-go/internal/db"
)

// JobLog represents a single job execution record from the job_logs table.
// It captures the outcome of each time a job was run — either by the
// Kafka consumer or the cron scheduler.
type JobLog struct {
	ID      int    `json:"id"`
	JobID   int    `json:"job_id"`
	RunTime string `json:"run_time"`
	Result  string `json:"result"` // command output or execution timestamp
	Status  string `json:"status"` // "success" or "failed"
}

// GetJobLogsHandler handles GET /api/job_logs.
// Returns the 20 most recent job execution logs ordered by run time descending.
// Used by the frontend to display the Job Logs table.
func GetJobLogsHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, job_id, result, status, run_time
		FROM job_logs
		ORDER BY run_time DESC
		LIMIT 20
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		http.Error(w, "Failed to fetch logs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Ensure the database connection is returned to the pool after reading.
	defer rows.Close()

	var logs []JobLog
	for rows.Next() {
		var log JobLog
		if err := rows.Scan(&log.ID, &log.JobID, &log.Result, &log.Status, &log.RunTime); err != nil {
			http.Error(w, "Error scanning logs", http.StatusInternalServerError)
			return
		}
		logs = append(logs, log)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
