package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Eroniction14/distributed-job-scheduler-go/db"
	"github.com/Eroniction14/distributed-job-scheduler-go/internal/job"
	"github.com/Eroniction14/distributed-job-scheduler-go/internal/kafka"
	"github.com/Eroniction14/distributed-job-scheduler-go/scheduler"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables from Docker")
	}

	log.Println("DB Host:", os.Getenv("DB_HOST"))

	dbConn := db.InitDB()

	if err := kafka.CreateTopic(); err != nil {
		log.Println("⚠️ Could not create Kafka topic:", err)
	}

	go scheduler.StartScheduler(dbConn)

	// === CORS Middleware ===
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

	// === Register All Routes ===
	http.HandleFunc("/api/jobs", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			job.CreateJobHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/jobs/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			job.GetJobByIDHandler(w, r)
		} else if r.Method == http.MethodPut {
			job.UpdateJobStatusHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/job_logs", enableCORS(job.GetJobLogsHandler))
	http.HandleFunc("/api/jobs/all", enableCORS(job.GetAllJobsHandler))

	// ✅ Serve static HTML frontend (optional)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("✅ Scheduler running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
