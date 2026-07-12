# Phase 5 — Scoring, Reports & Launch

**Goal:** close the loop — the **deterministic scoring engine** and **Overall ESG Score**, the executive dashboard,
the full **reporting** suite (Environmental / Social / Governance / ESG Summary / Custom Builder) with the 6 filters and
**PDF / Excel / CSV** export — then harden, make it responsive, and ship.
**Duration:** ~5 days · **Prerequisites:** Phases 0–4 · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first)
- **Scoring formulas**: Environmental (emissions vs goals), Social (participation/diversity/training), Governance
  (ack rate/audits/compliance); `dept_total = env+social+gov`; `overall = Σ(dept_total × weight)/Σweight`.
- **Recompute triggers**: `EmissionRecorded, ParticipationDecided, ComplianceIssueRaised/Resolved, PolicyAck` (or nightly).
- **Report matrix**: which figures per report type × the 6 filters (Department, Date Range, Module, Employee, Challenge,
  ESG Category); **export formats** PDF/Excel/CSV.

## Workstreams (parallel)

### P1 — Backend Core (scoring + report composition)
- **Scoring engine**: pure functions for env/social/gov, dept total, and the **weighted Overall ESG** (weights from
  `esg_config`); recompute on events → write `department_scores` (period-stamped).
- **Report composition** domain: assemble a report model from filters (type → sections → figures).
- **Deliverables:** scoring + report-composition use-cases, unit-tested against fixed fixtures.

### P2 — Backend Adapters (read models + export)
- `department_scores` read model + queries; dashboard aggregation endpoints (`/scores/departments`, `/scores/overall`).
- `POST /reports/generate` (type + filters → preview) and `GET /reports/{id}/export?fmt=pdf|xlsx|csv` using
  **encoding/csv**, **excelize**, **gopdf**.
- Query performance pass: indexes, pagination, EXPLAIN on heavy reads.
- **Deliverables:** dashboard + report endpoints + working exports in all 3 formats.

### P3 — Frontend (dashboard + reports + responsive polish)
- Wire the **executive dashboard** [`wireframes/dashboard.html`](../../wireframes/dashboard.html) to real scores, trend,
  ranking, activity, quick actions.
- **Reports** [`wireframes/reports.html`](../../wireframes/reports.html): report cards, **Custom Builder** (6 filters),
  preview, PDF/Excel/CSV export buttons.
- **Mobile-responsive pass** across every screen (sidebar drawer, stacked KPIs, scrollable tables); a11y sweep.
- **Deliverables:** live dashboard + reports + responsive, accessible UI.

### P4 — AI narrative · Hardening · Launch
- **Report Narrative** AI agent (OpenRouter): prose ESG-summary section (clearly labeled AI-generated).
- **Hardening**: caching for read models, load-test critical paths, error budgets.
- **Security review**: RBAC/authz audit, upload safety, secrets, dependency scan (`/security-review`).
- **Launch**: CI/CD, prod compose/manifests, migrations runbook, observability dashboards, seed prod demo, smoke tests.
- **Deliverables:** production deploy + runbook + green smoke tests.

## Integration & sync
Freeze scoring formula + report matrix day 1 (P1 owns the numbers, P2 the delivery). P3 builds dashboard/reports against
fixtures until endpoints land. End-of-phase = full end-to-end demo across all modules.
**Demo:** an approved challenge + a verified emission + a resolved compliance issue move the department and **Overall ESG**
scores on the dashboard → generate a filtered **ESG Summary** with AI narrative → export to PDF.

## Definition of Done
Scores deterministic & recompute on events · Overall ESG uses configurable weights · dashboard live · all 5 report types
+ 6 filters + PDF/Excel/CSV export · responsive + WCAG AA · security review passed · deployed with runbook + smoke tests.

---

## Post-launch backlog (v2, optional)
Dark mode · SSO/SAML · native mobile · deeper anomaly detection · ESG budget optimization suggestions ·
public ESG disclosure export (GRI/BRSR mapping).
