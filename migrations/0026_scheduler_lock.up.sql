-- Advisory-lock helper table for single-instance scheduler leadership (optional metadata).
CREATE TABLE IF NOT EXISTS scheduler_runs (
  job_name TEXT PRIMARY KEY,
  last_run_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
