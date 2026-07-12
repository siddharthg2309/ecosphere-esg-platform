# EcoSphere — Master Implementation Plan

> Source of truth for **what** we build and **how**. Requirements: [`notes.md`](notes.md) +
> [`EcoSphere ESG Management Platform.md`](EcoSphere%20ESG%20Management%20Platform.md).
> Visual language: [`design.md`](design.md) + [`wireframes/`](wireframes/index.html).
> Execution is split into 6 phases; each phase runs as **4 parallel workstreams** (see
> [`docs/phases/`](docs/phases/)).

---

## 1. Goal & Vision

Build **EcoSphere**, an ESG Management Platform that integrates Environmental, Social and Governance performance into
day-to-day ERP-style operations. It **measures** sustainability metrics from operational data, **manages** compliance and
policy, and **improves** outcomes through gamified employee participation — surfaced in a unified, role-aware dashboard.

**Product pillars**
- **Environmental** — carbon accounting, emission factors, goals, carbon reports.
- **Social** — CSR activities, employee participation, diversity metrics, training.
- **Governance** — policies, acknowledgements, audits, compliance tracking.
- **Gamification** — challenges, XP, badges, rewards, leaderboards.

**North-star:** a live **Overall ESG Score** (weighted roll-up of department scores) that management can trust, backed by
**deterministic** carbon math and an auditable trail.

---

## 2. Scope

**In scope (spec §6 + §8 — mandatory):** all four modules above; four master-data + nine transactional entities; the
six business rules (auto emission calc, evidence requirement, badge auto-award, reward redemption, compliance ownership,
notifications); reports (Environmental / Social / Governance / ESG Summary / Custom Builder) with 6 filters and
PDF/Excel/CSV export; settings & ESG configuration; role-based portals (employee · operational · auditor · admin).

**In scope (our extensions from `notes.md`):** ESG budget allocation + procurement-proof verification; AI-assisted
document categorization and evidence review (advisory only).

**Out of scope (v1):** full ERP (we ingest operational records, we are not the ERP); native mobile apps (web is
responsive); SSO/SAML (email+password + JWT in v1, pluggable later); multi-currency finance.

**Non-negotiable invariants**
1. Carbon numbers are **deterministic**: `computed_co2 = quantity × emission_factor`. AI never computes the number.
2. Every **Compliance Issue** has an Owner + Due Date; overdue-open issues are flagged → notification.
3. **Evidence Requirement** (when on) blocks CSR approval without a proof file.
4. **Reward redemption** checks stock and deducts points atomically.
5. ESG weightings total 100% and are configurable per org (default 40/30/30).

---

## 3. Tech Stack

| Layer | Choice | Rationale |
| --- | --- | --- |
| Backend | **Go 1.22+**, **Hexagonal (Ports & Adapters)**, modular monolith | Strong typing, fast, simple deploys; hexagon keeps domain pure and testable, modules stay extractable to services later. |
| Database | **PostgreSQL 16** | Relational integrity for ESG data, JSONB for flexible attrs (unlock rules, product profiles), window functions for scoring/leaderboards. |
| Migrations | **golang-migrate** + **sqlc** (typed queries) | Versioned schema; compile-time-safe SQL. |
| Frontend | **React 18 + TypeScript**, **MVVM** | Component UI with a clean View / ViewModel / Model split; testable logic outside components. |
| FE state | **TanStack Query** (server) + Zustand (UI) | Query = Model cache; ViewModels compose both. |
| Build | Vite | Fast dev + typed env. |
| AI | **OpenRouter** (model-routed) via an outbound `AIGateway` port | One API for many models; adapter isolates provider; advisory outputs only. |
| Notifications | Event bus → in-app store + **email** (html/template → SMTP/provider) | Decoupled, testable, 4 required notification types + reminders. |
| Auth | JWT (access + refresh), RBAC middleware | 4 roles → portal views. |
| Reports | Server-side compose → export **gopdf / excelize / encoding/csv** | PDF/Excel/CSV per spec. |
| Infra | Docker Compose (dev), single container + managed Postgres (prod) | Hackathon-friendly, cloud-portable. |

---

## 4. High-Level Design (HLD)

### 4.1 System context

