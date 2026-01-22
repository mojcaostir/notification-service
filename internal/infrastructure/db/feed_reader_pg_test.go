package db

import (
	"context"
	"testing"
	"time"

	"inbox-service/internal/application/ports"
)

func TestFeedReaderPG_OrderAndCursorPagination(t *testing.T) {
	pool := newTestPool(t)
	truncateAll(t, pool)

	r := NewFeedReaderPG(pool)

	tenant := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	user := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	// Insert two items with controlled timestamps
	t1 := time.Date(2026, 1, 22, 10, 0, 0, 0, time.UTC)
	t0 := t1.Add(-1 * time.Hour)

	insertInboxItem(t, pool,
		"11111111-1111-1111-1111-111111111111",
		tenant, user,
		"TASK_ASSIGNED", "UNREAD",
		"Newer", "Newer body", "https://x/newer",
		"22222222-2222-2222-2222-222222222222",
		"dedupe-newer",
		t1,
	)

	insertInboxItem(t, pool,
		"33333333-3333-3333-3333-333333333333",
		tenant, user,
		"TASK_ASSIGNED", "UNREAD",
		"Older", "Older body", "https://x/older",
		"44444444-4444-4444-4444-444444444444",
		"dedupe-older",
		t0,
	)

	// Page 1: limit 1 -> should return the "Newer" item
	page1, err := r.GetFeed(context.Background(), tenant, user, ports.FeedFilter{
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("GetFeed page1: %v", err)
	}
	if len(page1.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page1.Items))
	}
	if page1.Items[0].Title != "Newer" {
		t.Fatalf("expected 'Newer', got %q", page1.Items[0].Title)
	}
	if page1.NextCursor == nil {
		t.Fatalf("expected next cursor")
	}

	// Page 2: use cursor -> should return the "Older" item
	page2, err := r.GetFeed(context.Background(), tenant, user, ports.FeedFilter{
		Limit:  1,
		Cursor: page1.NextCursor,
	})
	if err != nil {
		t.Fatalf("GetFeed page2: %v", err)
	}
	if len(page2.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page2.Items))
	}
	if page2.Items[0].Title != "Older" {
		t.Fatalf("expected 'Older', got %q", page2.Items[0].Title)
	}
}

func TestFeedReaderPG_StatusFilter(t *testing.T) {
	pool := newTestPool(t)
	truncateAll(t, pool)

	r := NewFeedReaderPG(pool)

	tenant := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	user := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	now := time.Date(2026, 1, 22, 10, 0, 0, 0, time.UTC)

	// UNREAD item
	insertInboxItem(t, pool,
		"55555555-5555-5555-5555-555555555555",
		tenant, user,
		"TASK_ASSIGNED", "UNREAD",
		"Unread", "Body", "https://x/unread",
		"66666666-6666-6666-6666-666666666666",
		"dedupe-unread",
		now,
	)

	// READ item
	insertInboxItem(t, pool,
		"77777777-7777-7777-7777-777777777777",
		tenant, user,
		"TASK_ASSIGNED", "READ",
		"Read", "Body", "https://x/read",
		"88888888-8888-8888-8888-888888888888",
		"dedupe-read",
		now.Add(-time.Minute),
	)

	page, err := r.GetFeed(context.Background(), tenant, user, ports.FeedFilter{
		Status: "UNREAD",
		Limit:  50,
	})
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page.Items))
	}
	if page.Items[0].Title != "Unread" {
		t.Fatalf("expected 'Unread', got %q", page.Items[0].Title)
	}
}
