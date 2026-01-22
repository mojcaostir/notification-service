package db

import (
	"context"
	"fmt"
	"time"

	"inbox-service/internal/application/ports"
)

type OutboxWriterPG struct{}

func NewOutboxWriterPG() *OutboxWriterPG { return &OutboxWriterPG{} }

func (w *OutboxWriterPG) InsertOutbox(ctx context.Context, tx ports.Tx, e ports.OutboxEvent) error {
	now := time.Now().UTC()
	_, err := tx.Exec(ctx, `
		INSERT INTO outbox (
			id, tenant_id, event_type, payload_json,
			status, attempts, next_run_at,
			created_at, updated_at, version
		) VALUES (
			$1,$2,$3,$4,
			'PENDING', 0, $5,
			$6,$7,1
		)
	`, e.ID, e.TenantID, e.EventType, e.Payload, now, now, now)
	if err != nil {
		return fmt.Errorf("insert outbox: %w", err)
	}
	return nil
}