```text
        ┌────────────────────────────────────────────────────────────┐
        │                        Browser (SPA)                        │
        │   React + TS · MVVM · role-aware portals (4)                │
        └───────────────┬─────────────────────────────┬──────────────┘
                        │ HTTPS / REST (JSON, JWT)     │
                        ▼                             ▼
        ┌────────────────────────────────────────────────────────────┐
        │                 EcoSphere API (Go, Hexagonal)               │
        │  ┌──────────── inbound adapters: HTTP router + DTOs ─────┐  │
        │  │  identity · environmental · social · governance ·     │  │
        │  │  gamification · reporting · settings · notification   │  │
        │  └───────────── core: domain + use-cases (ports) ────────┘  │
        │  ┌──────────── outbound adapters ───────────────────────┐  │
        │  │ Postgres repos · OpenRouter AI · Email · Scheduler ·  │  │
        │  │ Object storage (proof files) · Event bus              │  │
        │  └───────────────────────────────────────────────────────┘  │
        └───┬───────────────┬──────────────┬───────────────┬─────────┘
            ▼               ▼              ▼               ▼
       PostgreSQL     OpenRouter      SMTP/Email     Object Storage
       (system of     (AI models,     (notif +       (proofs,
        record)        advisory)       reminders)     invoices)
```

### 4.2 Architecture principles
- **Hexagonal / Ports & Adapters:** the core (domain + application) depends on nothing external; adapters depend on the
  core. Every side-effect (DB, AI, email, storage, clock) is an **outbound port** with a swappable adapter.
- **Modular monolith:** one deployable, internally split by bounded context. No module imports another module's internals
  — only its published port interfaces + domain events.
- **Event-driven side-effects:** state changes emit **domain events**; notifications, badge auto-award, and score
  recomputation subscribe. Keeps use-cases small and rules composable.
- **CQRS-lite:** heavy read screens (dashboard, leaderboard, reports) use dedicated read models / SQL views, separate
  from the write path.

### 4.3 Key data flows

**A. Carbon emission (auto-calc):**
`Employee uploads operational doc → AI categorization (suggest category + parse qty/unit) → Dept Head verifies →
CarbonTransaction created with computed_co2 = qty × factor → EmissionRecorded event → Environmental score recompute.`

**B. CSR / Challenge participation:**
`Employee joins → uploads proof → (Evidence Requirement gate) → Head/Admin approves → points/XP awarded →
ParticipationApproved event → BadgeAutoAward check + Social/Gamification score recompute + notification.`

**C. Governance / compliance:**
`Auditor selects dept → reviews records → conducts Audit → raises ComplianceIssue (owner + due date) →
ComplianceIssueRaised event → notification; Scheduler flags overdue-open issues daily → notification.`

**D. Scoring roll-up:**
`Dept scores (env/social/gov) → Dept Total → Overall ESG = Σ(dept total × weight) → Dashboard & Reports.`

### 4.4 Deployment
- **Dev:** `docker compose up` → api + postgres + mailhog (email capture) + minio (object storage).
- **Prod:** API container (stateless, horizontally scalable) + managed Postgres + managed object store + email provider.
  Scheduler runs as a leader-elected goroutine (or separate cron container).

---

## 5. Low-Level Design (LLD)

### 5.1 Backend hexagon — folder structure

```text
/cmd/api/main.go                      # composition root: wire adapters → core → router
/internal/
  /platform/                          # shared infra (not domain)
    config/  db/  httpserver/  auth/  events/  ai/  email/  storage/  scheduler/  logger/
  /modules/
    /identity/                        # users, roles, auth, RBAC
      domain/ port/ app/ adapter/http/ adapter/postgres/
    /settings/                        # ESG config toggles, weightings, categories, departments
    /environmental/                   # emission factors, products, carbon txns, goals
    /social/                          # CSR activities, employee participation, diversity, training
    /governance/                      # policies, acknowledgements, audits, compliance issues
    /gamification/                    # challenges, participation, badges, XP, rewards, leaderboard
    /scoring/                         # deterministic score engine + read models
    /notification/                    # event subscribers, in-app store, email dispatch
    /reporting/                       # report composition + PDF/Excel/CSV export
/migrations/                          # golang-migrate SQL
/pkg/                                 # reusable, dependency-free helpers
```

**Per-module layering (hexagon):**
- `domain/` — entities, value objects, domain services, domain events. Pure Go, no imports of adapters/framework.
- `port/` — interfaces: **inbound** (use-case contracts the HTTP layer calls) + **outbound** (repositories, gateways).
- `app/` — use-case implementations (application services); orchestrate domain + outbound ports; emit events.
- `adapter/http/` — Gin/chi handlers, request/response DTOs, validation, maps to inbound ports.
- `adapter/postgres/` — sqlc-backed repository implementations of outbound ports.

