# Phase 2 — Environmental & Carbon Engine

**Goal:** the environmental core — carbon transactions with **deterministic** auto-calculation, the
upload → AI-categorize → verify flow, environmental goals with progress tracking, department carbon tracking, and the
environmental dashboard.
**Duration:** ~5 days · **Prerequisites:** Phases 0–1 · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first)
- Shared event names and payloads come from [`docs/contracts/events.md`](../contracts/events.md); do not redefine them locally.
- **Ingest flow**: `POST /carbon/ingest` (file) → `Suggestion{source,categoryId,quantity,unit,confidence}` →
  human edits → `POST /carbon/transactions` (status=`draft`) → `POST /carbon/transactions/{id}/verify`.
- **Invariant**: `computed_co2 = quantity × emission_factor.kgco2_per_unit`; verified txns immutable.
- **Goal status**: `on_track / at_risk / completed`. **`EmissionRecorded`** event payload.
- **AI suggestion JSON schema** (strict) + "advisory only" contract.

---

## P1 — Backend Core (carbon domain)

### Domain (`modules/environmental/carbon`)
```go
type Source string
const (Purchase Source="purchase"; Manufacturing Source="manufacturing"; Expense Source="expense"; Fleet Source="fleet")
type TxnStatus string; const (Draft TxnStatus="draft"; Verified TxnStatus="verified")

type CarbonTransaction struct {
    ID id.ID; DepartmentID id.ID; Source Source
    Quantity decimal.Decimal; EmissionFactorID id.ID; FactorValue decimal.Decimal // snapshot of factor at creation
    ComputedCO2 decimal.Decimal; TxnDate time.Time; EvidenceURL string
    Status TxnStatus; VerifiedBy *id.ID; VerifiedAt *time.Time
}

// deterministic — the ONLY place CO₂ is produced
func (t *CarbonTransaction) compute() { t.ComputedCO2 = t.Quantity.Mul(t.FactorValue).Round(3) }

func (t *CarbonTransaction) Verify(by id.ID, now time.Time) error {
    if t.Status == Verified { return errs.Conflict("already_verified", "cannot re-verify") } // immutability
    t.compute(); t.Status = Verified; t.VerifiedBy = &by; t.VerifiedAt = &now
    return nil
}
```

### Goal domain (`modules/environmental/goal`)
```go
type Goal struct { ID id.ID; Name string; DepartmentID id.ID
    TargetCO2, CurrentCO2 decimal.Decimal; Deadline time.Time; Status GoalStatus }
func (g *Goal) Recompute(now time.Time) {
    switch {
    case g.CurrentCO2.LessThanOrEqual(g.TargetCO2): g.Status = Completed
    case now.After(g.Deadline):                      g.Status = AtRisk
    case g.progress() < g.expectedByNow(now):        g.Status = AtRisk
    default:                                          g.Status = OnTrack
    }
}
```

### Ports & use-cases
```go
type CarbonRepo interface {
    Save(ctx,*CarbonTransaction) error; ByID(ctx,id.ID)(*CarbonTransaction,error)
    EmissionsBySource(ctx, dept *id.ID, from,to time.Time) (map[Source]decimal.Decimal, error)
}
type FactorReader interface { Get(ctx,id.ID) (name string, unit string, v decimal.Decimal, err error) }
type AIGateway interface { Categorize(ctx, DocInput) (Suggestion, error) } // advisory only

// use-cases
type IngestDocument interface { Execute(ctx, IngestCmd) (Suggestion, error) }         // AI → draft suggestion
type RecordEmission  interface { Execute(ctx, RecordCmd) (*CarbonTransaction, error) } // create draft txn
type VerifyTransaction interface { Execute(ctx, id.ID, by id.ID) (*CarbonTransaction, error) }
```
```go
func (s *verify) Execute(ctx, tid, by id.ID) (*CarbonTransaction, error) {
    t, err := s.repo.ByID(ctx, tid); …
    if !s.rbac.IsDeptHeadOf(by, t.DepartmentID) { return nil, errs.Forbidden("not_dept_head", …) }
    if err := t.Verify(by, s.clock.Now()); err != nil { return nil, err }
    if err := s.repo.Save(ctx, t); err != nil { return nil, err }
    _ = s.bus.Publish(ctx, EmissionRecorded{DeptID:t.DepartmentID, Source:t.Source, CO2:t.ComputedCO2, At:t.TxnDate})
    return t, nil
}
```

### Event
```go
type EmissionRecorded = events.EmissionRecorded // DepartmentID, Source, CO2, At
```

**Deliverables + tests:** `compute()` = `qty×factor` exact (decimal); `Verify` twice → `conflict`; non-dept-head verify
→ `forbidden`; goal status transitions table test.

---

## P2 — Backend Adapters

### Migrations (`0013`, `0014`)

