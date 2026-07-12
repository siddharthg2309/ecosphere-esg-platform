# Phase 4 — Governance & Notifications

**Goal:** governance — policies + acknowledgements, audits, and compliance issues with mandatory **owner + due date** and
**overdue flagging** — plus the cross-cutting **notification system** (in-app + email) and the scheduler that powers
reminders and overdue alerts.
**Duration:** ~5 days · **Prerequisites:** Phases 0–3 · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first)
- Shared event names and payloads come from [`docs/contracts/events.md`](../contracts/events.md); do not redefine them locally.
- **Notification catalog** (the 4 required + overdue) + payload shapes.
- **ComplianceIssue** invariant: owner + due date required; **overdue = due_date < today AND status = open**.
- **Policy acknowledgement** popup trigger (unacknowledged active policy on login).
- **Email** responsive template set + channel routing from Notification Settings.

---

## P1 — Backend Core (governance + notification domain)

### Governance (`modules/governance/*`)
```go
type Audit struct { ID id.ID; Title string; DepartmentID, AuditorID id.ID; AuditDate time.Time; Findings string; Status AuditStatus }

type Severity string; const (Low="low"; Medium="medium"; High="high")
type IssueStatus string; const (Open="open"; InProgress="in_progress"; Resolved="resolved")
type ComplianceIssue struct {
    ID id.ID; AuditID *id.ID; DepartmentID id.ID; Severity Severity; Description string
    OwnerID id.ID; DueDate time.Time; Status IssueStatus
}
// invariant enforced at construction — owner + due date mandatory
func NewIssue(dept, owner id.ID, sev Severity, desc string, due time.Time) (*ComplianceIssue, error) {
    if owner == "" { return nil, errs.Invalid("owner_required","every issue needs an owner") }
    if due.IsZero() { return nil, errs.Invalid("due_required","every issue needs a due date") }
    return &ComplianceIssue{…, Status:Open}, nil
}
func (i *ComplianceIssue) IsOverdue(now time.Time) bool { return i.Status==Open && now.After(i.DueDate) }

type Policy struct { /* from Phase 1 */ }
type PolicyAcknowledgement struct { ID id.ID; EmployeeID, PolicyID id.ID; Version int; AcknowledgedAt time.Time }
```

### Notification domain (`modules/notification/domain`)
```go
type NotifType string
const (NComplianceRaised="compliance_raised"; NApprovalDecision="approval_decision";
       NPolicyReminder="policy_reminder"; NBadgeUnlocked="badge_unlocked"; NComplianceOverdue="compliance_overdue")
type Notification struct { ID id.ID; UserID id.ID; Type NotifType; Payload json.RawMessage; ReadAt *time.Time; CreatedAt time.Time }

type Channel string; const (InApp="in_app"; Email="email")
type Prefs interface { Channels(t NotifType) []Channel }   // from Notification Settings
```

### Events
```go
type ComplianceIssueRaised = events.ComplianceIssueRaised
type PolicyPublished = events.PolicyPublished
type ComplianceOverdue = events.ComplianceOverdue // emitted by the scheduler
```

**Deliverables + tests:** `NewIssue` rejects missing owner/due; `IsOverdue` boundary (== due date not overdue,
`< today` overdue); ack uniqueness per (employee, policy, version).

---

## P2 — Backend Adapters

### Migrations (`0022`–`0026`)

This range is exclusively reserved for Phase 4. `compliance_overdue` already exists as a persisted notification preference from migration `0012`.
```sql
CREATE TABLE audits (id UUID PRIMARY KEY, title TEXT NOT NULL, department_id UUID REFERENCES departments(id),
  auditor_id UUID REFERENCES users(id), audit_date DATE NOT NULL, findings TEXT, status TEXT NOT NULL DEFAULT 'draft');
CREATE TABLE compliance_issues (id UUID PRIMARY KEY, audit_id UUID REFERENCES audits(id),
  department_id UUID NOT NULL REFERENCES departments(id),
  severity TEXT NOT NULL CHECK (severity IN ('low','medium','high')),
  description TEXT NOT NULL, owner_id UUID NOT NULL REFERENCES users(id), due_date DATE NOT NULL,
  status TEXT NOT NULL DEFAULT 'open');
CREATE INDEX idx_issue_status_due ON compliance_issues(status, due_date);
CREATE TABLE policy_acknowledgements (id UUID PRIMARY KEY, employee_id UUID REFERENCES users(id),
  policy_id UUID REFERENCES esg_policies(id), version INT NOT NULL, acknowledged_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(employee_id, policy_id, version));
CREATE TABLE notifications (id UUID PRIMARY KEY, user_id UUID REFERENCES users(id), type TEXT NOT NULL,
  payload JSONB NOT NULL, read_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
CREATE INDEX idx_notif_user_unread ON notifications(user_id) WHERE read_at IS NULL;
```

