CREATE TABLE policy_acknowledgements (
  id UUID PRIMARY KEY,
  employee_id UUID NOT NULL REFERENCES users(id),
  policy_id UUID NOT NULL REFERENCES esg_policies(id),
  version INT NOT NULL CHECK (version > 0),
  acknowledged_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (employee_id, policy_id, version)
);

CREATE INDEX idx_policy_ack_policy ON policy_acknowledgements(policy_id);