These numbers are reserved for Phase 2: `0013_carbon_transactions` and `0014_environmental_goals`.
```sql
CREATE TABLE carbon_transactions (
  id UUID PRIMARY KEY,
  department_id UUID NOT NULL REFERENCES departments(id),
  source TEXT NOT NULL CHECK (source IN ('purchase','manufacturing','expense','fleet')),
  quantity NUMERIC(14,3) NOT NULL,
  emission_factor_id UUID NOT NULL REFERENCES emission_factors(id),
  factor_value NUMERIC(14,4) NOT NULL,
  computed_co2 NUMERIC(14,3) NOT NULL,
  txn_date DATE NOT NULL,
  evidence_url TEXT,
  status TEXT NOT NULL DEFAULT 'draft',
  verified_by UUID REFERENCES users(id), verified_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now());
CREATE INDEX idx_carbon_dept_date ON carbon_transactions(department_id, txn_date);
CREATE TABLE environmental_goals (
  id UUID PRIMARY KEY, name TEXT NOT NULL, department_id UUID NOT NULL REFERENCES departments(id),
  target_co2 NUMERIC(14,3) NOT NULL, current_co2 NUMERIC(14,3) NOT NULL DEFAULT 0,
  deadline DATE NOT NULL, status TEXT NOT NULL DEFAULT 'on_track');
```

### Endpoints
| Method | Path | Role | Notes |
| --- | --- | --- | --- |
| POST | `/carbon/ingest` | employee+ | multipart file → stored + `Suggestion` (AI); gated by `auto_emission_calc` |
| POST | `/carbon/transactions` | employee+ | create `draft` from suggestion (editable) |
| POST | `/carbon/transactions/{id}/verify` | dept_head | deterministic compute + `EmissionRecorded` |
| GET | `/carbon/transactions?dept&from&to&source` | any | paginated |
| GET | `/carbon/summary?dept&from&to` | any | read model: by-source + totals |
| GET/POST/PUT | `/goals` | dept_head/admin | progress + status |

### Read-model query (`EmissionsBySource`)
```sql
SELECT source, SUM(computed_co2) FROM carbon_transactions
WHERE status='verified' AND department_id = $1 AND txn_date BETWEEN $2 AND $3 GROUP BY source;
```

**Deliverables + tests:** verify endpoint sets `computed_co2` server-side (ignores any client CO₂); summary sums only
`verified`; auto-calc off ⇒ ingest returns 409/`disabled`.

---

## P3 — Frontend (Environmental module)

Build [`wireframes/environmental.html`](../../wireframes/environmental.html): tabs, goals table with progress bars,
emissions-by-source card, KPI stat card.

### Upload → verify flow (state machine in the VM)
```ts
// useIngestVM: idle → uploading → suggested(editable) → submitting → draftCreated
export function useIngestVM() {
  const ingest = useMutation({ mutationFn: api.carbon.ingest });      // returns Suggestion
  const record = useMutation({ mutationFn: api.carbon.createTxn });
  // employee edits suggestion fields before record; CO₂ shown as read-only preview = qty × factor
}
// useVerifyQueueVM (dept head): list draft txns → approve → optimistic remove + toast
```
`CO2Preview` component computes `qty × factorValue` client-side **for display only**; the server is authoritative.

**Deliverables + tests:** ingest→edit→record→verify works; VM test: editing quantity updates the CO₂ preview; verify
removes row from the dept-head queue.

---

## P4 — AI · Storage · Platform

### Doc Categorization agent (`platform/ai/openrouter.go`)
```go
type DocInput struct { FileURL, MimeType string; Hint string }
type Suggestion struct { Source string; CategoryID *string; Quantity float64; Unit string; Confidence float64 }

func (a *OpenRouter) Categorize(ctx, in DocInput) (Suggestion, error) {
    body := chatRequest{ Model: a.cfg.Categorize, ResponseFormat: jsonSchema(suggestionSchema),
        Messages: []msg{ system(categorizePrompt), userWithImage(in.FileURL) } }
    // timeout 20s, 2 retries w/ backoff, cost guard (max tokens), reject if !schema-valid
    var s Suggestion; if err := a.call(ctx, body, &s); err != nil { return Suggestion{}, err }
    return s, nil // ADVISORY — never writes CO₂
}
```
- **Structured output** via JSON-schema response format; on low confidence or error → return zero-value → UI falls back
  to **manual entry**. Offline/test mode returns fixtures keyed by filename.

### Storage
```go
func (s *Minio) Put(ctx, key string, r io.Reader, mime string) (string, error) // validates mime∈{jpg,png,pdf}, ≤10MB
func (s *Minio) SignedURL(key string, ttl time.Duration) string
```

**Deliverables + tests:** real OpenRouter call returns schema-valid `Suggestion` (+ fixture mode); upload rejects >10MB /
wrong mime; low-confidence path degrades to manual entry.

---

## Integration & sync
Freeze ingest/verify contract day 1. P4 ships the AI stub immediately so P1/P3 build against it; swap to the live model
mid-phase. End-of-phase the `qty × factor` number is proven deterministic regardless of AI.
**Demo:** employee uploads a fuel invoice → AI suggests Fleet + 268 L diesel → dept head verifies → CO₂ computed →
goal progress + emissions-by-source update.

## Definition of Done
Ingest→AI→verify→transaction works · CO₂ deterministic & immutable post-verify · goals track progress/status · summary
aggregations correct · auto-calc respects the toggle · AI advisory-only · files stored securely.
