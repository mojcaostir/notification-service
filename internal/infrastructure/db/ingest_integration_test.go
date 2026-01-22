package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"inbox-service/internal/application/ingest"
)

func TestIngest_TaskAssigned_WritesAllTables(t *testing.T) {
	pool := newTestPool(t)
	truncateAll(t, pool)

	// Build real adapters
	txMgr := NewTxManagerPG(pool)
	inboxWriter := NewInboxWriterPG()
	deduper := NewEventDeduperPG()
	outboxWriter := NewOutboxWriterPG()

	h := ingest.NewHandler(txMgr, inboxWriter, deduper, outboxWriter)

	evt := ingest.TaskAssignedToUser{
		EventID:        "99999999-9999-9999-9999-999999999999",
		OccurredAt:     time.Now().UTC(),
		TenantID:       "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		TaskID:         "42",
		AssigneeUserID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		AssignerUserID: "cccccccc-cccc-cccc-cccc-cccccccccccc",
		TaskTitle:      "Prepare quarterly report",
		TaskURL:        "https://app.example.com/tasks/42",
		Version:        1,
	}

	itemID, err := h.HandleTaskAssigned(context.Background(), evt)
	if err != nil {
		t.Fatalf("HandleTaskAssigned: %v", err)
	}
	if itemID == "" {
		t.Fatalf("expected non-empty inbox item id")
	}

	// 1) inbox_items row exists
	var got int
	err = pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM inbox_items
		WHERE tenant_id = $1 AND user_id = $2 AND source_event_id = $3
	`, evt.TenantID, evt.AssigneeUserID, evt.EventID).Scan(&got)
	if err != nil {
		t.Fatalf("count inbox_items: %v", err)
	}
	if got != 1 {
		t.Fatalf("expected 1 inbox_items row, got %d", got)
	}

	// 2) processed_events row exists
	err = pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM processed_events
		WHERE tenant_id = $1 AND event_id = $2
	`, evt.TenantID, evt.EventID).Scan(&got)
	if err != nil {
		t.Fatalf("count processed_events: %v", err)
	}
	if got != 1 {
		t.Fatalf("expected 1 processed_events row, got %d", got)
	}

	// 3) outbox row exists and is PENDING
	var status string
	var payload []byte
	err = pool.QueryRow(context.Background(), `
		SELECT status, payload_json::text
		FROM outbox
		WHERE tenant_id = $1 AND event_type = 'InboxItemCreated'
		ORDER BY created_at DESC
		LIMIT 1
	`, evt.TenantID).Scan(&status, &payload)
	if err != nil {
		t.Fatalf("select outbox: %v", err)
	}
	if status != "PENDING" {
		t.Fatalf("expected outbox status PENDING, got %q", status)
	}

	// Optional: verify payload is valid json
	var tmp map[string]any
	if err := json.Unmarshal(payload, &tmp); err != nil {
		t.Fatalf("outbox payload is not valid json: %v", err)
	}
}


func TestIngest_TaskAssigned_IsIdempotent(t *testing.T) {
	pool := newTestPool(t)
	truncateAll(t, pool)

	txMgr := NewTxManagerPG(pool)
	inboxWriter := NewInboxWriterPG()
	deduper := NewEventDeduperPG()
	outboxWriter := NewOutboxWriterPG()

	h := ingest.NewHandler(txMgr, inboxWriter, deduper, outboxWriter)

	evt := ingest.TaskAssignedToUser{
		EventID:        "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		OccurredAt:     time.Now().UTC(),
		TenantID:       "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		TaskID:         "42",
		AssigneeUserID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		AssignerUserID: "cccccccc-cccc-cccc-cccc-cccccccccccc",
		TaskTitle:      "Prepare quarterly report",
		TaskURL:        "https://app.example.com/tasks/42",
		Version:        1,
	}

	_, err := h.HandleTaskAssigned(context.Background(), evt)
	if err != nil {
		t.Fatalf("first HandleTaskAssigned: %v", err)
	}
	_, err = h.HandleTaskAssigned(context.Background(), evt)
	if err != nil {
		t.Fatalf("second HandleTaskAssigned should be idempotent: %v", err)
	}

	// inbox_items should be 1 (dedupe)
	var cnt int
	err = pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM inbox_items
		WHERE tenant_id = $1 AND user_id = $2 AND source_event_id = $3
	`, evt.TenantID, evt.AssigneeUserID, evt.EventID).Scan(&cnt)
	if err != nil {
		t.Fatalf("count inbox_items: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 inbox item, got %d", cnt)
	}

	// processed_events should be 1
	err = pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM processed_events
		WHERE tenant_id = $1 AND event_id = $2
	`, evt.TenantID, evt.EventID).Scan(&cnt)
	if err != nil {
		t.Fatalf("count processed_events: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 processed event, got %d", cnt)
	}

	// outbox: we currently insert outbox only on first processing because we short-circuit on AlreadyProcessed.
	// So expected outbox rows: 1
	err = pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM outbox
		WHERE tenant_id = $1 AND event_type = 'InboxItemCreated'
	`, evt.TenantID).Scan(&cnt)
	if err != nil {
		t.Fatalf("count outbox: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 outbox row, got %d", cnt)
	}
}
