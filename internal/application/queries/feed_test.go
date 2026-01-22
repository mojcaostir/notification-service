package queries

import (
	"context"
	"testing"

	"inbox-service/internal/application/ports"
)

type fakeFeedReader struct {
	page ports.FeedPage
	err  error
}

func (f fakeFeedReader) GetFeed(ctx context.Context, tenantID, userID string, flt ports.FeedFilter) (ports.FeedPage, error) {
	return f.page, f.err
}

func TestFeedHandler_RequiresTenantAndUser(t *testing.T) {
	h := NewFeedHandler(fakeFeedReader{})

	_, err := h.Handle(context.Background(), FeedQuery{TenantID: "", UserID: "u"})
	if err == nil {
		t.Fatalf("expected error for missing tenant_id")
	}

	_, err = h.Handle(context.Background(), FeedQuery{TenantID: "t", UserID: ""})
	if err == nil {
		t.Fatalf("expected error for missing user_id")
	}
}