### Endpoints
| Method | Path | Role |
| --- | --- | --- |
| GET/POST | `/policies` · `POST /policies/{id}/acknowledge` | admin / employee |
| GET | `/policies/unacknowledged` | any (drives login popup) |
| GET/POST | `/audits` | auditor/admin |
| GET | `/audits/department/{id}/bundle` | auditor (data bundle below) |
| GET/POST/PUT | `/compliance-issues` `?status&overdue` | auditor/dept_head/admin |
| GET | `/notifications` · `POST /notifications/{id}/read` | any |

### Auditor data bundle (single query set)
`GET /audits/department/{id}/bundle` returns `{ operationalRecords, carbonTransactions, evidence, csr, acknowledgements,
priorIssues }` for the selected department — read-only, auditor-scoped RBAC.

**Deliverables + tests:** DB rejects issue without owner/due (NOT NULL); `?overdue=true` filter uses the status+due index;
notifications list returns unread first.

---

## P3 — Frontend (Governance + auditor portal + notifications UI)

Build [`governance.html`](../../wireframes/governance.html) + auditor portal + the notification bell.
```text
modules/governance/{policies,acknowledgements,audits,compliance}/{model,viewmodel,view}
modules/auditor/{selectDept,dataBundle,conductAudit,raiseIssue}/…
app/NotificationBell.tsx  app/PolicyAckModal.tsx  app/Toast.tsx
```
- **PolicyAckModal**: on login, `GET /policies/unacknowledged`; if non-empty, block with an acknowledge modal.
- **NotificationBell**: polls `/notifications` (or SSE later); dropdown lists 5 types with icon + timestamp; unread badge.
- **RaiseIssue form**: owner + due date **required** (client validation mirrors domain) → disabled submit until valid.

**Deliverables + tests:** RaiseIssue VM — submit disabled until owner+due set; unacknowledged policy forces the modal;
bell marks-read updates the badge.

---

## P4 — Notifications · Email · Scheduler

### Event → notification pipeline (`modules/notification/app`)
```go
func wire(bus Bus, store NotifStore, mail EmailSender, prefs Prefs, tmpl *Templates) {
  handle := func(userID id.ID, t NotifType, data any) error {
    for _, ch := range prefs.Channels(t) {
      switch ch {
      case InApp: store.Create(ctx, Notification{UserID:userID, Type:t, Payload:json(data)})
      case Email: mail.Send(ctx, tmpl.Render(t, data))     // html/template, inlined CSS
      }
    }
    return nil
  }
  bus.Subscribe(ComplianceIssueRaised{}.Name(), func(_,e){ h:=e.(...); return handle(h.OwnerID, NComplianceRaised, h) })
  bus.Subscribe(ParticipationDecided{}.Name(),  …NApprovalDecision…)   // retrofit Phase 3 events
  bus.Subscribe(BadgeUnlocked{}.Name(),         …NBadgeUnlocked…)
  bus.Subscribe(ComplianceOverdue{}.Name(),     …NComplianceOverdue…)
}
```

### Email templates (`platform/email/templates/*.html`)
Responsive HTML (max-width 600, inlined CSS from `design.md` tokens): `compliance_raised`, `approval_decision`,
`policy_reminder`, `badge_unlocked`, `compliance_overdue`. `Templates.Render(type, data) → Email{To,Subject,HTML}`.
Dev = MailHog; prod = provider via the `EmailSender` port.

### Scheduler (`platform/scheduler`)
```go
// daily @ 06:00 — overdue compliance
s.Cron("0 6 * * *", func(ctx){
  for _, i := range issueRepo.OpenPastDue(ctx, clock.Now()) { bus.Publish(ctx, ComplianceOverdue{i.ID, i.OwnerID}) }
})
// daily — policy reminders to employees with unacknowledged active policies
s.Cron("0 7 * * *", func(ctx){ for _, u := range userRepo.WithUnacked(ctx) { bus.Publish(ctx, policyReminder(u)) } })
```
Leader-elected (advisory lock) so only one instance fires in multi-replica prod.

**Deliverables + tests:** raising an issue writes an in-app notification + sends an email (captured in MailHog); scheduler
flags an issue whose due date passed; channel routing respects Notification Settings (email off ⇒ in-app only).

---

## Integration & sync
Freeze notification catalog day 1 so P1 emits and P4 consumes without churn. P3 renders the in-app store as soon as P2's
notifications endpoint lands. End-of-phase: raising a compliance issue emails the owner and shows a bell notification.
**Demo:** auditor raises a High issue in Logistics (owner + due date) → owner gets email + in-app alert → scheduler
flags it overdue next day → admin sees it re-flagged.

## Definition of Done
Policies/ack/audits/compliance work · owner+due-date enforced (domain + DB) · overdue auto-flagged by scheduler · all 4
notification types deliver in-app + email · channels honor Notification Settings · auditor bundle correct.
