# Phase 5 — Scoring, Reports & Launch

**Goal:** close the loop — the **deterministic scoring engine** and **Overall ESG Score**, the executive dashboard,
the full **reporting** suite (Environmental / Social / Governance / ESG Summary / Custom Builder) with the 6 filters and
**PDF / Excel / CSV** export — then harden, make it responsive, and ship.
**Duration:** ~5 days · **Prerequisites:** Phases 0–4 · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first)
- **Scoring formulas** (below) + recompute triggers; `dept_total` = pillar-weighted **0–100** using `esg_config` weights;
  `overall` = headcount-weighted mean of dept totals (**0–100**). Everything is on the same 0–100 scale.
- **Report matrix**: figures per report type × the 6 filters; **export formats** PDF/Excel/CSV.

---

## P1 — Backend Core (scoring + report composition)

### Scoring engine — pure functions (`modules/scoring/domain`)
```go
// each pillar returns 0..100; formulas are deterministic and unit-tested against fixtures
func Environmental(in EnvInputs) int {
    goalAttainment := clamp(avg(in.GoalProgressPct), 0, 100)          // % of goal targets met
    trend          := clamp(50 + in.YoYReductionPct, 0, 100)          // reduction improves score
    return round(0.6*goalAttainment + 0.4*trend)
}
func Social(in SocialInputs) int {   // participation %, diversity index, training completion %
    return round(0.4*in.CSRParticipationPct + 0.3*in.DiversityIndex + 0.3*in.TrainingCompletionPct)
}
func Governance(in GovInputs) int {  // ack rate, audit pass rate, penalty for open/overdue issues
    base := 0.5*in.PolicyAckPct + 0.5*in.AuditPassPct
    penalty := 3*in.OpenIssues + 5*in.OverdueIssues
    return clamp(round(base) - penalty, 0, 100)
}
type DepartmentScore struct { DeptID id.ID; Env, Social, Gov, Total int; Period string } // every field 0..100

// department total = PILLAR-WEIGHTED score, 0..100; weights from esg_config (default 40/30/30, must sum to 100)
func DeptTotal(env, social, gov, wEnv, wSocial, wGov int) int {
    return round(float64(env*wEnv + social*wSocial + gov*wGov) / float64(wEnv+wSocial+wGov))
}

// organization score = HEADCOUNT-WEIGHTED mean of department totals, 0..100 (simple mean if headcount unknown)
func OverallESG(scores []DepartmentScore, headcount map[id.ID]int) int {
    var num, den int
    for _, s := range scores {
        h := headcount[s.DeptID]; if h == 0 { h = 1 }
        num += s.Total * h; den += h
    }
    if den == 0 { return 0 }
    return round(float64(num) / float64(den))
}
```

### Recompute & report composition (`modules/scoring/app`, `modules/reporting/domain`)
```go
// recompute subscribes to events (or nightly) → writes department_scores read model
bus.Subscribe(EmissionRecorded{}.Name(), recomputeEnvFor(deptID))
bus.Subscribe(ParticipationDecided{}.Name(), recomputeSocialFor(...))
bus.Subscribe(ComplianceIssueRaised{}.Name(), recomputeGovFor(...))   // + Resolved, PolicyAck

// report composition — filters → sections → figures
type ReportType string // environmental|social|governance|esg_summary|custom
type Filters struct { DeptID *id.ID; From,To *time.Time; Module,Employee,Challenge,Category *string }
type Report struct { ID id.ID; Type ReportType; Filters Filters; Sections []Section; GeneratedAt time.Time }
type ComposeReport interface { Execute(ctx, ReportType, Filters) (Report, error) } // app layer persists it to `reports`
```

**Deliverables + tests:** each pillar function tested against fixed fixtures (known inputs → known score);
`OverallESG` matches a hand-computed 40/30/30 example; recompute fires on the right events.

---

## P2 — Backend Adapters (read models + export)

