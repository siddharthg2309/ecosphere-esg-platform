CREATE TABLE csr_activities (
  id UUID PRIMARY KEY,
  title TEXT NOT NULL,
  category_id UUID NOT NULL REFERENCES categories(id),
  description TEXT NOT NULL DEFAULT '',
  points INT NOT NULL DEFAULT 0 CHECK (points >= 0),
  evidence_required BOOL NOT NULL DEFAULT true,
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
  activity_date DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_csr_activities_status ON csr_activities(status);
