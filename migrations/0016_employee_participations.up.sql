CREATE TABLE employee_participations (
  id UUID PRIMARY KEY,
  employee_id UUID NOT NULL REFERENCES users(id),
  activity_id UUID NOT NULL REFERENCES csr_activities(id),
  proof_url TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  approval TEXT NOT NULL DEFAULT 'pending' CHECK (approval IN ('pending', 'approved', 'rejected')),
  points_earned INT NOT NULL DEFAULT 0 CHECK (points_earned >= 0),
  completion_date DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (employee_id, activity_id)
);

CREATE INDEX idx_employee_participations_activity ON employee_participations(activity_id);
CREATE INDEX idx_employee_participations_approval ON employee_participations(approval);