### Migration (`0025`) + queries
```sql
CREATE TABLE department_scores (id UUID PRIMARY KEY, department_id UUID REFERENCES departments(id),
  period TEXT NOT NULL, environmental INT, social INT, governance INT, total INT,   -- all 0..100
  computed_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE(department_id, period));
-- generated reports are persisted so /export can re-read them by id
CREATE TABLE reports (id UUID PRIMARY KEY, type TEXT NOT NULL, filters JSONB NOT NULL,
  result JSONB NOT NULL, generated_by UUID REFERENCES users(id), generated_at TIMESTAMPTZ NOT NULL DEFAULT now());
```
`GET /scores/departments?period=` and `GET /scores/overall?period=` read this table (fast dashboard).

### Export adapters (`modules/reporting/adapter/export`)
```go
type Exporter interface { Export(ctx, Report) ([]byte, string /*mime*/, error) }
// CSV  — encoding/csv (one section = one table block)
// XLSX — xuri/excelize (a sheet per section, header styling)
// PDF  — jung-kurt/gofpdf or maroto (cover + sections + charts as images)
```
| Method | Path | Returns |
| --- | --- | --- |
| POST | `/reports/generate` `{type, filters}` | composes + **persists** to `reports`; returns `Report` (incl. `id`) |
| GET | `/reports/{id}/export?fmt=pdf\|xlsx\|csv` | re-reads report by `id` → `Exporter` → file stream, `Content-Disposition` |
| GET | `/scores/departments` · `/scores/overall` | read model |

**Perf pass:** indexes verified with `EXPLAIN`; dashboard/report reads paginated; heavy aggregations use the read model,
never live-scan transactional tables.

**Deliverables + tests:** all 3 exporters produce openable files; overall score endpoint = engine output; report reflects
each of the 6 filters.

---

## P3 — Frontend (dashboard + reports + responsive polish)

- Wire [`dashboard.html`](../../wireframes/dashboard.html): single stat card ← `/scores/*`, trend + ranking ←
  `/carbon/summary` + `/scores/departments`, activity feed ← recent events.
- [`reports.html`](../../wireframes/reports.html): report cards → `POST /reports/generate` → preview; Custom Builder
  binds the 6 filters; export buttons hit `/export?fmt=`.
```ts
export function useReportBuilderVM(){
  const [f,setF] = useState<Filters>({});
  const gen = useMutation({ mutationFn: () => api.reports.generate({type:f.type, filters:f}) });
  const exportAs = (fmt:'pdf'|'xlsx'|'csv') => download(api.reports.exportUrl(gen.data!.id, fmt));
  return { filters:f, setF, preview:gen.data, generate:gen.mutateAsync, exportAs };
}
```
- **Mobile-responsive pass** across every screen (sidebar → drawer, stat card stacks, tables scroll); a11y sweep
  (focus rings, labels, contrast).

**Deliverables + tests:** dashboard renders live scores; builder VM composes filters into the request; export triggers a
download; responsive breakpoints verified.

---

## P4 — AI narrative · Hardening · Launch

### Report Narrative agent (advisory, labeled)
```go
type Narrator interface { Summarize(ctx, figures ReportFigures) (prose string, err error) }
```
Adds an "Executive summary (AI-generated)" section to the ESG Summary report; deterministic figures unchanged.

### Hardening & launch
- **Caching** for `department_scores`/overall (invalidate on recompute); load-test dashboard + export paths.
- **Security review** (`/security-review`): RBAC on every route, upload safety, JWT/secret handling, dependency scan.
- **CI/CD**: build + migrate + deploy; prod compose/manifests; **migrations runbook**; observability (logs, metrics,
  request IDs); seed prod demo; **smoke tests** (login → dashboard → generate + export a report).

**Deliverables + tests:** narrative section renders (+ offline fallback); security review passes; smoke tests green in
staging.

---

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