**Example port (Go-ish):**
```go
// environmental/port/out.go
type CarbonRepo interface {
    Save(ctx, *domain.CarbonTransaction) error
    ByDepartment(ctx, deptID, period) ([]domain.CarbonTransaction, error)
}
type AIGateway interface { // implemented by platform/ai (OpenRouter)
    CategorizeDocument(ctx, DocInput) (Suggestion, error) // advisory only
}
// environmental/port/in.go
type RecordEmission interface {
    Execute(ctx, RecordEmissionCmd) (domain.CarbonTransaction, error)
}
```

### 5.2 Domain model (aggregates)

| Aggregate | Root | Key rules |
| --- | --- | --- |
| Department | Department | code unique; parent forms hierarchy; head is a User |
| Category | Category | type ∈ {csr_activity, challenge}; shared by Social + Gamification |
| EmissionFactor | EmissionFactor | kgco2_per_unit > 0; belongs to a category |
| CarbonTransaction | CarbonTransaction | `computed_co2 = quantity × factor`; immutable once verified |
| EnvironmentalGoal | Goal | target/current CO₂; status derived from progress + deadline |
| CSRActivity | Activity | evidence_required flag |
| EmployeeParticipation (CSR) | Participation | approval blocked without proof when required; points on approval |
| Challenge | Challenge | lifecycle Draft→Active→Under Review→Completed (Archived anytime) |
| ChallengeParticipation | Participation | progress, proof, XP awarded on approval |
| Policy / Acknowledgement | Policy | versioned; ack per employee |
| Audit | Audit | belongs to dept + auditor; findings |
| ComplianceIssue | Issue | **owner + due date required**; overdue-open → flagged |
| Badge | Badge | `unlock_rule` (xp ≥ N or completed ≥ M) |
| Reward / Redemption | Reward | stock ≥ 0; redemption deducts points atomically |
| DepartmentScore | Score | env/social/gov + total; period-stamped read model |

### 5.3 Database schema (core tables)

```text
departments(id, name, code UNIQUE, head_id→users, parent_id→departments, employee_count, status, ...)
users(id, name, email UNIQUE, password_hash, role[employee|dept_head|auditor|admin], department_id, xp, points)
categories(id, name, type[csr_activity|challenge], status)
emission_factors(id, name, category_id, unit, kgco2_per_unit NUMERIC, status)
product_esg_profiles(id, product_name, attributes JSONB, emission_factor_id)
environmental_goals(id, name, department_id, target_co2, current_co2, deadline, status)
carbon_transactions(id, source[purchase|manufacturing|expense|fleet], quantity, emission_factor_id,
                    computed_co2, txn_date, department_id, evidence_url, verified_by, verified_at)
esg_policies(id, title, body, version, effective_date)
policy_acknowledgements(id, employee_id, policy_id, acknowledged_at)   -- UNIQUE(employee_id, policy_id, version)
csr_activities(id, title, category_id, evidence_required BOOL, status)
employee_participations(id, employee_id, activity_id, proof_url, approval_status, points_earned, completion_date)
challenges(id, title, category_id, description, xp, difficulty, evidence_required, deadline, status)
challenge_participations(id, challenge_id, employee_id, progress, proof_url, approval, xp_awarded)
badges(id, name, description, unlock_rule JSONB, icon)
employee_badges(id, employee_id, badge_id, awarded_at)                  -- UNIQUE(employee_id, badge_id)
rewards(id, name, description, points_required, stock, status)
reward_redemptions(id, employee_id, reward_id, points_spent, created_at)
audits(id, title, department_id, auditor_id, audit_date, findings, status)
compliance_issues(id, audit_id, department_id, severity, description, owner_id, due_date, status)
trainings(id, name, ...) / training_completions(id, employee_id, training_id, completed_at)
department_scores(id, department_id, period, environmental, social, governance, total)  -- read model
esg_config(id, org_id, auto_emission_calc, require_csr_evidence, auto_award_badges,
           notify_compliance_email, weight_env, weight_social, weight_gov)               -- weights sum=100
notifications(id, user_id, type, payload JSONB, read_at, created_at)
audit_log(id, actor_id, entity, entity_id, action, diff JSONB, at)      -- change history / traceability
```
Indexes on FKs, `carbon_transactions(department_id, txn_date)`, `compliance_issues(status, due_date)`,
`challenge_participations(employee_id)`. Numeric CO₂ as `NUMERIC(14,3)`.

### 5.4 API surface (REST, representative)

