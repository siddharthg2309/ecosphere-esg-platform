# Phase 3 — Social & Gamification

**Goal:** employee-facing engagement — CSR activities + participation (with the evidence gate), diversity & training;
and the full gamification loop — challenge lifecycle, XP, **badge auto-award**, rewards **redemption**, and leaderboards.
**Duration:** ~5 days · **Prerequisites:** Phases 0–2 · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first)
- **Two separate participation models** (CSR vs Challenge) — do not merge.
- **Evidence gate**; **challenge state machine** transitions; **badge unlock** eval + `BadgeUnlocked`;
  **reward redemption** semantics (check stock → deduct points → decrement stock, atomic).
- **Leaderboard** definition (XP by employee & by department, per period).

---

## P1 — Backend Core (social + gamification domain)

### Social (`modules/social/*`)
```go
type CSRActivity struct { ID id.ID; Title string; CategoryID id.ID; Points int; EvidenceRequired bool; Status Status }

type ApprovalStatus string; const (Pending ApprovalStatus="pending"; Approved ApprovalStatus="approved"; Rejected ApprovalStatus="rejected")
type EmployeeParticipation struct {          // CSR ONLY
    ID id.ID; EmployeeID, ActivityID id.ID
    ProofURL string; Approval ApprovalStatus; PointsEarned int; CompletionDate *time.Time
}
// evidence gate — enforced in the use-case; pts = CSRActivity.Points (looked up by the use-case)
func (p *EmployeeParticipation) Approve(pts int, requireEvidence bool, now time.Time) error {
    if requireEvidence && p.ProofURL == "" { return errs.Invalid("evidence_required","proof file required") }
    p.Approval = Approved; p.PointsEarned = pts; p.CompletionDate = &now; return nil
}
```

### Gamification — challenge state machine (`modules/gamification/challenge`)
```go
type ChallengeStatus string
const (SDraft="draft"; SActive="active"; SUnderReview="under_review"; SCompleted="completed"; SArchived="archived")

var transitions = map[ChallengeStatus][]ChallengeStatus{
    SDraft:{SActive, SArchived}, SActive:{SUnderReview, SArchived},
    SUnderReview:{SCompleted, SActive, SArchived}, SCompleted:{SArchived},
}
func (c *Challenge) Transition(to ChallengeStatus) error {
    for _, ok := range transitions[c.Status] { if ok==to { c.Status=to; return nil } }
    return errs.Invalid("bad_transition", fmt.Sprintf("%s → %s not allowed", c.Status, to))
}

type ChallengeParticipation struct {          // CHALLENGE ONLY
    ID id.ID; ChallengeID, EmployeeID id.ID; Progress int; ProofURL string
    Approval ApprovalStatus; XPAwarded int
}
```

### Reward redemption + badge auto-award
```go
// redemption is a use-case that runs inside a DB tx (P2 provides the tx); domain guards:
func (r *Reward) Redeem(emp *User) error {
    if !r.CanRedeem() { return errs.Conflict("out_of_stock","reward unavailable") }
    if emp.Points < r.PointsRequired { return errs.Invalid("insufficient_points","not enough points") }
    emp.Points -= r.PointsRequired; r.Stock--; return nil
}
// badge auto-award subscriber logic (P4 wires it): for each badge, if rule.Satisfied(emp.XP, emp.CompletedChallenges) && !owned → award

// approving a CHALLENGE participation awards XP, bumps the completed-challenge count, and emits BOTH events
func (s *approveChallenge) Execute(ctx, pid, by id.ID) error {
    cp := s.repo.Get(pid); ch := s.chRepo.Get(cp.ChallengeID)
    cp.Approval = Approved; cp.XPAwarded = ch.XP
    s.users.AddXP(cp.EmployeeID, ch.XP); s.users.IncCompleted(cp.EmployeeID)   // completed_challenges++
    if err := s.repo.Save(ctx, cp); err != nil { return err }
    return s.bus.Publish(ctx,
        ParticipationDecided{Kind:"challenge", EmployeeID:cp.EmployeeID, Approved:true, XP:ch.XP},
        ChallengeCompleted{EmployeeID:cp.EmployeeID, ChallengeID:ch.ID, XP:ch.XP})  // → triggers badge auto-award
}
```

