# Phase 0 — Foundation & Walking Skeleton

**Goal:** stand up the hexagonal Go backend, Postgres + migrations, the React/MVVM shell, auth/RBAC, CI, and prove the
whole stack with **one vertical slice** (Departments CRUD) that flows through every layer. After this phase, every later
feature is "fill in the template."
**Duration:** ~3 days · **Prerequisites:** none · **Parent:** [`plan.md`](../../plan.md)

## Contract freeze (agree first — 30 min)
- Repo + **module template** (`domain / port / app / adapter/http / adapter/postgres`).
- **Auth token** shape (JWT claims: `sub, role, dept_id, exp`) and refresh flow.
- **Error envelope**: `{ code, message, fields? }` + HTTP status mapping.
- **DB conventions**: snake_case, `id UUID`, `created_at/updated_at timestamptz`, FK `<entity>_id`.
- **OpenAPI baseline** file that P2/P3 share; **event bus** interface signature.

---

## Repo layout (created this phase)

```text
ecosphere/
  cmd/api/main.go                     # composition root
  internal/
    platform/
      config/config.go                # env → typed Config
      db/pool.go  db/tx.go            # pgxpool + tx helper
      httpserver/server.go router.go  # chi mux + middleware chain
      auth/jwt.go  auth/middleware.go # issue/verify + RBAC guard
      events/bus.go                   # in-process pub/sub
      ai/openrouter.go                # AIGateway adapter (stub in P0)
      email/sender.go                 # EmailSender adapter (MailHog)
      storage/store.go                # object storage (minio)
      logger/logger.go                # slog wrapper + request-id
    modules/
      identity/  {domain,port,app,adapter/http,adapter/postgres}
      settings/department/  {domain,port,app,adapter/http,adapter/postgres}
  migrations/0001_users.up.sql ...    # golang-migrate
  pkg/{id,errs,clock,page}/           # dependency-free kernel
  api/openapi.yaml                    # shared contract
  docker-compose.yml  Makefile  .env.example
  web/                                # React app (P3)
```

---

## P1 — Backend Core

### Shared kernel (`pkg/`)
```go
// pkg/id
type ID string
func New() ID { return ID(uuid.Must(uuid.NewV7()).String()) } // time-ordered

// pkg/errs — typed domain errors, mapped to HTTP by the adapter
type Kind string
const (KindInvalid Kind="invalid"; KindNotFound Kind="not_found";
       KindConflict Kind="conflict"; KindForbidden Kind="forbidden")
type Error struct { Kind Kind; Code, Message string; Fields map[string]string }
func (e *Error) Error() string { return e.Message }
func Invalid(code,msg string, f map[string]string) *Error { … }
func NotFound(code,msg string) *Error { … }

// pkg/clock
type Clock interface { Now() time.Time }

// pkg/page
type Page struct { Limit, Offset int }               // Limit default 20, max 100
type Result[T any] struct { Items []T; Total int }
```

### Event bus port (`internal/platform/events`)
```go
type Event interface { Name() string }
type Handler func(ctx context.Context, e Event) error
type Bus interface {
    Publish(ctx context.Context, events ...Event) error
    Subscribe(name string, h Handler)                // name = Event.Name()
}
```
P0 ships an in-process synchronous `Bus` (dispatch after the DB tx commits). Later phases only add events + handlers.

### identity domain (`modules/identity/domain`)
```go
type Role string
const (RoleEmployee Role="employee"; RoleDeptHead Role="dept_head";
       RoleAuditor Role="auditor"; RoleAdmin Role="admin")
func (r Role) Valid() bool { … }

type User struct {
    ID           id.ID
    Name, Email  string
    PasswordHash string
    Role         Role
    DepartmentID *id.ID
    CreatedAt    time.Time
}
func NewUser(name, email string, role Role, dept *id.ID) (*User, error) // validates email + role
```

