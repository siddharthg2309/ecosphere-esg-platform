.PHONY: dev up down migrate-up migrate-down seed test test-go test-web lint build

up:
	docker compose up --build

down:
	docker compose down -v

migrate-up:
	docker compose run --rm migrate -path=/migrations -database "$${DATABASE_URL:-postgres://ecosphere:ecosphere@postgres:5432/ecosphere?sslmode=disable}" up

migrate-down:
	docker compose run --rm migrate -path=/migrations -database "$${DATABASE_URL:-postgres://ecosphere:ecosphere@postgres:5432/ecosphere?sslmode=disable}" down 1

seed:
	docker compose run --rm api go run ./cmd/seed

test: test-go test-web

test-go:
	docker run --rm -v "$$(pwd):/app" -w /app golang:1.22-alpine go test ./...

test-web:
	cd web && pnpm test --run && pnpm build

lint:
	docker run --rm -v "$$(pwd):/app" -w /app golang:1.22-alpine sh -c 'gofmt -l . | tee /tmp/gofmt && test ! -s /tmp/gofmt && go vet ./...'
	cd web && pnpm lint

build:
	docker compose build api web
