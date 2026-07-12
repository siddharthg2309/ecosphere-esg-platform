CREATE TABLE badges (
  id UUID PRIMARY KEY, name TEXT UNIQUE NOT NULL, description TEXT NOT NULL DEFAULT '', icon TEXT NOT NULL DEFAULT '',
  unlock_rule JSONB NOT NULL CHECK (unlock_rule->>'type' IN ('xp','challenges') AND (unlock_rule->>'value')::int > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
