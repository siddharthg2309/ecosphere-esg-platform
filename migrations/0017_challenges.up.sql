CREATE TABLE challenges (
  id UUID PRIMARY KEY,
  title TEXT NOT NULL,
  category_id UUID NOT NULL REFERENCES categories(id),
  description TEXT NOT NULL DEFAULT '',
  xp INT NOT NULL DEFAULT 0 CHECK (xp >= 0),
  difficulty TEXT NOT NULL DEFAULT 'medium' CHECK (difficulty IN ('easy', 'medium', 'hard')),
  evidence_required BOOL NOT NULL DEFAULT true,
  deadline DATE,
  status TEXT NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft', 'active', 'under_review', 'completed', 'archived')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_challenges_status ON challenges(status);
