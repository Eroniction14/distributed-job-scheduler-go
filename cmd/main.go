package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Eroniction14/distributed-job-scheduler-go/internal/db"
	"github.com/Eroniction14/distributed-job-scheduler-go/internal/job"
	"github.com/Eroniction14/distributed-job-scheduler-go/internal/kafka"
	"github.com/Eroniction14/distributed-job-scheduler-go/internal/scheduler"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file.
	// In Docker, variables are injected via env_file in docker-compose.yml,
	// so this gracefully falls back if .env is not present.
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables from Docker")
	}

	log.Println("DB Host:", os.Getenv("DB_HOST"))

	// Initialize PostgreSQL connection pool.
	// Reads DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME from environment.
	// Calls log.Fatal if the connection cannot be established.
	dbConn := db.InitDB()

	// Create the Kafka topic "jobs.pending" if it doesn't already exist.
	// Uses retry logic (up to 10 attempts) to wait for Kafka to be ready on startup.
	// Non-fatal — logs a warning and continues if topic creation fails.
	if err := kafka.CreateTopic(); err != nil {
		log.Println("⚠️ Could not create Kafka topic:", err)
	}

	// Start the cron-based scheduler in a background goroutine.
	// Fetches all active jobs from PostgreSQL and schedules them based on their cron expression.
	// Runs job commands on schedule and logs execution results to job_logs.
	go scheduler.StartScheduler(dbConn)

	// Start the Kafka consumer in a background goroutine.
	// Listens to the "jobs.pending" topic as part of the "job-workers" consumer group.
	// Executes job commands, updates job status (running → done/failed), and logs results.
	go kafka.StartConsumer(dbConn)

	// enableCORS is a middleware wrapper that adds CORS headers to every response.
	// This allows the static frontend (served from /static) to call the API.
	// OPTIONS preflight requests are handled and returned immediately.
	enableCORS := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next(w, r)
		}
	}

	// POST /api/jobs — create a new job (validates cron, inserts to DB, publishes to Kafka)
	http.HandleFunc("/api/jobs", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			job.CreateJobHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// GET /api/jobs/:id  — fetch a single job by ID
	// PUT /api/jobs/:id  — update job status (active/paused)
	http.HandleFunc("/api/jobs/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			job.GetJobByIDHandler(w, r)
		} else if r.Method == http.MethodPut {
			job.UpdateJobStatusHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// GET /api/job_logs — fetch the 20 most recent job execution logs
	http.HandleFunc("/api/job_logs", enableCORS(job.GetJobLogsHandler))

	// GET /api/jobs/all — fetch all jobs in the system
	http.HandleFunc("/api/jobs/all", enableCORS(job.GetAllJobsHandler))

	// Serve the static HTML frontend from the ./static directory.
	// index.html provides a simple UI for creating jobs and viewing logs.
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// GET /health — simple health check endpoint for monitoring and K8s liveness probes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("✅ Scheduler running on :8080")

	// Start the HTTP server. Blocks forever, handling incoming requests.
	// log.Fatal if the server fails to start (e.g. port already in use).
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
