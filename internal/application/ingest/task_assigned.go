package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"inbox-service/internal/application/ports"

	"github.com/google/uuid"
)

type TaskAssignedToUser struct {
	EventID         string    `json:"event_id"`
	OccurredAt      time.Time `json:"occurred_at"`
	TenantID        string    `json:"tenant_id"`
	TaskID          string    `json:"task_id"`
	AssigneeUserID  string    `json:"assignee_user_id"`
	AssignerUserID  string    `json:"assigner_user_id"`
	TaskTitle       string    `json:"task_title"`
	TaskURL         string    `json:"task_url"`
	Version         int       `json:"version"`
}

type Handler struct {
	Tx       ports.TxManager
	Inbox    ports.InboxWriter
	Deduper  ports.EventDeduper
	Outbox   ports.OutboxWriter
}

func NewHandler(tx ports.TxManager, inbox ports.InboxWriter, deduper ports.EventDeduper, outbox ports.OutboxWriter) *Handler {
	return &Handler{Tx: tx, Inbox: inbox, Deduper: deduper, Outbox: outbox}
}

// HandleTaskAssigned is idempotent: safe under at-least-once delivery.
func (h *Handler) HandleTaskAssigned(ctx context.Context, evt TaskAssignedToUser) (string, error) {
	// minimal validation
	if evt.EventID == "" || evt.TenantID == "" || evt.AssigneeUserID == "" || evt.TaskID == "" {
		return "", fmt.Errorf("invalid event: missing required fields")
	}
	if evt.TaskURL == "" {
		return "", fmt.Errorf("invalid event: task_url required")
	}

	inboxItemID := uuid.NewString()
	dedupeKey := fmt.Sprintf("TASK_ASSIGNED:%s:%s", evt.TaskID, evt.AssigneeUserID)

	// Prepare outbox payload (event-carried state: we include enough for downstream)
	outPayload, err := json.Marshal(map[string]any{
		"event_id":         uuid.NewString(),
		"occurred_at":      time.Now().UTC().Format(time.RFC3339Nano),
		"tenant_id":        evt.TenantID,
		"user_id":          evt.AssigneeUserID,
		"inbox_item_id":    inboxItemID,
		"type":             "TASK_ASSIGNED",
		"source_event_id":  evt.EventID,
		"task_id":          evt.TaskID,
		"task_title":       evt.TaskTitle,
		"task_url":         evt.TaskURL,
		"schema_version":   1,
	})
	if err != nil {
		return "", fmt.Errorf("marshal outbox payload: %w", err)
	}

	err = h.Tx.WithTx(ctx, func(ctx context.Context, tx ports.Tx) error {
		dup, err := h.Deduper.AlreadyProcessed(ctx, tx, evt.TenantID, evt.EventID)
		if err != nil {
			return err
		}
		if dup {
			// already processed -> idempotent no-op
			return nil
		}

		// Insert inbox item (may still be deduped by unique constraint)
		err = h.Inbox.InsertInboxItem(ctx, tx, ports.InsertInboxItemParams{
			ID:            inboxItemID,
			TenantID:      evt.TenantID,
			UserID:        evt.AssigneeUserID,
			Type:          "TASK_ASSIGNED",
			Status:        "UNREAD",
			Title:         "Task assigned to you",
			Body:          fmt.Sprintf("You have been assigned a new task: %s", evt.TaskTitle),
			ActionURL:     evt.TaskURL,
			SourceEventID: evt.EventID,
			DedupeKey:     dedupeKey,
		})
		if err != nil {
			return err
		}

		// Mark event processed (idempotency)
		if err := h.Deduper.MarkProcessed(ctx, tx, evt.TenantID, evt.EventID); err != nil {
			return err
		}

		// Insert outbox event (PENDING)
		if err := h.Outbox.InsertOutbox(ctx, tx, ports.OutboxEvent{
			ID:        uuid.NewString(),
			TenantID:  evt.TenantID,
			EventType: "InboxItemCreated",
			Payload:   outPayload,
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}
	return inboxItemID, nil
}
