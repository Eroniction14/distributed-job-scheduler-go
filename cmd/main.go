package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Eroniction14/distributed-job-scheduler-go/db"
	"github.com/Eroniction14/distributed-job-scheduler-go/internal/job"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables from Docker")
	}

	log.Println("DB Host:", os.Getenv("DB_HOST"))

	// Initialize DB connection
	db.InitDB()

	// Register routes
	http.HandleFunc("/api/jobs", job.CreateJobHandler)

	// ✅ Add this health check endpoint
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
