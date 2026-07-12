CREATE TABLE esg_policies (
  id UUID PRIMARY KEY, title TEXT UNIQUE NOT NULL, body TEXT NOT NULL, version INT NOT NULL DEFAULT 1 CHECK(version > 0),
  effective_date DATE NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
