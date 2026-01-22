package queries

import (
	"context"
	"fmt"

	"inbox-service/internal/application/ports"
)

type FeedQuery struct {
	TenantID string
	UserID   string
	Status   string
	Limit    int
	Cursor   *ports.FeedCursor
}

type FeedHandler struct {
	reader ports.FeedReader
}

func NewFeedHandler(reader ports.FeedReader) *FeedHandler {
	return &FeedHandler{reader: reader}
}

func (h *FeedHandler) Handle(ctx context.Context, q FeedQuery) (ports.FeedPage, error) {
	if q.TenantID == "" || q.UserID == "" {
		return ports.FeedPage{}, fmt.Errorf("tenant_id and user_id are required")
	}
	return h.reader.GetFeed(ctx, q.TenantID, q.UserID, ports.FeedFilter{
		Status: q.Status,
		Limit:  q.Limit,
		Cursor: q.Cursor,
	})
}
