CREATE TABLE challenge_participations (
  id UUID PRIMARY KEY,
  challenge_id UUID NOT NULL REFERENCES challenges(id),
  employee_id UUID NOT NULL REFERENCES users(id),
  progress INT NOT NULL DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
  proof_url TEXT NOT NULL DEFAULT '',
  approval TEXT NOT NULL DEFAULT 'pending'
    CHECK (approval IN ('pending', 'approved', 'rejected', 'in_progress')),
  xp_awarded INT NOT NULL DEFAULT 0 CHECK (xp_awarded >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (challenge_id, employee_id)
);

CREATE INDEX idx_challenge_participations_employee ON challenge_participations(employee_id);
CREATE INDEX idx_challenge_participations_approval ON challenge_participations(approval);
