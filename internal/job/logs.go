package job

import (
	"encoding/json"
	"net/http"

	"github.com/Eroniction14/distributed-job-scheduler-go/internal/db"
)

type JobLog struct {
	ID      int    `json:"id"`
	JobID   int    `json:"job_id"`
	RunTime string `json:"run_time"`
	Result  string `json:"result"`
	Status  string `json:"status"`
}

func GetJobLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Optional: filter by job_id using /api/logs?job_id=1
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
