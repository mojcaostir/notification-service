package db

import (
	"context"
	"fmt"
	"time"

	"inbox-service/internal/application/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedReaderPG struct {
	pool *pgxpool.Pool
}

func NewFeedReaderPG(pool *pgxpool.Pool) *FeedReaderPG {
	return &FeedReaderPG{pool: pool}
}

func (r *FeedReaderPG) GetFeed(ctx context.Context, tenantID, userID string, f ports.FeedFilter) (ports.FeedPage, error) {
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	args := []any{tenantID, userID}
	where := "WHERE tenant_id = $1 AND user_id = $2"

	argN := 3
	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, f.Status)
		argN++
	}

	if f.Cursor != nil {
		where += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", argN, argN+1)
		args = append(args, f.Cursor.CreatedAt, f.Cursor.ID)
		argN += 2
	}

	args = append(args, limit)

	q := fmt.Sprintf(`
		SELECT id, type, status, title, body, action_url, created_at
		FROM inbox_items
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, where, argN)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return ports.FeedPage{}, fmt.Errorf("query feed: %w", err)
	}
	defer rows.Close()

	items := make([]ports.FeedItem, 0, limit)
	var lastCreatedAt time.Time
	var lastID string

	for rows.Next() {
		var it ports.FeedItem
		if err := rows.Scan(&it.ID, &it.Type, &it.Status, &it.Title, &it.Body, &it.ActionURL, &it.CreatedAt); err != nil {
			return ports.FeedPage{}, fmt.Errorf("scan feed row: %w", err)
		}
		items = append(items, it)
		lastCreatedAt = it.CreatedAt
		lastID = it.ID
	}

	if err := rows.Err(); err != nil {
		return ports.FeedPage{}, fmt.Errorf("feed rows err: %w", err)
	}

	var next *ports.FeedCursor
	if len(items) == limit {
		next = &ports.FeedCursor{CreatedAt: lastCreatedAt, ID: lastID}
	}

	return ports.FeedPage{Items: items, NextCursor: next}, nil
}
