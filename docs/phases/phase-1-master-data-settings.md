# Phase 1 — Master Data & Settings

**Goal:** deliver every master-data entity and the ESG configuration layer that all transactional modules depend on:
Departments, Categories, Emission Factors, Product ESG Profiles, ESG Policies, Badges, Rewards, and **ESG Config**
(the six-rule toggles + score weightings) + user/employee management.
**Duration:** ~4 days · **Prerequisites:** Phase 0 · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first)
- Full **DB schema** for all master tables + **OpenAPI** for their CRUD.
- **Category.type** enum `{csr_activity, challenge}`; **Badge.unlock_rule** JSON `{type:"xp"|"challenges", value:int}`.
- **esg_config**: 4 booleans + `weight_env/social/gov` (sum = 100); **feature-flag read API** signature.

---

## P1 — Backend Core (master-data domain)

### Domain types (`modules/settings/*`, `modules/environmental/emissionfactor`, `modules/gamification/{badge,reward}`)
```go
type Category struct { ID id.ID; Name string; Type CategoryType; Status Status }
type CategoryType string; const (CatCSR CategoryType="csr_activity"; CatChallenge CategoryType="challenge")

type EmissionFactor struct {
    ID id.ID; Name string; CategoryID id.ID; Unit string
    KgCO2PerUnit decimal.Decimal; Status Status
}
func NewEmissionFactor(name string, cat id.ID, unit string, v decimal.Decimal) (*EmissionFactor, error) {
    if v.LessThanOrEqual(decimal.Zero) { return nil, errs.Invalid("factor_positive","kgco2 must be > 0", …) }
    …
}

type ProductESGProfile struct { ID id.ID; Product string; Attributes json.RawMessage; EmissionFactorID *id.ID }

type Policy struct { ID id.ID; Title, Body string; Version int; EffectiveDate time.Time }
func (p *Policy) Publish(now time.Time) { p.Version++; p.EffectiveDate = now }  // version bump on edit

type UnlockRule struct { Type string `json:"type"`; Value int `json:"value"` }  // "xp" | "challenges"
type Badge struct { ID id.ID; Name, Description, Icon string; Unlock UnlockRule }
func (r UnlockRule) Satisfied(xp, completed int) bool {
    switch r.Type { case "xp": return xp >= r.Value; case "challenges": return completed >= r.Value }
    return false
}

type Reward struct { ID id.ID; Name, Description string; PointsRequired, Stock int; Status Status }
func (r *Reward) CanRedeem() bool { return r.Status==Active && r.Stock > 0 }

// ESG config aggregate — org-level rules
type ESGConfig struct {
    OrgID id.ID
    AutoEmissionCalc, RequireCSREvidence, AutoAwardBadges, NotifyComplianceEmail bool
    WeightEnv, WeightSocial, WeightGov int
}
func (c ESGConfig) Validate() error {
    if c.WeightEnv+c.WeightSocial+c.WeightGov != 100 { return errs.Invalid("weights_sum","must total 100", …) }
    return nil
}
```

### Ports & use-cases
```go
type CategoryRepo interface { Save(ctx,*Category) error; List(ctx, typ *CategoryType, p page.Page) (page.Result[Category],error) }
type ESGConfigRepo interface { Get(ctx, org id.ID) (*ESGConfig, error); Save(ctx,*ESGConfig) error }

type UpdateESGConfig interface { Execute(ctx, ESGConfig) error } // validates → save → publish ESGConfigChanged
```
Emit `events.ESGConfigChanged{ChangedAt}` using the frozen `settings.esg_config_changed` name from [`docs/contracts/events.md`](../contracts/events.md) (invalidates the feature-flag cache in P4).

**Deliverables + tests:** factor rejects ≤0; unlock-rule `Satisfied` table test; weights guard rejects ≠100; policy
version increments on publish.

---

## P2 — Backend Adapters

### Migrations (`0003`–`0012`)

`0011` creates notification preferences and `0012` adds the separately configurable `compliance_overdue` channel. Phase 2 begins at `0013`.
```sql
CREATE TABLE categories (id UUID PRIMARY KEY, name TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('csr_activity','challenge')), status TEXT NOT NULL DEFAULT 'active');
CREATE TABLE emission_factors (id UUID PRIMARY KEY, name TEXT NOT NULL,
  category_id UUID NOT NULL REFERENCES categories(id), unit TEXT NOT NULL,
  kgco2_per_unit NUMERIC(14,4) NOT NULL CHECK (kgco2_per_unit > 0), status TEXT NOT NULL DEFAULT 'active');
CREATE TABLE product_esg_profiles (id UUID PRIMARY KEY, product TEXT NOT NULL,
  attributes JSONB NOT NULL DEFAULT '{}', emission_factor_id UUID REFERENCES emission_factors(id));
CREATE TABLE esg_policies (id UUID PRIMARY KEY, title TEXT NOT NULL, body TEXT NOT NULL,
  version INT NOT NULL DEFAULT 1, effective_date DATE NOT NULL);
CREATE TABLE badges (id UUID PRIMARY KEY, name TEXT NOT NULL, description TEXT, icon TEXT,
  unlock_rule JSONB NOT NULL);
CREATE TABLE rewards (id UUID PRIMARY KEY, name TEXT NOT NULL, description TEXT,
  points_required INT NOT NULL CHECK (points_required >= 0), stock INT NOT NULL CHECK (stock >= 0),
  status TEXT NOT NULL DEFAULT 'active');
CREATE TABLE esg_config (
  org_id UUID PRIMARY KEY,
  auto_emission_calc BOOL NOT NULL DEFAULT true,
  require_csr_evidence BOOL NOT NULL DEFAULT true,
  auto_award_badges BOOL NOT NULL DEFAULT true,
  notify_compliance_email BOOL NOT NULL DEFAULT false,
  weight_env INT NOT NULL DEFAULT 40, weight_social INT NOT NULL DEFAULT 30, weight_gov INT NOT NULL DEFAULT 30,
  CONSTRAINT weights_sum CHECK (weight_env + weight_social + weight_gov = 100));
-- extend users for gamification balances
ALTER TABLE users ADD COLUMN xp INT NOT NULL DEFAULT 0,
  ADD COLUMN points INT NOT NULL DEFAULT 0,
  ADD COLUMN completed_challenges INT NOT NULL DEFAULT 0;  -- fuels badge unlock rules
```

