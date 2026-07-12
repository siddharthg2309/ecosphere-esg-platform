# EcoSphere Domain Event Contract

This catalog is the cross-phase source of truth. Event names and payload fields are frozen for Phases 2–5 and are implemented in `internal/platform/events/catalog.go`.

| Event name | Producer | Payload |
| --- | --- | --- |
| `settings.esg_config_changed` | Settings | `changedAt` |
| `environmental.emission_recorded` | Environmental | `departmentId`, `source`, `co2`, `at` |
| `engagement.participation_decided` | Social/Gamification | `kind`, `employeeId`, `approved`, `points`, `xp` |
| `gamification.challenge_completed` | Gamification | `employeeId`, `challengeId`, `xp` |
| `gamification.badge_unlocked` | Gamification | `employeeId`, `badgeId` |
| `gamification.reward_redeemed` | Gamification | `employeeId`, `rewardId`, `points` |
| `governance.policy_published` | Governance | `policyId`, `version` |
| `governance.compliance_issue_raised` | Governance | `issueId`, `ownerId`, `departmentId`, `severity` |
| `governance.compliance_overdue` | Governance scheduler | `issueId`, `ownerId` |

Rules:

- Producers publish only after their database transaction commits.
- Subscribers must be idempotent and use the namespaced event name, never a Go type name or display label.
- Carbon values use decimal serialization; identifiers are UUID strings; timestamps are UTC RFC 3339 values.
- `engagement.participation_decided.kind` is `csr` or `challenge`. Only the relevant `points` or `xp` field is non-zero.
- Notification routing uses five persisted event preferences: compliance raised, approval decision, policy reminder, badge unlock, and compliance overdue.
