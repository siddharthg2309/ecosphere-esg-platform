CREATE TABLE department_scores (
  id UUID PRIMARY KEY,
  department_id UUID NOT NULL REFERENCES departments(id),
  period TEXT NOT NULL,
  environmental INT NOT NULL DEFAULT 0 CHECK (environmental BETWEEN 0 AND 100),
  social INT NOT NULL DEFAULT 0 CHECK (social BETWEEN 0 AND 100),
  governance INT NOT NULL DEFAULT 0 CHECK (governance BETWEEN 0 AND 100),
  total INT NOT NULL DEFAULT 0 CHECK (total BETWEEN 0 AND 100),
  computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (department_id, period)
);

CREATE INDEX idx_department_scores_period ON department_scores(period);
