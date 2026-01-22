package ingest

import (
	"context"
	"testing"
	"time"

	"inbox-service/internal/application/ports"
)

// --- fakes ---

type fakeTxMgr struct {
	fn func(ctx context.Context, tx ports.Tx) error
}

func (m fakeTxMgr) WithTx(ctx context.Context, fn func(ctx context.Context, tx ports.Tx) error) error {
	// We won't call fn for validation tests; return nil by default.
	if m.fn != nil {
		return m.fn(ctx, nil)
	}
	return nil
}

type fakeInboxWriter struct{}
func (w fakeInboxWriter) InsertInboxItem(ctx context.Context, tx ports.Tx, in ports.InsertInboxItemParams) error { return nil }

type fakeDeduper struct{}
func (d fakeDeduper) AlreadyProcessed(ctx context.Context, tx ports.Tx, tenantID, eventID string) (bool, error) { return false, nil }
func (d fakeDeduper) MarkProcessed(ctx context.Context, tx ports.Tx, tenantID, eventID string) error { return nil }

type fakeOutboxWriter struct{}
func (w fakeOutboxWriter) InsertOutbox(ctx context.Context, tx ports.Tx, e ports.OutboxEvent) error { return nil }

// --- tests ---

func TestHandleTaskAssigned_Validation(t *testing.T) {
	h := NewHandler(fakeTxMgr{}, fakeInboxWriter{}, fakeDeduper{}, fakeOutboxWriter{})

	now := time.Now().UTC()

	_, err := h.HandleTaskAssigned(context.Background(), TaskAssignedToUser{
		EventID:        "",
		OccurredAt:     now,
		TenantID:       "t",
		TaskID:         "42",
		AssigneeUserID: "u",
		TaskTitle:      "X",
		TaskURL:        "https://x",
	})
	if err == nil {
		t.Fatalf("expected error for missing event_id")
	}

	_, err = h.HandleTaskAssigned(context.Background(), TaskAssignedToUser{
		EventID:        "e",
		OccurredAt:     now,
		TenantID:       "",
		TaskID:         "42",
		AssigneeUserID: "u",
		TaskTitle:      "X",
		TaskURL:        "https://x",
	})
	if err == nil {
		t.Fatalf("expected error for missing tenant_id")
	}

	_, err = h.HandleTaskAssigned(context.Background(), TaskAssignedToUser{
		EventID:        "e",
		OccurredAt:     now,
		TenantID:       "t",
		TaskID:         "",
		AssigneeUserID: "u",
		TaskTitle:      "X",
		TaskURL:        "https://x",
	})
	if err == nil {
		t.Fatalf("expected error for missing task_id")
	}

	_, err = h.HandleTaskAssigned(context.Background(), TaskAssignedToUser{
		EventID:        "e",
		OccurredAt:     now,
		TenantID:       "t",
		TaskID:         "42",
		AssigneeUserID: "u",
		TaskTitle:      "X",
		TaskURL:        "",
	})
	if err == nil {
		t.Fatalf("expected error for missing task_url")
	}
}
