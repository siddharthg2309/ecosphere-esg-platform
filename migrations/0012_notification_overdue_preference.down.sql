DELETE FROM notification_preferences WHERE event_type = 'compliance_overdue';

ALTER TABLE notification_preferences
  DROP CONSTRAINT notification_preferences_event_type_check;

ALTER TABLE notification_preferences
  ADD CONSTRAINT notification_preferences_event_type_check
  CHECK(event_type IN (
    'compliance_raised',
    'approval_decision',
    'policy_reminder',
    'badge_unlock'
  ));
