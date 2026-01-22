package http

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, h *Handlers) {
	v1 := e.Group("/v1")
	v1.GET("/inbox/feed", h.GetFeed)

	// dev-only for now
	v1.POST("/dev/ingest/task-assigned", h.DevIngestTaskAssigned)
}
