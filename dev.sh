#!/usr/bin/env bash
# EcoSphere local development — start Postgres, migrate, API, web, MinIO, MailHog.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

COMPOSE=(docker compose)
API_URL="${API_URL:-http://localhost:8080}"
WEB_URL="${WEB_URL:-http://localhost:5173}"
MAILHOG_URL="${MAILHOG_URL:-http://localhost:8025}"
MINIO_URL="${MINIO_URL:-http://localhost:9001}"
SEED_EMAIL="${SEED_ADMIN_EMAIL:-admin@ecosphere.local}"
SEED_PASSWORD="${SEED_ADMIN_PASSWORD:-ChangeMe123!}"

usage() {
  cat <<'EOF'
Usage: ./dev.sh [command] [options]

Commands:
  up        Start all services (default) — DB, migrate, API, web, MinIO, MailHog
  down      Stop and remove containers (keeps volumes)
  reset     Stop and remove containers + volumes (fresh DB)
  seed      Load demo seed data (admin + departments)
  status    Show container status and endpoint health
  logs      Tail logs (optional service: api|web|postgres|minio|mailhog)
  rebuild   Rebuild images and restart the stack
  help      Show this help

Options for up/rebuild:
  --no-seed   Skip demo seed after start
  --detach    Already the default; stack runs in background

Examples:
  ./dev.sh
  ./dev.sh up
  ./dev.sh up --no-seed
  ./dev.sh seed
  ./dev.sh logs api
  ./dev.sh status
  ./dev.sh down
  ./dev.sh reset
EOF
}

need_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "error: docker is required" >&2
    exit 1
  fi
  if ! docker compose version >/dev/null 2>&1; then
    echo "error: docker compose plugin is required" >&2
    exit 1
  fi
}

wait_http() {
  local url="$1" name="$2" attempts="${3:-60}"
  local i
  for ((i = 1; i <= attempts; i++)); do
    if curl -sf -o /dev/null "$url" 2>/dev/null; then
      echo "  ✓ $name ready ($url)"
      return 0
    fi
    sleep 1
  done
  echo "  ✗ $name not ready after ${attempts}s ($url)" >&2
  return 1
}

wait_postgres() {
  local i
  for ((i = 1; i <= 60; i++)); do
    if "${COMPOSE[@]}" exec -T postgres pg_isready -U ecosphere >/dev/null 2>&1; then
      echo "  ✓ postgres ready (localhost:5433)"
      return 0
    fi
    sleep 1
  done
  echo "  ✗ postgres not ready" >&2
  return 1
}

print_banner() {
  cat <<EOF

╔══════════════════════════════════════════════════════════╗
║              EcoSphere ESG — local stack                 ║
╠══════════════════════════════════════════════════════════╣
║  Web (portals)     $WEB_URL
║  API               $API_URL
║  Postgres          localhost:5433  (user/db: ecosphere)
║  MinIO console     $MINIO_URL  (ecosphere / ecosphere-secret)
║  MailHog UI        $MAILHOG_URL
╠══════════════════════════════════════════════════════════╣
║  Demo login        $SEED_EMAIL
║  Password          $SEED_PASSWORD
╚══════════════════════════════════════════════════════════╝

  Portals (after login): dashboard · environmental · social
  governance · gamification · reports · settings

  Logs:    ./dev.sh logs [service]
  Status:  ./dev.sh status
  Stop:    ./dev.sh down

EOF
}

cmd_seed() {
  echo "→ Seeding demo data…"
  # Seed binary ships in the API image (see Dockerfile).
  SEED_ADMIN_EMAIL="$SEED_EMAIL" SEED_ADMIN_PASSWORD="$SEED_PASSWORD" \
    "${COMPOSE[@]}" run --rm --entrypoint /usr/local/bin/seed \
    -e DATABASE_URL=postgres://ecosphere:ecosphere@postgres:5432/ecosphere?sslmode=disable \
    -e SEED_ADMIN_EMAIL="$SEED_EMAIL" \
    -e SEED_ADMIN_PASSWORD="$SEED_PASSWORD" \
    api
  echo "  ✓ seed complete"
}

cmd_up() {
  local do_seed=1
  local build_flag=(--build)
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --no-seed) do_seed=0 ;;
      --no-build) build_flag=() ;;
      --detach|-d) ;; # always detached
      *)
        echo "unknown option: $1" >&2
        usage
        exit 1
        ;;
    esac
    shift
  done

  need_docker
  echo "→ Starting EcoSphere (postgres · migrate · api · web · minio · mailhog)…"
  "${COMPOSE[@]}" up -d "${build_flag[@]}"

  echo "→ Waiting for services…"
  wait_postgres
  wait_http "$API_URL/health" "api"
  wait_http "$WEB_URL" "web" 90 || true

  if [[ "$do_seed" -eq 1 ]]; then
    cmd_seed || echo "  ! seed failed (stack is still up — re-run: ./dev.sh seed)" >&2
  fi

  print_banner
  cmd_status_brief
}

cmd_rebuild() {
  need_docker
  echo "→ Rebuilding images…"
  "${COMPOSE[@]}" build
  shift || true
  cmd_up --no-build "$@"
}

cmd_down() {
  need_docker
  echo "→ Stopping stack (volumes preserved)…"
  "${COMPOSE[@]}" down
  echo "  ✓ stopped"
}

cmd_reset() {
  need_docker
  echo "→ Resetting stack (containers + volumes)…"
  "${COMPOSE[@]}" down -v
  echo "  ✓ reset complete — run ./dev.sh up for a fresh environment"
}

cmd_status_brief() {
  echo "Service status:"
  "${COMPOSE[@]}" ps --format 'table {{.Name}}\t{{.Status}}\t{{.Ports}}' 2>/dev/null \
    || "${COMPOSE[@]}" ps
  echo
  local h w
  h=$(curl -sf -o /dev/null -w '%{http_code}' "$API_URL/health" 2>/dev/null || echo down)
  w=$(curl -sf -o /dev/null -w '%{http_code}' "$WEB_URL" 2>/dev/null || echo down)
  echo "  API health:  $h"
  echo "  Web:         $w"
}

cmd_status() {
  need_docker
  print_banner
  cmd_status_brief
  echo
  if "${COMPOSE[@]}" exec -T postgres pg_isready -U ecosphere >/dev/null 2>&1; then
    echo "DB tables (sample):"
    "${COMPOSE[@]}" exec -T postgres psql -U ecosphere -d ecosphere -c \
      "SELECT COUNT(*) AS users FROM users; SELECT COUNT(*) AS departments FROM departments;" 2>/dev/null || true
  fi
}

cmd_logs() {
  need_docker
  local svc="${1:-}"
  if [[ -n "$svc" ]]; then
    "${COMPOSE[@]}" logs -f --tail=200 "$svc"
  else
    "${COMPOSE[@]}" logs -f --tail=100
  fi
}

main() {
  local cmd="${1:-up}"
  shift || true
  case "$cmd" in
    up|start) cmd_up "$@" ;;
    down|stop) cmd_down "$@" ;;
    reset|destroy) cmd_reset "$@" ;;
    seed) need_docker; cmd_seed "$@" ;;
    status|ps) cmd_status "$@" ;;
    logs|log) cmd_logs "$@" ;;
    rebuild) cmd_rebuild "$@" ;;
    help|-h|--help) usage ;;
    *)
      echo "unknown command: $cmd" >&2
      usage
      exit 1
      ;;
  esac
}

main "$@"
