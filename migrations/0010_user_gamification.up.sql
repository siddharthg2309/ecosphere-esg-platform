ALTER TABLE users ADD COLUMN xp INT NOT NULL DEFAULT 0 CHECK(xp >= 0),
  ADD COLUMN points INT NOT NULL DEFAULT 0 CHECK(points >= 0),
  ADD COLUMN completed_challenges INT NOT NULL DEFAULT 0 CHECK(completed_challenges >= 0),
  ADD COLUMN status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','inactive'));
