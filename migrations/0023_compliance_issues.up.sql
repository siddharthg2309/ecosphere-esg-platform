CREATE TABLE compliance_issues (
  id UUID PRIMARY KEY,
  audit_id UUID REFERENCES audits(id),
  department_id UUID NOT NULL REFERENCES departments(id),
  severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high')),
  description TEXT NOT NULL,
  owner_id UUID NOT NULL REFERENCES users(id),
  due_date DATE NOT NULL,
  status TEXT NOT NULL DEFAULT 'open'
    CHECK (status IN ('open', 'in_progress', 'resolved')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_issue_status_due ON compliance_issues(status, due_date);
CREATE INDEX idx_issue_department ON compliance_issues(department_id);
