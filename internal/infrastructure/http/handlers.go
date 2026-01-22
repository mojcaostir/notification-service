package http

import (
	"net/http"
	"strconv"
	"time"

	"inbox-service/internal/application/ports"
	"inbox-service/internal/application/queries"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	Feed *queries.FeedHandler
}

func NewHandlers(feed *queries.FeedHandler) *Handlers {
	return &Handlers{Feed: feed}
}

// For now: tenant_id and user_id come from headers to keep it simple.
// Later: pull from JWT claims.
func (h *Handlers) GetFeed(c echo.Context) error {
	tenantID := c.Request().Header.Get("X-Tenant-Id")
	userID := c.Request().Header.Get("X-User-Id")

	status := c.QueryParam("status")
	limitStr := c.QueryParam("limit")

	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
		}
	}

	// Cursor as query params: cursor_created_at + cursor_id
	var cursor *ports.FeedCursor
	cc := c.QueryParam("cursor_created_at")
	ci := c.QueryParam("cursor_id")
	if cc != "" && ci != "" {
		t, err := time.Parse(time.RFC3339Nano, cc)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": "invalid cursor_created_at; must be RFC3339Nano"})
		}
		cursor = &ports.FeedCursor{CreatedAt: t, ID: ci}
	}

	page, err := h.Feed.Handle(c.Request().Context(), queries.FeedQuery{
		TenantID: tenantID,
		UserID:   userID,
		Status:   status,
		Limit:    limit,
		Cursor:   cursor,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	resp := map[string]any{"items": page.Items}
	if page.NextCursor != nil {
		resp["next_cursor"] = map[string]any{
			"created_at": page.NextCursor.CreatedAt.Format(time.RFC3339Nano),
			"id":         page.NextCursor.ID,
		}
	}
	return c.JSON(http.StatusOK, resp)
}
