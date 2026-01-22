package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"inbox-service/internal/application/ports"

	"github.com/jackc/pgx/v5"
)

type InboxWriterPG struct{}

func NewInboxWriterPG() *InboxWriterPG { return &InboxWriterPG{} }

func (w *InboxWriterPG) InsertInboxItem(ctx context.Context, tx ports.Tx, in ports.InsertInboxItemParams) error {
	now := time.Now().UTC()

	_, err := tx.Exec(ctx, `
		INSERT INTO inbox_items (
			id, tenant_id, user_id,
			type, status,
			title, body, action_url,
			source_event_id, dedupe_key,
			created_at, updated_at, version
		) VALUES (
			$1,$2,$3,
			$4,$5,
			$6,$7,$8,
			$9,$10,
			$11,$12,1
		)
	`, in.ID, in.TenantID, in.UserID,
		in.Type, in.Status,
		in.Title, in.Body, in.ActionURL,
		in.SourceEventID, in.DedupeKey,
		now, now,
	)
	if err == nil {
		return nil
	}

	// Treat unique constraint on dedupe as idempotent success
	if isUniqueViolation(err) {
		return nil
	}

	return fmt.Errorf("insert inbox item: %w", err)
}

func isUniqueViolation(err error) bool {
	// pgx exposes pgconn.PgError normally; but keep it simple for now.
	// This string check is not ideal; later replace with errors.As(*pgconn.PgError).
	msg := err.Error()
	return errors.Is(err, pgx.ErrNoRows) || strings.Contains(msg, "duplicate key value") || strings.Contains(msg, "unique constraint")
}