```text
POST /auth/login · POST /auth/refresh · GET /me
# settings
GET/POST/PUT departments · categories · emission-factors · products
GET/PUT /settings/esg-config          # toggles + weightings
# environmental
POST /carbon/transactions             # manual or from verified doc
POST /carbon/ingest                   # upload doc → AI categorize → suggestion
POST /carbon/transactions/{id}/verify # dept head → triggers deterministic calc + event
GET  /goals · PUT /goals/{id}
# social
GET/POST /csr/activities · POST /csr/participations · POST /csr/participations/{id}/approve
GET /diversity · GET/POST /trainings/completions
# governance
GET/POST /policies · POST /policies/{id}/acknowledge
GET/POST /audits · GET/POST /compliance-issues · PUT /compliance-issues/{id}
# gamification
GET/POST /challenges · PUT /challenges/{id}/transition   # lifecycle
POST /challenges/{id}/participate · POST /challenge-participations/{id}/approve
GET /badges · GET /leaderboard · GET/POST /rewards · POST /rewards/{id}/redeem
# scoring + reports
GET /scores/departments · GET /scores/overall
POST /reports/generate  (type, filters[]) → {preview}     ·  GET /reports/{id}/export?fmt=pdf|xlsx|csv
# notifications
GET /notifications · POST /notifications/{id}/read
```
All writes: server-side validation (struct tags + domain guards), RBAC-checked, and audited.

### 5.5 AI agents (OpenRouter) — advisory only

Single outbound port `AIGateway`; OpenRouter adapter routes to a suitable model per task, uses **structured JSON
outputs**, timeouts, retries, and a cost/però-request guard. **No AI output is authoritative for numbers or approvals.**

| Agent | Input | Output (advisory) | Consumer |
| --- | --- | --- | --- |
| **Doc Categorization** | uploaded operational doc/photo | `{source, category_id, quantity, unit, confidence}` | Env: pre-fills a CarbonTransaction for **human verification**; calc stays `qty × factor` |
| **Evidence Assist** | challenge/CSR proof image | `{looks_valid: bool, confidence, notes}` | shows the approver a hint; **human decides** |
| **Report Narrative** | aggregated ESG figures | prose summary for ESG Summary report | Reporting (clearly labeled AI-generated) |
| **Insight/Anomaly** (opt) | trends | flags unusual emissions | Dashboard |

### 5.6 Notification system + email

Event bus (in-process, interface-based). Subscribers:
- **ComplianceIssueRaised** → notify owner + admin (in-app + optional email).
- **ParticipationDecided** (CSR/Challenge approve/reject) → notify employee.
- **PolicyAckDue** (scheduler) → remind employees who haven't acknowledged.
- **BadgeUnlocked** → notify employee.
- **ComplianceOverdue** (scheduler daily) → re-flag + notify.

Email rendered via Go `html/template` into responsive, inlined-CSS HTML (matching `design.md` tokens); sent through the
`EmailSender` port (SMTP in dev via MailHog, provider in prod). All channels toggle via **Notification Settings**.

### 5.7 Scoring engine (deterministic)

`scoring` module recomputes on relevant events (or nightly):
- **Environmental** = f(emissions vs goal targets, trend). **Social** = f(CSR participation %, diversity metrics,
  training completion). **Governance** = f(policy-ack rate, audit outcomes, open/overdue compliance).
- **Dept Total** = env + social + gov. **Overall ESG** = `Σ(dept_total × weight) / Σweight` using `esg_config` weights.
Results stored in `department_scores` (period-stamped) as the read model for dashboard/reports.

### 5.8 Frontend — React + TypeScript, MVVM

```text
/src/
  /app/            router, RoleGuard, AppShell (sidebar + topbar), providers
  /design/         tokens (from design.md/theme.css), primitives: Button, Card, Pill, Tile, Table, Tabs, Switch
  /lib/            apiClient (typed, from OpenAPI), auth, formatters, queryKeys
  /modules/
    /<module>/
      model/       types + query hooks (TanStack Query)         ← Model
      viewmodel/   useXViewModel() — state, actions, derived    ← ViewModel (framework-light, unit-tested)
      view/        presentational components + pages             ← View
```
- **Model:** typed API client + server-cache (TanStack Query). No UI concerns.
- **ViewModel:** a hook per screen (`useDashboardViewModel`) exposing view-ready state + intents; holds validation,
  optimistic updates, derived selectors. Contains the logic; easy to unit-test without rendering.
- **View:** dumb components that render ViewModel output and call its intents. Styled only via design tokens.
- **Portals** = role-filtered navigation over the same module Views (employee sees challenges/CSR; auditor sees
  audits/compliance; etc.).

### 5.9 Cross-cutting
- **AuthZ:** JWT → middleware injects principal; per-route role guard + per-record ownership checks (e.g., dept head
  only their department).
