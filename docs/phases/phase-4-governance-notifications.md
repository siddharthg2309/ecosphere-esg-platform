# Phase 4 — Governance & Notifications

**Goal:** governance — policies + acknowledgements, audits, and compliance issues with mandatory **owner + due date** and
**overdue flagging** — plus the cross-cutting **notification system** (in-app + email) and the scheduler that powers
reminders and overdue alerts.
**Duration:** ~5 days · **Prerequisites:** Phases 0–3 · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first)
- **Notification catalog** (the 4 required + overdue): `ComplianceIssueRaised, ParticipationDecided, PolicyAckReminder,
  BadgeUnlocked, ComplianceOverdue` — payload shapes.
- **ComplianceIssue** invariant: owner + due date required; **overdue = due_date < today AND status = Open** → flag.
- **Policy acknowledgement** popup trigger (unacknowledged active policy on login).
- **Email**: responsive HTML template set + channel routing from Notification Settings.

## Workstreams (parallel)

### P1 — Backend Core (governance + notification domain)
- `ESGPolicy` (versioned) + `PolicyAcknowledgement`; `Audit`; `ComplianceIssue` with owner+due-date invariant and
  `isOverdue()` logic; guarded status transitions.
- Emit `ComplianceIssueRaised`, `PolicyPublished`, `ComplianceOverdue`.
- `Notification` domain (type, payload, read state) + preference resolution (which channels per type).
- **Deliverables:** governance + notification use-cases, unit-tested (owner/due-date guard, overdue detection).

### P2 — Backend Adapters
- Migrations: `esg_policies`, `policy_acknowledgements`, `audits`, `compliance_issues`, `notifications`.
- Endpoints: policies CRUD + acknowledge, audits CRUD, compliance CRUD + assign owner + transition; notifications
  list/read.
- **Auditor-portal query**: department data bundle (operational records, carbon txns, evidence/invoices, CSR, acks,
  prior issues); compliance filter by status/due.
- **Deliverables:** governance endpoints + auditor data bundle live, RBAC-guarded (auditor scope).

### P3 — Frontend (Governance + auditor portal + notifications UI)
- Governance [`wireframes/governance.html`](../../wireframes/governance.html): policies, acknowledgements, **audits
  table**, **compliance issues** (severity, owner, due date, overdue styling).
- **Auditor portal**: select department → view data bundle → conduct audit → raise issue → assign to department.
- **Policy acknowledgement popup** on login; **notification bell** dropdown + toast + Notification Settings wiring.
- **Deliverables:** governance UI + auditor flow + live in-app notifications.

### P4 — Notifications · Email · Scheduler
- **Event bus subscribers** for the full catalog → write in-app notifications + dispatch email per channel routing.
- **EmailSender**: `html/template` responsive emails (inlined CSS from `design.md` tokens); MailHog (dev) / provider
  (prod).
- **Scheduler**: daily job flags overdue-open compliance issues (`ComplianceOverdue`) + sends policy-ack reminders.
- Retrofit Phase 3 events (`ParticipationDecided`, `BadgeUnlocked`) into email/in-app now.
- **Deliverables:** all 4 notification types + overdue alerts deliver in-app and by email, channel-configurable.

## Integration & sync
Freeze notification catalog day 1 so P1 emits and P4 consumes without churn. P3 renders the in-app store as soon as P2's
notifications endpoint lands. End-of-phase: raising a compliance issue emails the owner and shows a bell notification.
**Demo:** auditor raises a High issue in Logistics (owner + due date) → owner gets email + in-app alert → scheduler
flags it overdue next day → admin sees it re-flagged.

## Definition of Done
Policies/ack/audits/compliance work · owner+due-date enforced · overdue auto-flagged · all 4 notification types deliver
in-app + email · channels honor Notification Settings · auditor portal bundle correct · scheduler running.
