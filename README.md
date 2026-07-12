# EcoSphere

**ESG management platform** that unifies Environmental, Social, and Governance performance with gamified employee participation — and a single, trustworthy **Overall ESG Score**.

EcoSphere turns operational data into measurable sustainability outcomes: deterministic carbon accounting, CSR engagement, policy compliance, audits, challenges & rewards, and exportable reports.

---

## Features

| Pillar | Capabilities |
|--------|----------------|
| **Environmental** | Emission factors, carbon transactions, goals, source breakdown |
| **Social** | CSR activities, participation & proof, diversity metrics, training |
| **Governance** | Policies & acknowledgements, audits, compliance issues |
| **Gamification** | Challenges, XP, badges, rewards, leaderboards |
| **Insights** | Live ESG scores, executive dashboard, reports with PDF / Excel / CSV |
| **Platform** | Role-based portals, notifications (in-app + email), configurable ESG weights |

**AI (advisory only)**  
Optional assist for document categorization, participation proof review, and report narrative. Humans always approve; CO₂ remains **quantity × emission factor** (never invented by AI).

**Roles**  
Admin · Department head · Auditor · Employee — each with a focused home experience.

---

## Screenshots & design

- Product wireframes: [`wireframes/index.html`](wireframes/index.html)  
- Design system: [`design.md`](design.md)

---

## Tech stack

| Layer | Stack |
|-------|--------|
| API | Go (modular / hexagonal layout) |
| Web | React, TypeScript, Vite |
| Database | PostgreSQL |
| Object storage | MinIO (S3-compatible) |
| Email (dev) | SMTP → MailHog |
| AI (optional) | OpenRouter-compatible gateway |

---

## Quick start

**Requirements:** [Docker](https://docs.docker.com/get-docker/) and Docker Compose.

```bash
git clone <your-repo-url> ecosphere-esg-platform
cd ecosphere-esg-platform

cp .env.example .env
./dev.sh
```

This builds and starts Postgres, migrations, API, web, MinIO, and MailHog, then loads demo seed data.

| Service | URL |
|---------|-----|
| Application | http://localhost:5173 |
| API | http://localhost:8080 |
| Health | http://localhost:8080/health |
| Email (local capture) | http://localhost:8025 |
| MinIO console | http://localhost:9001 |

**Demo admin** (from seed; change in production):

```
Email:    admin@ecosphere.local
Password: ChangeMe123!
```

Other roles are seeded for department heads, auditors, and employees (same default password). See `.env.example` for seed overrides.

---

## Configuration

Copy [`.env.example`](.env.example) to `.env` and adjust as needed.

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Signing secret (≥ 24 characters) |
| `SMTP_ADDR` | SMTP host for outbound mail |
| `OPENROUTER_API_KEY` | Optional — live AI assist |
| `AI_FIXTURE_MODE` | `true` uses offline AI fixtures; `false` uses the live gateway when a key is set |

Do not commit secrets. Prefer environment injection in production.

---

## Development

```bash
./dev.sh status          # service health
./dev.sh logs api        # API logs
./dev.sh down            # stop stack (volumes kept)
./dev.sh reset           # stop and remove volumes

# Tests
go test ./...
cd web && pnpm install && pnpm test --run && pnpm build
```

Useful Makefile targets: `make up`, `make down`, `make seed`, `make test`, `make build`.

Architecture and phase plans live under [`plan.md`](plan.md) and [`docs/phases/`](docs/phases/).

---

## Project layout

```
cmd/api/              HTTP API entrypoint
cmd/seed/             Demo data seeder
internal/modules/     Domain modules (environmental, social, governance, …)
internal/platform/    Cross-cutting (auth, AI, email, events, storage)
migrations/           SQL migrations
web/                  React frontend
wireframes/           Design reference
api/openapi.yaml      API sketch
```

---

## Security notes

- AI outputs are **advisory**; approval and carbon math remain human / deterministic.
- Proof verification can run on ephemeral uploads; design for least privilege on storage and keys.
- Rotate JWT secrets, SMTP credentials, and AI keys for any shared or production deployment.
- Default seed passwords are for **local demo only**.

---

## License

Proprietary — all rights reserved unless otherwise stated by the repository owners.

---

## Contributing

Internal development follows the phase docs and design system. Open an issue or PR according to your team’s process; keep secrets out of the tree and match wireframe patterns for UI changes.