### Events
```go
ParticipationDecided{Kind:"csr"|"challenge"; EmployeeID; Approved bool; Points/XP int}
ChallengeCompleted{EmployeeID; ChallengeID; XP int}
BadgeUnlocked{EmployeeID; BadgeID id.ID}
RewardRedeemed{EmployeeID; RewardID id.ID; Points int}
```

**Deliverables + tests:** evidence gate blocks approve w/o proof; illegal transition (`draft→completed`) → invalid;
`Redeem` with insufficient points/zero stock → error; unlock rule boundary tests.

---

## P2 — Backend Adapters

### Migrations (`0013`–`0019`)
```sql
CREATE TABLE csr_activities (id UUID PRIMARY KEY, title TEXT NOT NULL, category_id UUID REFERENCES categories(id),
  points INT NOT NULL DEFAULT 0, evidence_required BOOL NOT NULL DEFAULT true, status TEXT NOT NULL DEFAULT 'active');
CREATE TABLE employee_participations (id UUID PRIMARY KEY, employee_id UUID REFERENCES users(id),
  activity_id UUID REFERENCES csr_activities(id), proof_url TEXT,
  approval TEXT NOT NULL DEFAULT 'pending', points_earned INT NOT NULL DEFAULT 0, completion_date DATE,
  UNIQUE(employee_id, activity_id));
CREATE TABLE challenges (id UUID PRIMARY KEY, title TEXT NOT NULL, category_id UUID REFERENCES categories(id),
  description TEXT, xp INT NOT NULL, difficulty TEXT, evidence_required BOOL NOT NULL DEFAULT true,
  deadline DATE, status TEXT NOT NULL DEFAULT 'draft');
CREATE TABLE challenge_participations (id UUID PRIMARY KEY, challenge_id UUID REFERENCES challenges(id),
  employee_id UUID REFERENCES users(id), progress INT NOT NULL DEFAULT 0, proof_url TEXT,
  approval TEXT NOT NULL DEFAULT 'pending', xp_awarded INT NOT NULL DEFAULT 0, UNIQUE(challenge_id, employee_id));
CREATE TABLE employee_badges (id UUID PRIMARY KEY, employee_id UUID REFERENCES users(id),
  badge_id UUID REFERENCES badges(id), awarded_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE(employee_id, badge_id));
CREATE TABLE reward_redemptions (id UUID PRIMARY KEY, employee_id UUID REFERENCES users(id),
  reward_id UUID REFERENCES rewards(id), points_spent INT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now());
CREATE TABLE trainings (id UUID PRIMARY KEY, name TEXT NOT NULL);
CREATE TABLE training_completions (id UUID PRIMARY KEY, employee_id UUID, training_id UUID, completed_at DATE,
  UNIQUE(employee_id, training_id));
```

### Concurrency-safe redemption (`adapter/postgres` tx)
```sql
BEGIN;
SELECT stock, status FROM rewards WHERE id=$1 FOR UPDATE;              -- row lock
SELECT points FROM users WHERE id=$2 FOR UPDATE;
-- domain guards pass →
UPDATE rewards SET stock = stock - 1 WHERE id=$1 AND stock > 0;        -- guarded
UPDATE users SET points = points - $3 WHERE id=$2;
INSERT INTO reward_redemptions(...);
COMMIT;
```

