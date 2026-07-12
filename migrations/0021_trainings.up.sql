CREATE TABLE trainings (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  assigned_to TEXT NOT NULL DEFAULT 'All employees',
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE training_completions (
  id UUID PRIMARY KEY,
  employee_id UUID NOT NULL REFERENCES users(id),
  training_id UUID NOT NULL REFERENCES trainings(id),
  completed_at DATE NOT NULL DEFAULT CURRENT_DATE,
  UNIQUE (employee_id, training_id)
);

CREATE INDEX idx_training_completions_training ON training_completions(training_id);

-- Optional demographics for diversity dashboard aggregation
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS gender TEXT NOT NULL DEFAULT 'prefer_not',
  ADD COLUMN IF NOT EXISTS is_leadership BOOLEAN NOT NULL DEFAULT false;
