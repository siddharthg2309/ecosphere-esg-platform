# Phase 2 — Environmental & Carbon Engine

**Goal:** the environmental core — carbon transactions with **deterministic** auto-calculation, the
upload → AI-categorize → verify flow, environmental goals with progress tracking, department carbon tracking, and the
environmental dashboard.
**Duration:** ~5 days · **Prerequisites:** Phases 0–1 · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first)
- **Ingest flow**: `POST /carbon/ingest` (file) → AI `Suggestion{source,category_id,quantity,unit,confidence}` →
  human edits → `POST /carbon/transactions` → `POST /carbon/transactions/{id}/verify`.
- **Invariant**: `computed_co2 = quantity × emission_factor.kgco2_per_unit`; verified txns immutable.
- **Goal status** rules: `On Track / At Risk / Completed` from progress vs deadline.
- **`EmissionRecorded`** event payload (dept, source, co2, period) consumed by scoring later.
- **AI suggestion JSON schema** (strict) + "advisory only" contract.

## Workstreams (parallel)

### P1 — Backend Core (carbon domain)
- `CarbonTransaction` aggregate: compute + immutability-after-verify; `source ∈ {purchase,manufacturing,expense,fleet}`.
- Use-cases: `IngestDocument` (calls AIGateway, returns draft), `RecordEmission`, `VerifyTransaction` (dept-head only →
  emits `EmissionRecorded`).
- `EnvironmentalGoal` domain: `current_co2` update from transactions, status derivation.
- **Deliverables:** carbon + goal use-cases, unit-tested incl. the `qty × factor` math and immutability guard.

### P2 — Backend Adapters
- Migrations: `carbon_transactions`, `environmental_goals` (+ indexes on `(department_id, txn_date)`).
- Repos + endpoints: ingest, transactions CRUD, verify, goals CRUD.
- **Read models / queries**: emissions by source, by department, by period; goal progress rollup — feed the dashboard.
- Gate auto-calc behind `SettingsService.IsEnabled("auto_emission_calc")`.
- **Deliverables:** environmental endpoints live; aggregation queries performant + paginated.

### P3 — Frontend (Environmental module)
- Build [`wireframes/environmental.html`](../../wireframes/environmental.html): tabs (Emission Factors · Product ESG
  Profiles · Carbon Transactions · Goals), goals table with progress bars, emissions-by-source cards, KPI tiles.
- **Upload → verify UI**: employee uploads a doc → sees AI suggestion (editable) → dept head verify screen.
- Environmental dashboard widgets (emissions trend, totals). Model/ViewModel per view.
- **Deliverables:** full environmental UI incl. the human-in-the-loop verify flow.

### P4 — AI · Storage · Platform
- **Doc Categorization agent** on the OpenRouter adapter: structured JSON output, confidence, timeout/retry, cost guard;
  returns advisory suggestion only (never writes CO₂).
- **Object storage** for uploaded docs/proofs (signed URLs, type/size validation).
- Wire `AIGateway.CategorizeDocument` into `IngestDocument`; provide deterministic test fixtures + a "manual entry"
  fallback when AI is off/low-confidence.
- **Deliverables:** working ingest with real OpenRouter call (and offline fixture mode) + secure file storage.

## Integration & sync
Freeze ingest/verify contract day 1. P4 delivers AI stub immediately so P1/P3 build against it; swap to live model
mid-phase. End-of-phase the `qty × factor` number is proven deterministic regardless of AI.
**Demo:** employee uploads a fuel invoice → AI suggests Fleet + 268 L diesel → dept head verifies → CO₂ computed →
goal progress + emissions-by-source update.

## Definition of Done
Ingest→AI→verify→transaction works · CO₂ deterministic & immutable post-verify · goals track progress/status ·
dashboard aggregations correct · auto-calc respects the toggle · AI is advisory-only · files stored securely.
