package job

import (
	"encoding/json"
	"net/http"

	"github.com/Eroniction14/distributed-job-scheduler-go/db"
)

type Job struct {
	Name           string `json:"name"`
	Command        string `json:"command"`
	Schedule       string `json:"schedule"`
	AssignedWorker string `json:"assigned_worker"`
}

func CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO jobs (name, command, schedule, assigned_worker)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.DB.Exec(query, job.Name, job.Command, job.Schedule, job.AssignedWorker)
	if err != nil {
		http.Error(w, "Failed to insert job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Job created successfully"})
}
