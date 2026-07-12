CREATE TABLE emission_factors (
  id UUID PRIMARY KEY, name TEXT NOT NULL, category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
  unit TEXT NOT NULL, kgco2_per_unit NUMERIC(14,4) NOT NULL CHECK (kgco2_per_unit > 0),
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_emission_factors_category ON emission_factors(category_id);
