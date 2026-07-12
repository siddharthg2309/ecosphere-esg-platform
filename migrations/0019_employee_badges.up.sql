CREATE TABLE employee_badges (
  id UUID PRIMARY KEY,
  employee_id UUID NOT NULL REFERENCES users(id),
  badge_id UUID NOT NULL REFERENCES badges(id),
  awarded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (employee_id, badge_id)
);

CREATE INDEX idx_employee_badges_employee ON employee_badges(employee_id);
