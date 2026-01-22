package db

import (
	"context"
	"fmt"
	"time"

	"inbox-service/internal/application/ports"
)

type EventDeduperPG struct{}

func NewEventDeduperPG() *EventDeduperPG { return &EventDeduperPG{} }

func (d *EventDeduperPG) AlreadyProcessed(ctx context.Context, tx ports.Tx, tenantID, eventID string) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM processed_events
			WHERE tenant_id = $1 AND event_id = $2
		)
	`, tenantID, eventID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check processed: %w", err)
	}
	return exists, nil
}

func (d *EventDeduperPG) MarkProcessed(ctx context.Context, tx ports.Tx, tenantID, eventID string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO processed_events (tenant_id, event_id, processed_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, event_id) DO NOTHING
	`, tenantID, eventID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}
	return nil
}
