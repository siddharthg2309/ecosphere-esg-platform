CREATE TABLE reward_redemptions (
  id UUID PRIMARY KEY,
  employee_id UUID NOT NULL REFERENCES users(id),
  reward_id UUID NOT NULL REFERENCES rewards(id),
  points_spent INT NOT NULL CHECK (points_spent >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reward_redemptions_employee ON reward_redemptions(employee_id);
