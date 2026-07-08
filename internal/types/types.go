// Package types defines shared structs used across packages to avoid circular imports.
package types

// Job maps to the jobs table in PostgreSQL.
// Used by the API handlers, Kafka producer/consumer, and cron scheduler.
type Job struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Command  string  `json:"command"`  // shell command to execute
	Schedule string  `json:"schedule"` // cron expression e.g. "0 2 * * *"
	Status   string  `json:"status"`   // pending|active|paused|running|done|failed
	LastRun  *string `json:"last_run"` // pointer = nullable, nil for new jobs
}
