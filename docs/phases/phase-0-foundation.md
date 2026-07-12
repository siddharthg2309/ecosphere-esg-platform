# Phase 0 — Foundation & Walking Skeleton

**Goal:** stand up the hexagonal Go backend, Postgres + migrations, the React/MVVM shell, auth/RBAC, CI, and prove the
whole stack with **one vertical slice** (Departments CRUD) that flows through every layer. After this phase, every later
feature is "fill in the template."
**Duration:** ~3 days · **Prerequisites:** none · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first — 30 min)
- Repo + **module template** (`domain / port / app / adapter/http / adapter/postgres`).
- **Auth token** shape (JWT claims: `sub, role, dept_id, exp`) and refresh flow.
- **Error envelope**: `{ code, message, fields? }` + HTTP status mapping.
- **DB conventions**: snake_case, `id` UUID, `created_at/updated_at`, FK naming.
- **OpenAPI baseline** file that P2/P3 share; **event bus** interface signature.

## Workstreams (parallel)

### P1 — Backend Core
- Define hexagon conventions + shared kernel (`pkg/`: ids, clock, errors, pagination).
- `events` port (Publish/Subscribe interface) + in-process implementation contract.
- **identity domain**: `User`, `Role` (employee|dept_head|auditor|admin) value object, invariants.
- **settings/department domain** + use-cases `CreateDepartment`, `ListDepartments`, `AssignHead` (the vertical slice).
- Inbound + outbound port interfaces for identity + department.
- **Deliverables:** `internal/modules/identity/*`, `internal/modules/settings/department/*`, module README template.

### P2 — Backend Adapters
- Repo scaffolding, `cmd/api/main.go` composition root, config wiring.
- Postgres pool (`platform/db`), **golang-migrate** setup + first migrations: `users`, `departments`.
- `sqlc` config; implement `UserRepo`, `DepartmentRepo`.
- HTTP server (chi/gin), router, middleware chain, DTO + request validation, `Department` + `auth/login` handlers.
- RBAC middleware skeleton (reads JWT → principal → role guard).
- **Deliverables:** running API with `POST/GET /departments`, `POST /auth/login`; migrations up/down clean.

### P3 — Frontend (MVVM)
- Vite + React + TS scaffold; import design tokens from [`design.md`](../../design.md)/`theme.css` into a `design/` layer.
- **Primitives**: Button, Card, Pill, Tile, Table, Tabs, Switch, Input (match wireframes).
- **AppShell**: sidebar + topbar (from `wireframes/`), router, `RoleGuard`, 404/empty states.
- Typed `apiClient` (generated from OpenAPI), auth store, login page.
- Vertical slice: **Departments** list + create (Model query hook + `useDepartmentsViewModel` + View).
- **Deliverables:** SPA that logs in and does Departments CRUD against the API.

### P4 — Platform · AI · Notifications · DevOps
- **Docker Compose**: api + postgres + **mailhog** + **minio**; `.env.example`; Makefile targets.
- `platform/config` (12-factor), `platform/logger` (structured), request-ID middleware.
- **Auth service**: JWT issue/verify, bcrypt/argon2 password hashing.
- Outbound port **stubs + adapters**: `AIGateway` → OpenRouter client skeleton; `EmailSender` → MailHog; `Storage` → minio.
- **CI**: lint (golangci-lint, eslint), test, build, migrations-check.
- **Deliverables:** `docker compose up` yields full local stack; CI green on PR.

## Integration & sync
Mid-phase: P2+P3 align on OpenAPI for `/departments`. End-of-phase: wire P1 use-cases into P2 handlers, P3 hits real API,
P4 confirms auth + compose. **Demo:** log in as admin → create a department → see it listed.

## Definition of Done
Compose stack boots · login issues JWT · Departments CRUD works browser→DB · migrations reversible · CI green ·
module template documented so Phase 1 can parallelize cleanly.
