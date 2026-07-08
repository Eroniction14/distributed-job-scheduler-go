# 🧠 Distributed Job Scheduler in Go

A lightweight job scheduling engine using Go, PostgreSQL, and Docker.

## 🚀 Features

- Add jobs with cron-style schedule
- Pause/Resume job execution
- Logs all job runs
- REST API for job management
- Dockerized with PostgreSQL backend

## 🧱 Tech Stack

- Go + net/http
- PostgreSQL
- robfig/cron v3
- Docker & Docker Compose

---

## ⚙️ Setup

### 1. Clone the repo

```bash
git clone https://github.com/<your-username>/distributed-job-scheduler-go.git
cd distributed-job-scheduler-go


DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=job_scheduler

docker-compose up --build

+--------------------------+
|     Go Scheduler API     |
| - REST + Cron Engine     |
+------------+-------------+
             |
             |
+------------v-------------+
|      PostgreSQL DB       |
| - Jobs & Execution Logs  |
+--------------------------+


---

## ✅ 2. 📥 Postman Collection Export

Here’s a JSON template you can copy → save as `JobScheduler.postman_collection.json`, then import into Postman:

```json
{
  "info": {
    "_postman_id": "uuid-here",
    "name": "Job Scheduler API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Create Job",
      "request": {
        "method": "POST",
        "header": [{ "key": "Content-Type", "value": "application/json" }],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"name\": \"Say Hello\",\n  \"command\": \"echo Hello\",\n  \"schedule\": \"*/1 * * * *\",\n  \"status\": \"active\"\n}"
        },
        "url": { "raw": "http://localhost:8080/api/jobs", "protocol": "http", "host": ["localhost"], "port": "8080", "path": ["api", "jobs"] }
      }
    },
    {
      "name": "List All Jobs",
      "request": {
        "method": "GET",
        "url": { "raw": "http://localhost:8080/api/jobs/all", "protocol": "http", "host": ["localhost"], "port": "8080", "path": ["api", "jobs", "all"] }
      }
    },
    {
      "name": "Get Job By ID",
      "request": {
        "method": "GET",
        "url": { "raw": "http://localhost:8080/api/jobs/1", "protocol": "http", "host": ["localhost"], "port": "8080", "path": ["api", "jobs", "1"] }
      }
    },
    {
      "name": "Update Job Status (Pause/Resume)",
      "request": {
        "method": "PUT",
        "header": [{ "key": "Content-Type", "value": "application/json" }],
        "body": {
          "mode": "raw",
          "raw": "{ \"status\": \"paused\" }"
        },
        "url": { "raw": "http://localhost:8080/api/jobs/1", "protocol": "http", "host": ["localhost"], "port": "8080", "path": ["api", "jobs", "1"] }
      }
    },
    {
      "name": "Get Job Logs",
      "request": {
        "method": "GET",
        "url": { "raw": "http://localhost:8080/api/job_logs", "protocol": "http", "host": ["localhost"], "port": "8080", "path": ["api", "job_logs"] }
      }
    }
  ]
}
