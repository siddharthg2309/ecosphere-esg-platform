ALTER TABLE notification_preferences
  DROP CONSTRAINT notification_preferences_event_type_check;

ALTER TABLE notification_preferences
  ADD CONSTRAINT notification_preferences_event_type_check
  CHECK(event_type IN (
    'compliance_raised',
    'approval_decision',
    'policy_reminder',
    'badge_unlock',
    'compliance_overdue'
  ));

INSERT INTO notification_preferences(event_type, in_app_enabled, email_enabled)
VALUES ('compliance_overdue', TRUE, TRUE)
ON CONFLICT(event_type) DO NOTHING;
