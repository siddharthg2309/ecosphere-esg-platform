CREATE TABLE environmental_goals (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL CHECK (length(trim(name)) > 0),
  department_id UUID NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
  target_co2 NUMERIC(14,3) NOT NULL CHECK (target_co2 > 0),
  current_co2 NUMERIC(14,3) NOT NULL DEFAULT 0 CHECK (current_co2 >= 0),
  deadline DATE NOT NULL,
  status TEXT NOT NULL DEFAULT 'on_track' CHECK (status IN ('on_track','at_risk','completed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_environmental_goals_dept ON environmental_goals(department_id);
CREATE INDEX idx_environmental_goals_status ON environmental_goals(status, deadline);
