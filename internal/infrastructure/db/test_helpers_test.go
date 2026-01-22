package db

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testDBURL() string {
	if v := os.Getenv("TEST_DB_URL"); v != "" {
		return v
	}
	// default matches your docker-compose
	return "postgres://inbox:inbox@localhost:5432/inbox?sslmode=disable"
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()
	pool, err := NewPool(ctx, Config{URL: testDBURL()})
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func truncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Keep this list in sync with your schema
	_, err := pool.Exec(ctx, `TRUNCATE TABLE inbox_items, processed_events, outbox`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func insertInboxItem(
	t *testing.T,
	pool *pgxpool.Pool,
	id, tenantID, userID, typ, status, title, body, url, sourceEventID, dedupeKey string,
	createdAt time.Time,
) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, `
		INSERT INTO inbox_items (
			id, tenant_id, user_id, type, status,
			title, body, action_url,
			source_event_id, dedupe_key,
			created_at, updated_at, version
		) VALUES (
			$1,$2,$3,$4,$5,
			$6,$7,$8,
			$9,$10,
			$11,$12,1
		)
	`, id, tenantID, userID, typ, status, title, body, url, sourceEventID, dedupeKey, createdAt, createdAt)
	if err != nil {
		t.Fatalf("insertInboxItem: %v", fmt.Errorf("%w", err))
	}
}
