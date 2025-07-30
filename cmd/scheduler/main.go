package main

import (
    "encoding/json"
    "log"
    "net/http"
)

type CreateJobRequest struct {
    Name     string `json:"name"`
    Command  string `json:"command"`
    Schedule string `json:"schedule"`
}

func createJobHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateJobRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    log.Printf("Job received: %s | %s | %s\n", req.Name, req.Command, req.Schedule)
    w.WriteHeader(http.StatusCreated)
}

func main() {
    http.HandleFunc("/api/jobs", createJobHandler)
    log.Println("Scheduler running on :8080")
    http.ListenAndServe(":8080", nil)
}
