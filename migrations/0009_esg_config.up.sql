CREATE TABLE esg_config (
  singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK(singleton),
  auto_emission_calc BOOLEAN NOT NULL DEFAULT TRUE,
  require_csr_evidence BOOLEAN NOT NULL DEFAULT TRUE,
  auto_award_badges BOOLEAN NOT NULL DEFAULT TRUE,
  notify_compliance_email BOOLEAN NOT NULL DEFAULT FALSE,
  weight_env INT NOT NULL DEFAULT 40 CHECK(weight_env BETWEEN 0 AND 100),
  weight_social INT NOT NULL DEFAULT 30 CHECK(weight_social BETWEEN 0 AND 100),
  weight_gov INT NOT NULL DEFAULT 30 CHECK(weight_gov BETWEEN 0 AND 100),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT esg_weights_sum CHECK(weight_env + weight_social + weight_gov = 100)
);
INSERT INTO esg_config(singleton) VALUES(TRUE);
