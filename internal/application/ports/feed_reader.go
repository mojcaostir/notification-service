package ports

import (
	"context"
	"time"
)

type FeedItem struct {
	ID        string
	Type      string
	Status    string
	Title     string
	Body      string
	ActionURL string
	CreatedAt time.Time
}

type FeedCursor struct {
	CreatedAt time.Time
	ID        string
}

type FeedFilter struct {
	Status string // optional: "UNREAD", "READ", ...
	Limit  int
	Cursor *FeedCursor
}

type FeedPage struct {
	Items      []FeedItem
	NextCursor *FeedCursor
}

type FeedReader interface {
	GetFeed(ctx context.Context, tenantID, userID string, f FeedFilter) (FeedPage, error)
}
