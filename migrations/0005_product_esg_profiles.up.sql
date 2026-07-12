CREATE TABLE product_esg_profiles (
  id UUID PRIMARY KEY, product TEXT UNIQUE NOT NULL, attributes JSONB NOT NULL DEFAULT '{}',
  emission_factor_id UUID REFERENCES emission_factors(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
