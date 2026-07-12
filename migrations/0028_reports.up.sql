CREATE TABLE reports (
  id UUID PRIMARY KEY,
  type TEXT NOT NULL,
  filters JSONB NOT NULL DEFAULT '{}',
  result JSONB NOT NULL DEFAULT '{}',
  generated_by UUID REFERENCES users(id),
  generated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reports_generated_at ON reports(generated_at DESC);