### department domain + use-cases (`modules/settings/department`)
```go
// domain
type Status string; const (Active Status="active"; Inactive Status="inactive")
type Department struct {
    ID id.ID; Name, Code string; HeadID, ParentID *id.ID
    EmployeeCount int; Status Status; CreatedAt time.Time
}
func New(name, code string) (*Department, error)   // code non-empty, uppercased, unique enforced by repo
func (d *Department) AssignHead(u id.ID) { d.HeadID=&u }

// port/out.go
type Repo interface {
    Save(ctx context.Context, d *Department) error
    ByID(ctx context.Context, id id.ID) (*Department, error)
    List(ctx context.Context, p page.Page) (page.Result[Department], error)
    ExistsCode(ctx context.Context, code string) (bool, error)
}
// port/in.go — use-case contracts the HTTP layer calls
type CreateDepartment interface{ Execute(ctx, CreateDepartmentCmd) (*Department, error) }
type ListDepartments  interface{ Execute(ctx, page.Page) (page.Result[Department], error) }

// app/create_department.go
type CreateDepartmentCmd struct { Name, Code string; HeadID *id.ID }
func (s *createDept) Execute(ctx, c CreateDepartmentCmd) (*Department, error) {
    if ok,_ := s.repo.ExistsCode(ctx, c.Code); ok { return nil, errs.Conflict("code_taken", …) }
    d, err := domain.New(c.Name, c.Code); if err != nil { return nil, err }
    if c.HeadID != nil { d.AssignHead(*c.HeadID) }
    return d, s.repo.Save(ctx, d)
}
```

**Deliverables:** `pkg/*`, event bus, identity + department domain/ports/use-cases, module README template.
**Tests:** `New()` rejects empty code; `CreateDepartment` returns `conflict` on duplicate code; role validation table test.

---

## P2 — Backend Adapters

### Composition root (`cmd/api/main.go`)
```go
cfg := config.Load()
pool := db.New(cfg.DatabaseURL)
bus  := events.NewInProcess()
// wire outbound adapters → use-cases → handlers
deptRepo := pg.NewDepartmentRepo(pool)
createDept := app.NewCreateDepartment(deptRepo)
r := httpserver.NewRouter(cfg, authMW, deptHandler(createDept, listDept), authHandler)
httpserver.Run(cfg.Addr, r)
```

### Config
```go
type Config struct {
    Addr string `env:"ADDR" default:":8080"`
    DatabaseURL string `env:"DATABASE_URL"`
    JWTSecret string `env:"JWT_SECRET"`
    AccessTTL time.Duration `env:"ACCESS_TTL" default:"15m"`
}
```

### Migrations (`migrations/`)
```sql
-- 0001_users.up.sql
CREATE EXTENSION IF NOT EXISTS citext;                  -- case-insensitive email
CREATE TABLE users (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  email CITEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('employee','dept_head','auditor','admin')),
  department_id UUID,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE refresh_tokens (                            -- hashed, rotated on use
  id UUID PRIMARY KEY, user_id UUID NOT NULL REFERENCES users(id),
  token_hash TEXT NOT NULL, expires_at TIMESTAMPTZ NOT NULL, revoked_at TIMESTAMPTZ
);
-- 0002_departments.up.sql
CREATE TABLE departments (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  code TEXT UNIQUE NOT NULL,
  head_id UUID REFERENCES users(id),
  parent_id UUID REFERENCES departments(id),
  employee_count INT NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE users ADD CONSTRAINT fk_user_dept FOREIGN KEY (department_id) REFERENCES departments(id);
```

### Repository (sqlc-backed, `adapter/postgres`)
```go
type DepartmentRepo struct { q *sqlc.Queries }
func (r *DepartmentRepo) Save(ctx, d *domain.Department) error {
    return r.q.UpsertDepartment(ctx, sqlc.UpsertDepartmentParams{ ID:d.ID, Name:d.Name, Code:d.Code, … })
}
```

### HTTP layer (`adapter/http` + `platform/httpserver`)
- **Middleware chain:** `Recover → RequestID → Logger → Auth(JWT) → RBAC(role) → handler`.
- **DTO + validation** (go-playground/validator):
```go
type createDeptReq struct {
    Name string `json:"name" validate:"required,max=120"`
    Code string `json:"code" validate:"required,alphanum,max=12"`
    HeadID *string `json:"headId" validate:"omitempty,uuid"`
}
```
- **Error mapping** (single helper): `errs.Kind → status` (`invalid→400, not_found→404, conflict→409, forbidden→403`,
  default `500`), body = the error envelope.

**Endpoints delivered**
| Method | Path | Role | Body → Resp | Codes |
| --- | --- | --- | --- | --- |
| POST | `/auth/login` | public | `{email,password}` → `{accessToken,refreshToken,user}` | 200/401 |
| POST | `/auth/refresh` | public | `{refreshToken}` → `{accessToken,refreshToken}` (verifies + rotates stored token) | 200/401 |
| GET | `/me` | any | → `{id,name,role,departmentId}` | 200 |
| POST | `/departments` | admin | `createDeptReq` → `Department` | 201/400/409 |
| GET | `/departments?limit&offset` | any | → `Result<Department>` | 200 |

**Tests:** migration up/down clean; `POST /departments` 409 on dup code; RBAC blocks non-admin with 403; login bad
password 401.

---

## P3 — Frontend (MVVM)