### Endpoints
| Method | Path | Role |
| --- | --- | --- |
| GET/POST | `/csr/activities` · `/csr/participations` | any / employee |
| POST | `/csr/participations/{id}/approve` `/reject` | dept_head/admin |
| GET/POST | `/challenges` | any / admin+dept_head |
| PUT | `/challenges/{id}/transition` `{to}` | admin/dept_head |
| POST | `/challenges/{id}/participate` · `/challenge-participations/{id}/approve` | employee / dept_head |
| GET | `/leaderboard?scope=employee|department&period=` | any |
| GET/POST | `/rewards` · `POST /rewards/{id}/redeem` | any / employee |
| GET/POST | `/trainings` · `/trainings/{id}/complete` | admin / employee |

### Leaderboard query (window function)
```sql
-- scope=employee
SELECT employee_id, xp, RANK() OVER (ORDER BY xp DESC) rk FROM users WHERE role='employee' ORDER BY xp DESC LIMIT 20;
-- scope=department (aggregate XP by department)
SELECT department_id, SUM(xp) AS xp, RANK() OVER (ORDER BY SUM(xp) DESC) rk
FROM users GROUP BY department_id ORDER BY xp DESC LIMIT 20;
```

**Deliverables + tests:** redemption under concurrent requests never oversells (integration test with parallel calls);
duplicate participation blocked by UNIQUE; leaderboard ranks correct.

---

## P3 — Frontend (Social + Gamification + employee portal)

Build [`social.html`](../../wireframes/social.html) + [`gamification.html`](../../wireframes/gamification.html).
```text
modules/social/{csr,participationQueue,diversity,training}/{model,viewmodel,view}
modules/gamification/{challenges,badges,leaderboard,rewards}/{model,viewmodel,view}
```
- **Approval queue VM**: proof-less rows show a disabled Approve + "No proof" pill (mirrors the evidence gate).
- **Challenge board VM**: groups by status; transition buttons only offer legal `to` states (reads `transitions` map
  shipped as a TS constant, kept in sync with the domain).
- **Redeem VM**: optimistic points decrement + rollback on error; out-of-stock disables the button.

**Deliverables + tests:** VM test — proof-less participation ⇒ Approve disabled; transition VM only shows allowed targets;
redeem rolls back points on server error.

---

## P4 — AI · Rules · Notifications-hooks

### Badge auto-award subscriber (`modules/gamification/app`)
```go
bus.Subscribe(ChallengeCompleted{}.Name(), func(ctx, e Event) error {
    if !flags.IsEnabled(ctx,"auto_award_badges") { return nil }
    ev := e.(ChallengeCompleted); emp := userRepo.Get(ev.EmployeeID)
    for _, b := range badgeRepo.All(ctx) {
        if b.Unlock.Satisfied(emp.XP, emp.CompletedChallenges) && !ownsBadge(emp, b) {
            award(emp, b); bus.Publish(ctx, BadgeUnlocked{emp.ID, b.ID})
        }
    }
    return nil
})
```

### Evidence Assist agent (advisory)
```go
type EvidenceAssist interface { Review(ctx, imgURL string) (looksValid bool, confidence float64, notes string) }
```
Surfaced to the approver as a hint chip; **human still approves**. Falls back silently if AI unavailable.

**Deliverables + tests:** completing a challenge that crosses a threshold awards exactly one badge (idempotent — UNIQUE);
auto-award off ⇒ no award; evidence-assist result attached to the participation payload.

---

## Integration & sync
Freeze the two participation contracts + state machine day 1. P3 uses fixtures for the leaderboard until P2's query lands.
End-of-phase: completing a challenge auto-awards a badge and updates the leaderboard.
**Demo:** employee joins "Commute Green Week" → uploads proof → head approves (+120 XP) → crosses 100 XP →
"Green Beginner" badge auto-awarded → redeems a reward → stock decrements.

## Definition of Done
CSR + challenge participation flows · evidence gate enforced by toggle · lifecycle transitions guarded · badge auto-award
idempotent · leaderboard correct · redemption atomic (no oversell) · diversity + training tracked · events emitted.
