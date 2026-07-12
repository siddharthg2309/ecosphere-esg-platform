CREATE TABLE carbon_transactions (
  id UUID PRIMARY KEY,
  department_id UUID NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
  source TEXT NOT NULL CHECK (source IN ('purchase','manufacturing','expense','fleet')),
  quantity NUMERIC(14,3) NOT NULL CHECK (quantity > 0),
  emission_factor_id UUID NOT NULL REFERENCES emission_factors(id) ON DELETE RESTRICT,
  factor_value NUMERIC(14,4) NOT NULL CHECK (factor_value > 0),
  computed_co2 NUMERIC(14,3) NOT NULL DEFAULT 0 CHECK (computed_co2 >= 0),
  txn_date DATE NOT NULL,
  evidence_url TEXT,
  status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','verified')),
  verified_by UUID REFERENCES users(id) ON DELETE RESTRICT,
  verified_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT carbon_verification_state CHECK (
    (status = 'draft' AND verified_by IS NULL AND verified_at IS NULL) OR
    (status = 'verified' AND verified_by IS NOT NULL AND verified_at IS NOT NULL)
  )
);

CREATE INDEX idx_carbon_dept_date ON carbon_transactions(department_id, txn_date);
CREATE INDEX idx_carbon_draft_queue ON carbon_transactions(department_id, created_at) WHERE status = 'draft';
CREATE INDEX idx_carbon_source_date ON carbon_transactions(source, txn_date);