### Structure (`web/src`)
```text
app/  router.tsx  AppShell.tsx  RoleGuard.tsx  providers.tsx
design/ tokens.ts  {Button,Card,Pill,Tile,Table,Tabs,Switch,Input,StatBar}.tsx
lib/  apiClient.ts  auth.ts  queryKeys.ts  format.ts
modules/department/ model/useDepartments.ts  viewmodel/useDepartmentsVM.ts  view/DepartmentsPage.tsx
```

### Design tokens (`design/tokens.ts`) — mirror `theme.css`
```ts
export const color = { brand:'#714B67', yellow:'#E0A800', danger:'#C0392B',
  ink:'#1F2328', muted:'#6C757D', line:'#E9ECEF', canvas:'#F6F7F8' } as const;
```

### Typed API client (`lib/apiClient.ts`) — generated from `api/openapi.yaml`
```ts
export const api = {
  departments: {
    list: (p:PageParams) => get<Result<Department>>('/departments', p),
    create: (b:CreateDeptReq) => post<Department>('/departments', b),
  },
  auth: { login:(b)=>post<AuthResp>('/auth/login', b) },
};
```

### Auth store (Zustand)
```ts
type AuthState = { user?:User; token?:string; login(email,pw):Promise<void>; logout():void };
```

### Departments — Model / ViewModel / View
```ts
// model — server cache only
export const useDepartments = (p:PageParams) =>
  useQuery({ queryKey: queryKeys.departments(p), queryFn: () => api.departments.list(p) });

// viewmodel — state + intents + derived (unit-tested, no JSX)
export function useDepartmentsVM() {
  const [page,setPage] = useState({limit:20,offset:0});
  const q = useDepartments(page);
  const create = useMutation({ mutationFn: api.departments.create,
    onSuccess: () => qc.invalidateQueries({queryKey:queryKeys.departments(page)}) });
  return { rows:q.data?.items ?? [], total:q.data?.total ?? 0, loading:q.isLoading,
           page, setPage, createDepartment: create.mutateAsync, creating: create.isPending };
}

// view — dumb; renders VM, uses design primitives only
export function DepartmentsPage(){ const vm = useDepartmentsVM(); return <Table …/> }
```

**Deliverables:** SPA logs in, lists + creates departments; `RoleGuard` hides admin-only actions.
**Tests:** `useDepartmentsVM` unit test with mocked `api` (create → invalidate); RoleGuard renders/blocks by role.

---

## P4 — Platform · AI · Notifications · DevOps

### docker-compose.yml (services)
`api` (Go), `postgres:16`, `mailhog` (SMTP 1025 / UI 8025), `minio` (S3 API). `.env.example` documents every var.

### Auth service (`platform/auth`)
```go
type Claims struct { Sub string; Role string; DeptID *string; jwt.RegisteredClaims }
func Issue(u User, ttl time.Duration, secret []byte) (string, error)
func Verify(token string, secret []byte) (*Claims, error)
func Hash(pw string) (string, error)        // argon2id
func Check(pw, hash string) bool
// refresh tokens are persisted hashed in refresh_tokens; Refresh verifies + ROTATES (revoke old, issue new pair)
func Refresh(ctx context.Context, raw string) (access, refresh string, err error)
```

### RBAC middleware
```go
func RequireRole(roles ...Role) func(http.Handler) http.Handler // 401 if no/!valid token, 403 if role not allowed
```

### Outbound port stubs (real adapters land in later phases)
```go
// AIGateway — implemented by platform/ai (OpenRouter). P0 = stub returning fixtures.
type AIGateway interface { Categorize(ctx, DocInput) (Suggestion, error) }
// EmailSender — MailHog in dev.
type EmailSender interface { Send(ctx, Email) error }
// Storage — minio.
type Storage interface { Put(ctx, key string, r io.Reader) (url string, err error); SignedURL(key string) string }
```

### CI (`.github/workflows/ci.yml`)
`golangci-lint` · `go test ./...` · `migrate up` against ephemeral PG · `pnpm lint && pnpm test && pnpm build`.

**Deliverables:** `docker compose up` → full local stack; JWT issue/verify + hashing; CI green on PR.

---

## Integration & sync
Mid-phase: P2+P3 align on OpenAPI for `/departments`. End-of-phase: wire P1 use-cases into P2 handlers, P3 hits the real
API, P4 confirms auth + compose. **Demo:** log in as admin → create a department → see it listed.

## Definition of Done
Compose stack boots · login issues JWT · Departments CRUD works browser→DB · migrations reversible · CI green ·
module template documented so Phase 1 can parallelize cleanly.
