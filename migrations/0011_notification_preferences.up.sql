CREATE TABLE notification_preferences (
  event_type TEXT PRIMARY KEY CHECK(event_type IN ('compliance_raised','approval_decision','policy_reminder','badge_unlock')),
  in_app_enabled BOOLEAN NOT NULL DEFAULT TRUE, email_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO notification_preferences(event_type,in_app_enabled,email_enabled) VALUES
  ('compliance_raised',TRUE,TRUE),('approval_decision',TRUE,FALSE),('policy_reminder',TRUE,TRUE),('badge_unlock',TRUE,FALSE);
