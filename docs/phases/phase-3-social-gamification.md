# Phase 3 — Social & Gamification

**Goal:** employee-facing engagement — CSR activities + participation (with the evidence gate), diversity & training;
and the full gamification loop — challenge lifecycle, XP, **badge auto-award**, rewards **redemption**, and leaderboards.
**Duration:** ~5 days · **Prerequisites:** Phases 0–2 · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first)
- **Two separate participation models**: `EmployeeParticipation` (CSR) vs `ChallengeParticipation` — do not merge.
- **Evidence gate**: approve blocked without proof when `require_csr_evidence` is on.
- **Challenge state machine**: `Draft→Active→Under Review→Completed` (`Archived` from any state).
- **Badge unlock** evaluation contract + `BadgeUnlocked` event; **Reward redemption** transaction semantics
  (check stock → deduct points → decrement stock, atomic).
- **Leaderboard** definition (XP by employee and by department, per period).

## Workstreams (parallel)

### P1 — Backend Core (social + gamification domain)
- Social: `CSRActivity`, `EmployeeParticipation` (approval awards points; evidence guard), `Training` + completion.
- Gamification: `Challenge` **lifecycle state machine** (guarded transitions), `ChallengeParticipation`
  (XP on approval), `Badge` **unlock-rule evaluator**, `Reward` **redemption** (atomic stock+points).
- Emit `ParticipationDecided`, `ChallengeCompleted`, `BadgeUnlocked`, `RewardRedeemed`.
- **Deliverables:** use-cases + unit tests (evidence gate, illegal transitions, redemption race).

### P2 — Backend Adapters
- Migrations: activities, participations (both), challenges, badges + `employee_badges`, rewards + `reward_redemptions`,
  trainings + completions.
- Endpoints: join/participate, approve/reject, challenge transition, redeem; **leaderboard** (SQL window functions);
  **diversity** aggregation query.
- Redemption endpoint runs in a DB transaction with row locks; unique constraint on `employee_badges`.
- **Deliverables:** all social/gamification endpoints live, concurrency-safe redemption + award.

### P3 — Frontend (Social + Gamification + employee portal)
- Social [`wireframes/social.html`](../../wireframes/social.html): CSR cards + Join, **participation approval queue**
  (Approve/Reject, blocked-without-proof state), diversity dashboard, training completion.
- Gamification [`wireframes/gamification.html`](../../wireframes/gamification.html): lifecycle board, badge gallery
  (locked/unlocked), leaderboard, rewards catalog + **redeem** (stock/points states).
- Employee-portal flows: join challenge → upload proof → track XP/badges → redeem.
- **Deliverables:** complete social + gamification UI with role-appropriate actions.

### P4 — AI · Rules · Notifications-hooks
- **Badge auto-award** subscriber: on `ChallengeCompleted`/XP change, evaluate unlock rules, award (gated by
  `auto_award_badges` toggle), emit `BadgeUnlocked`.
- **Evidence-requirement** enforcement wired to the toggle across CSR approval.
- **Evidence Assist** AI agent (OpenRouter): proof-image → `{looks_valid, confidence, notes}` shown to approver
  (advisory; human approves).
- Emit in-app notification events for approval decisions + badge unlocks (email wiring lands Phase 4).
- **Deliverables:** auto-award works end-to-end; evidence assist visible to approvers.

## Integration & sync
Freeze the two participation contracts + state machine day 1. P3 uses fixtures for leaderboard until P2's query lands.
End-of-phase: completing a challenge auto-awards a badge and updates the leaderboard.
**Demo:** employee joins "Commute Green Week" → uploads proof → head approves (+120 XP) → crosses 100 XP →
"Green Beginner" badge auto-awarded → redeems a reward → stock decrements.

## Definition of Done
CSR + challenge participation flows work · evidence gate enforced by toggle · lifecycle transitions guarded ·
badge auto-award + leaderboard correct · redemption atomic (stock+points) · diversity + training tracked · events emitted.
