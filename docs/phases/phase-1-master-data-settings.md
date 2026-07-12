# Phase 1 — Master Data & Settings

**Goal:** deliver every master-data entity and the ESG configuration layer that all transactional modules depend on:
Departments, Categories, Emission Factors, Product ESG Profiles, ESG Policies, Badges, Rewards, and **ESG Config**
(the six-rule toggles + score weightings) + user/employee management.
**Duration:** ~4 days · **Prerequisites:** Phase 0 · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first)
- Full **DB schema** for all master tables (see `plan.md` §5.3) + **OpenAPI** for their CRUD.
- **Category.type** enum `{csr_activity, challenge}`; **Badge.unlock_rule** JSON shape `{type: xp|challenges, value}`.
- **esg_config** shape: 4 boolean toggles + `weight_env/social/gov` (must sum to 100).
- **Feature-flag read API** (`SettingsService.IsEnabled(key)`) that later phases consume.

## Workstreams (parallel)

### P1 — Backend Core (master-data domain)
- Domain + use-cases for **Category, EmissionFactor, ProductESGProfile, ESGPolicy (versioned), Badge, Reward**.
- Invariants: `kgco2_per_unit > 0`, reward `stock ≥ 0`, badge `unlock_rule` valid, policy version bump on edit.
- `ESGConfig` domain: toggle set + **weights-sum-100 guard**; emit `ESGConfigChanged`.
- **Deliverables:** use-cases + unit tests for all master aggregates.

### P2 — Backend Adapters
- Migrations + sqlc repos + HTTP CRUD for all master entities; list endpoints **paginated + filterable**.
- `GET/PUT /settings/esg-config`; category filter by type; emission-factor filter by category.
- Extend RBAC: only **admin** writes master data; all roles read where relevant.
- **Deliverables:** all master endpoints live, validated, RBAC-guarded; OpenAPI in sync.

### P3 — Frontend (Settings + master data UI)
- Build the **Settings** screens per [`wireframes/settings.html`](../../wireframes/settings.html): Departments table,
  Categories, **ESG Configuration** toggles, weighting editor, Notification Settings.
- Master-data management views (Emission Factors, Products, Policies, Badges, Rewards) with create/edit modals + inline
  validation; Model hooks + ViewModels per entity.
- Weighting editor enforces sum=100 client-side (server re-validates).
- **Deliverables:** admin can manage all master data + flip ESG toggles from the UI.

### P4 — Platform · Config · Data
- **SettingsService / feature-flag** reader backed by `esg_config` (cached, invalidated on `ESGConfigChanged`).
- **User/employee management**: create employees, assign to department, assign roles (RBAC surface).
- **Seed & fixtures**: demo org (departments, categories, factors, policies, badges, rewards, sample users) for
  every later phase + demos.
- OpenRouter **model config** (which model per agent) surfaced via config; AI stub returns deterministic fixtures.
- **Deliverables:** feature flags usable by other modules; seeded demo data; user admin working.

## Integration & sync
Freeze schema+OpenAPI on day 1. P3 builds against mock server until P2 endpoints land (~day 2). End-of-phase: toggles set
in UI actually change `SettingsService.IsEnabled(...)`.
**Demo:** admin creates a category + emission factor, adds a badge with an unlock rule, sets weights 40/30/30, toggles
"Require CSR evidence" on.

## Definition of Done
All master entities CRUD + validated + RBAC · esg_config persisted & read via feature flags · weights guard enforced ·
seed data loads · Settings UI matches wireframe · tests green.