### Endpoints (all admin-write, any-read)
| Method | Path | Notes |
| --- | --- | --- |
| GET/POST/PUT | `/categories` `?type=` | filter by type |
| GET/POST/PUT | `/emission-factors` `?category=` | |
| GET/POST/PUT | `/products` | ESG profiles |
| GET/POST/PUT | `/policies` | POST bumps version |
| GET/POST/PUT | `/badges` | validates unlock_rule JSON |
| GET/POST/PUT | `/rewards` | |
| GET/PUT | `/settings/esg-config` | 400 if weights ≠ 100 |
| GET/POST/PUT | `/employees` | create user, assign dept + role |

All list endpoints paginated (`limit/offset`) and validated; RBAC via `RequireRole(admin)` on writes.

**Deliverables + tests:** category type filter; emission-factor create rejects ≤0 at DB + domain; PUT esg-config with
weights 40/30/40 → 400.

---

## P3 — Frontend (Settings + master data UI)

### Screens (per [`wireframes/settings.html`](../../wireframes/settings.html))
```text
modules/settings/
  departments/ {model,viewmodel,view}
  categories/  {model,viewmodel,view}
  esgConfig/   model/useEsgConfig.ts  viewmodel/useEsgConfigVM.ts  view/EsgConfigTab.tsx
  masterData/  emissionFactors|products|policies|badges|rewards (shared CRUD table + modal pattern)
```

### ESG-config ViewModel (weights guard client-side)
```ts
export function useEsgConfigVM() {
  const q = useEsgConfig();
  const [w,setW] = useState({env:40,social:30,gov:30});
  const sum = w.env+w.social+w.gov;
  const valid = sum===100;                            // disable Save until valid; server re-checks
  const save = useMutation({ mutationFn: api.settings.updateConfig });
  return { toggles:q.data, weights:w, setW, sum, valid, save: save.mutateAsync };
}
```
Reusable `<CrudTable columns rows onCreate onEdit onDelete>` + `<EntityModal fields validate>` drive all six master
entities to keep the UI consistent (design.md primitives only).

**Deliverables + tests:** admin manages all master data + flips toggles; VM unit test: weights 40/30/40 ⇒ `valid=false`,
Save disabled.

---

## P4 — Platform · Config · Data

### Feature-flag service (`platform/settings`)
```go
type Flags interface {
    IsEnabled(ctx context.Context, key string) bool     // "auto_emission_calc" | "require_csr_evidence" | …
    Weights(ctx context.Context) (env, social, gov int)
}
```
Backed by `esg_config`, **cached** (in-memory, 60s TTL); subscribes to `ESGConfigChanged` → invalidate. Later phases
depend only on this interface (never read the table directly).

### User/employee management + RBAC surface
`CreateEmployee(name,email,role,deptID)` issues a temp password (emailed via P0 EmailSender), enforces role enum.

### Seed & fixtures (`cmd/seed` + `testdata/`)
Idempotent seeder: 5 departments, categories (both types), ~8 emission factors, 3 policies, 4 badges, 4 rewards, ~20
users across roles, default `esg_config` (40/30/30, evidence on). Every later phase + demo boots against this.

### OpenRouter model config
`ai.ModelConfig{ Categorize:"anthropic/claude-…", EvidenceAssist:"…", Narrative:"…" }` from env; P1 AI adapter still
returns deterministic fixtures (real calls in Phase 2).

**Deliverables + tests:** `Flags.IsEnabled` reflects a UI toggle change within TTL; seeder is idempotent (run twice = same
rows).

---

## Integration & sync
Freeze schema+OpenAPI on day 1. P3 builds against a mock server until P2 endpoints land (~day 2). End-of-phase: toggles set
in the UI actually change `Flags.IsEnabled(...)`.
**Demo:** admin creates a category + emission factor, adds a badge with an unlock rule, sets weights 40/30/30, toggles
"Require CSR evidence" on.

## Definition of Done
All master entities CRUD + validated + RBAC · esg_config persisted & read via feature flags · weights guard enforced
(domain + DB CHECK + UI) · seed data loads · Settings UI matches wireframe · tests green.
