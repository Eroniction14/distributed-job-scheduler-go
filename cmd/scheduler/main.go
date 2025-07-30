package main

import (
    "log"
    "net/http"

    "distributed-job-scheduler-go/db"
    "distributed-job-scheduler-go/internal/job"
)

func main() {
    // Initialize the DB connection
    db.InitDB()

    // Register the handler from your `job` package
    http.HandleFunc("/api/jobs", job.CreateJobHandler)

    log.Println("✅ Scheduler running on :8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