- **Validation:** DTO validation at the edge + invariant checks in the domain (never trust the edge alone).
- **Errors:** typed domain errors → consistent HTTP problem responses.
- **Observability:** structured logs, request IDs, basic metrics; `audit_log` for who-changed-what.
- **Config:** 12-factor env; feature toggles read from `esg_config`.

---

## 6. Non-Functional Requirements

| NFR | Target |
| --- | --- |
| Performance | Dashboard/reads < 300ms p95 via read models; list endpoints paginated |
| Security | RBAC on every route; input validation; hashed passwords; signed JWT; file-type/size checks on uploads |
| Integrity | Money/points & stock changes transactional; CO₂ deterministic; FK constraints |
| Auditability | `audit_log` + immutable verified carbon txns |
| Accessibility | WCAG AA (see `design.md`), keyboard + screen-reader friendly |
| Responsiveness | Mobile-first breakpoints; tables scroll, not the page |
| Testability | Domain unit-tested; use-cases with fake ports; contract tests on adapters |
| Portability | Modules extractable to services; adapters swappable |

---

## 7. Team Model — 4 parallel workstreams

The **same four people** flow through every phase, each owning a consistent slice so they can work in parallel and sync
at contracts (OpenAPI spec, DB migrations, TS types, event definitions).

| # | Persona | Owns |
| --- | --- | --- |
| **P1** | **Backend Core** | Domain models, use-cases, ports, business rules, scoring, events |
| **P2** | **Backend Adapters** | Postgres schema/migrations, sqlc repos, HTTP handlers/DTOs, RBAC middleware |
| **P3** | **Frontend (MVVM)** | Design primitives, module Models/ViewModels/Views, portals, responsiveness |
| **P4** | **AI · Notifications · Platform** | OpenRouter adapter, AI agents, notification/email, scheduler, storage, DevOps/CI |

**Parallelization rule:** each phase begins with a 30-min **contract sync** (freeze the phase's API + DB + event
shapes). Then P1–P4 build against the frozen contracts and integrate at the phase's end-of-phase demo.

---

## 8. Phase Roadmap

| Phase | Theme | Delivers | File |
| --- | --- | --- | --- |
| **0** | Foundation & Walking Skeleton | Repo, hexagon skeleton, DB+migrate, React shell, auth/RBAC, CI, one end-to-end vertical slice | [phase-0](docs/phases/phase-0-foundation.md) |
| **1** | Master Data & Settings | Departments, Categories, Emission Factors, Products, Policies, Badges, Rewards, ESG Config toggles/weights | [phase-1](docs/phases/phase-1-master-data-settings.md) |
| **2** | Environmental & Carbon Engine | Carbon transactions, auto-calc, goals, AI doc categorization, environmental dashboard | [phase-2](docs/phases/phase-2-environmental.md) |
| **3** | Social & Gamification | CSR + participation, diversity, training; challenges lifecycle, badges, XP, rewards, leaderboard | [phase-3](docs/phases/phase-3-social-gamification.md) |
| **4** | Governance & Notifications | Policies/ack, audits, compliance + ownership/overdue; event bus, in-app + email, scheduler | [phase-4](docs/phases/phase-4-governance-notifications.md) |
| **5** | Scoring, Reports & Launch | Scoring engine, dashboards, reports + PDF/Excel/CSV, hardening, mobile polish, deploy | [phase-5](docs/phases/phase-5-reports-scoring-launch.md) |

Phases are mostly sequential (each builds on the prior), but **within a phase all 4 workstreams run in parallel**.
Suggested cadence: Phase 0 ≈ 3 days, Phases 1–5 ≈ 4–5 days each.

---

## 9. Risks & Mitigations

| Risk | Mitigation |
| --- | --- |
| AI over-reach (numbers/approvals) | Hard rule: AI advisory only; deterministic calc + human approval enforced in domain |
| Module coupling creep | Modules talk via ports + events only; lint import boundaries |
| Scoring ambiguity | Freeze scoring formula in Phase 1 contract; store period-stamped read model |
| Parallel contention | Contract-first per phase (OpenAPI + migrations + events frozen before build) |
| Report export scope | Ship ESG Summary + CSV first, then PDF/Excel; Custom Builder last |
| File upload security | Type/size validation, virus-scan hook, signed URLs, no direct exec |

---

## 10. Definition of Done (per feature)
Domain unit tests green · use-case tested with fake ports · endpoint RBAC-guarded + validated · migration reversible ·
ViewModel unit-tested · View matches `design.md` · event emitted + subscriber wired · demoable end-to-end.
