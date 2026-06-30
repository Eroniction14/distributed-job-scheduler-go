# Distributed Job Scheduler in Go

A distributed job scheduler built in Go with PostgreSQL persistence and Docker. Provides a REST API for creating and managing scheduled jobs, with a clean architecture designed for horizontal scaling.

## Architecture

```
┌─────────────────────────────────────┐
│           HTTP API (Go)             │
│   POST /api/jobs  |  GET /health    │
├─────────────────────────────────────┤
│         Job Handler Layer           │
│   Decode → Validate → Persist       │
├─────────────────────────────────────┤
│        PostgreSQL Database          │
│   jobs (name, command, schedule,    │
│          assigned_worker)           │
└─────────────────────────────────────┘
         Containerized via Docker
```

## Project Structure

```
distributed-job-scheduler-go/
├── cmd/
│   └── main.go           # Entry point, route registration
├── db/
│   └── db.go             # PostgreSQL connection pool
├── internal/
│   └── job/
│       └── handler.go    # Job struct + CreateJobHandler
├── Dockerfile            # Multi-stage Go build
├── docker-compose.yml    # App + PostgreSQL services
├── go.mod
└── go.sum
```

## API

### Create a Job
```
POST /api/jobs
Content-Type: application/json

{
  "name": "backup-job",
  "command": "pg_dump mydb > backup.sql",
  "schedule": "0 2 * * *",
  "assigned_worker": "worker-1"
}
```

Response:
```json
{ "message": "Job created successfully" }
```

### Health Check
```
GET /health
→ 200 OK
```

## Tech Stack

- **Go** — HTTP server using `net/http` (stdlib, no framework overhead)
- **PostgreSQL** — Job persistence with parameterized queries
- **Docker + Docker Compose** — Containerized app and database
- **godotenv** — Environment variable management

## Getting Started

### Prerequisites
- Go 1.21+
- Docker + Docker Compose

### Run with Docker (recommended)

```bash
git clone https://github.com/Eroniction14/distributed-job-scheduler-go.git
cd distributed-job-scheduler-go

docker-compose up --build
```

Server starts on `http://localhost:8080`

### Run locally

```bash
# Set up PostgreSQL and create a .env file
cp .env.example .env
# Edit .env with your DB credentials

go mod download
go run cmd/main.go
```

### Test the API

```bash
# Create a job
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "daily-backup",
    "command": "pg_dump mydb",
    "schedule": "0 2 * * *",
    "assigned_worker": "worker-1"
  }'

# Health check
curl http://localhost:8080/health
```

## Database Schema

```sql
CREATE TABLE jobs (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    command         TEXT NOT NULL,
    schedule        VARCHAR(100),
    assigned_worker VARCHAR(100),
    created_at      TIMESTAMP DEFAULT NOW()
);
```

## Roadmap

- [ ] Job status tracking (pending → running → complete → failed)
- [ ] Worker heartbeat + failure detection
- [ ] Job retry logic with backoff
- [ ] Cron-based job execution engine
- [ ] GET /api/jobs — list and filter jobs
- [ ] Distributed locking (prevent duplicate execution)

## Author

**Eroniction Presley** — Northeastern University, Khoury College of Computer Science
