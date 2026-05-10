# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Start Postgres (required for running the service and integration tests)
docker-compose up -d

# Run the API server (defaults to :8080)
go run ./cmd/api

# Run all tests (requires Postgres to be running)
go test ./... -count=1

# Run a single test package
go test ./internal/infrastructure/db/... -count=1

# Run a single test by name
go test ./internal/infrastructure/db/... -run TestIngest_TaskAssigned_IsIdempotent -count=1

# Reset Postgres schema (when migrations don't apply cleanly)
docker-compose down -v && docker-compose up -d
```

Environment variables (see `.env.example`):
- `DB_URL` — main connection string (default: `postgres://inbox:inbox@localhost:5432/inbox?sslmode=disable`)
- `HTTP_ADDR` — listen address (default: `:8080`)
- `TEST_DB_URL` — override DB URL for integration tests

## Architecture

This is a Go microservice implementing a **user inbox** via CQRS and the Transactional Outbox pattern.

### Layer structure

```
cmd/api/              — entrypoint: wires all dependencies, starts Echo
internal/
  application/
    ports/            — interfaces (FeedReader, InboxWriter, EventDeduper, OutboxWriter, TxManager)
    queries/          — CQRS read side: FeedHandler reads DTOs, no domain logic
    ingest/           — CQRS write side: HandleTaskAssigned, idempotent event ingestion
  infrastructure/
    db/               — Postgres implementations of all ports (pgx/v5)
    http/             — Echo handlers and route registration
```

### Key design decisions

**CQRS split**: `application/queries` is the read path (returns `FeedPage` DTOs directly from SQL). `application/ingest` is the write path (business rules, transaction coordination). They share no domain objects.

**Ports pattern**: Application logic only depends on interfaces in `application/ports`. Infrastructure implementations in `infrastructure/db` are injected in `cmd/api/main.go`. This is what makes the unit tests possible without a DB.

**Idempotent ingest**: Every inbound event carries an `event_id`. `HandleTaskAssigned` checks `processed_events` inside the same transaction before writing. A unique index on `inbox_items(tenant_id, dedupe_key)` provides a second safety net for race conditions.

**Transactional Outbox**: A single `WithTx` call writes `inbox_items`, `processed_events`, and `outbox` atomically. The outbox dispatcher (not yet implemented) will publish `InboxItemCreated` events asynchronously.

### DB schema

Three tables (defined in `internal/infrastructure/db/migrations/init.sql`):
- `inbox_items` — one row per notification; unique on `(tenant_id, dedupe_key)`
- `processed_events` — idempotency log; primary key `(tenant_id, event_id)`
- `outbox` — reliable publish queue; status values: `PENDING | PROCESSING | SENT | FAILED`

### HTTP API

- `GET /v1/inbox/feed` — cursor-based paginated feed. Auth headers: `X-Tenant-Id`, `X-User-Id`. Cursor params: `cursor_created_at` (RFC3339Nano) + `cursor_id`.
- `POST /v1/dev/ingest/task-assigned` — dev-only endpoint to trigger event ingestion.

### Testing

Integration tests live alongside the infrastructure code in `internal/infrastructure/db/` and run against a real Postgres instance. `test_helpers_test.go` provides `newTestPool`, `truncateAll`, and `insertInboxItem` helpers shared across test files. There is no mock for the DB layer.
