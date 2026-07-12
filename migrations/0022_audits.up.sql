CREATE TABLE audits (
  id UUID PRIMARY KEY,
  title TEXT NOT NULL,
  department_id UUID NOT NULL REFERENCES departments(id),
  auditor_id UUID NOT NULL REFERENCES users(id),
  audit_date DATE NOT NULL,
  findings TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft', 'under_review', 'completed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audits_department ON audits(department_id);
