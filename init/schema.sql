CREATE TABLE IF NOT EXISTS jobs (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  schedule TEXT NOT NULL,
  command TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('pending', 'active', 'paused', 'running', 'done', 'failed')),
  last_run TIMESTAMP,
  worker_id TEXT
);

CREATE TABLE IF NOT EXISTS job_logs (
  id SERIAL PRIMARY KEY,
  job_id INTEGER REFERENCES jobs(id) ON DELETE SET NULL,
  run_time TIMESTAMP NOT NULL,
  result TEXT,
  status TEXT CHECK (status IN ('success', 'failed'))
);
