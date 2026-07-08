# Distributed Job Scheduler in Go

A fault-tolerant distributed job scheduler built in Go with Kafka-based job distribution, PostgreSQL persistence, cron scheduling, and Kubernetes deployment manifests.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP API (Go)                         │
│         POST /api/jobs  |  GET /health                  │
└──────────────────────┬──────────────────────────────────┘
                       │ INSERT + PUBLISH
          ┌────────────▼────────────┐
          │      PostgreSQL         │
          │  jobs, job_logs tables  │
          └────────────────────────┘
                       │ PUBLISH
          ┌────────────▼────────────┐
          │    Kafka (jobs.pending) │
          │    Topic + Zookeeper    │
          └────────────────────────┘
                       │ CONSUME
          ┌────────────▼────────────┐
          │     Worker Pool         │
          │  (Kafka Consumer Group) │
          │  execute → log → update │
          └────────────────────────┘
```

## How It Works

1. Client sends `POST /api/jobs` with a job name, command, and cron schedule
2. API server validates the cron expression, inserts the job into PostgreSQL, and publishes it to the `jobs.pending` Kafka topic
3. Kafka consumer workers pick up the message, execute the shell command, and update the job status (`pending → running → done/failed`)
4. Execution results are logged to `job_logs` table
5. Cron scheduler also runs active jobs on their defined schedule

## Project Structure

```
distributed-job-scheduler-go/
├── cmd/
│   └── main.go                  # Entry point, wires all components
├── db/
│   └── db.go                    # PostgreSQL connection pool
├── internal/
│   ├── job/
│   │   ├── handler.go           # CRUD API handlers
│   │   └── logs.go              # Job logs handler
│   └── kafka/
│       ├── producer.go          # Kafka producer + topic creation
│       └── consumer.go          # Kafka consumer + job execution
├── scheduler/
│   └── scheduler.go             # Cron-based job scheduler
├── k8s/
│   ├── configmap.yaml           # Environment configuration
│   ├── postgres.yaml            # PostgreSQL deployment + service
│   ├── kafka.yaml               # Kafka + Zookeeper deployment
│   └── scheduler.yaml           # Go app deployment (3 replicas) + service
├── init/
│   └── schema.sql               # PostgreSQL schema
├── static/
│   └── index.html               # Web UI
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Tech Stack

- **Go** — HTTP API, Kafka producer/consumer, cron scheduler
- **PostgreSQL** — Job persistence with status tracking and execution logs
- **Apache Kafka** — Distributed job queue (jobs.pending topic)
- **Zookeeper** — Kafka cluster coordination
- **Docker + Docker Compose** — Local containerized development
- **Kubernetes** — Production deployment manifests (3 scheduler replicas)

## API Endpoints

### Create a Job
```
POST /api/jobs
Content-Type: application/json

{
  "name": "daily-backup",
  "command": "pg_dump mydb",
  "schedule": "0 2 * * *",
  "status": "active"
}
```

### Get All Jobs
```
GET /api/jobs/all
```

### Get Job by ID
```
GET /api/jobs/:id
```

### Update Job Status
```
PUT /api/jobs/:id
{ "status": "paused" }
```

### Get Job Logs
```
GET /api/job_logs
```

### Health Check
```
GET /health → 200 OK
```

## Database Schema

```sql
CREATE TABLE jobs (
  id       SERIAL PRIMARY KEY,
  name     TEXT NOT NULL,
  command  TEXT NOT NULL,
  schedule TEXT NOT NULL,
  status   TEXT NOT NULL CHECK (status IN ('pending','active','paused','running','done','failed')),
  last_run TIMESTAMP,
  worker_id TEXT
);

CREATE TABLE job_logs (
  id       SERIAL PRIMARY KEY,
  job_id   INTEGER REFERENCES jobs(id) ON DELETE SET NULL,
  run_time TIMESTAMP NOT NULL,
  result   TEXT,
  status   TEXT CHECK (status IN ('success', 'failed'))
);
```

## Getting Started

### Prerequisites
- Go 1.23+
- Docker + Docker Compose

### Run locally

```bash
git clone https://github.com/Eroniction14/distributed-job-scheduler-go.git
cd distributed-job-scheduler-go

# Create .env file
cp .env.example .env

docker-compose up --build
```

Open `http://localhost:8080` for the web UI.

### Test the API

```bash
# Create a job
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "daily-backup",
    "command": "echo hello",
    "schedule": "* * * * *",
    "status": "active"
  }'

# Health check
curl http://localhost:8080/health
```

## Kubernetes Deployment

Apply all manifests to a cluster:

```bash
kubectl apply -f k8s/
```

Services:
- Scheduler API: NodePort 30080
- Postgres: ClusterIP on 5432
- Kafka: ClusterIP on 9092

Scale workers:
```bash
kubectl scale deployment scheduler --replicas=5
```

## Roadmap

- [x] REST API for job management
- [x] Cron-based job scheduling
- [x] Kafka producer — publish jobs on creation
- [x] Kafka consumer — distributed job execution
- [x] Job status tracking (pending → running → done/failed)
- [x] Execution logging to job_logs
- [x] Kubernetes manifests
- [ ] Worker heartbeat + failure detection
- [ ] Job retry logic with backoff
- [ ] Terraform infrastructure module
- [ ] Prometheus + Grafana observability

## Author

**Eroniction Presley** — Northeastern University, Khoury College of Computer Science