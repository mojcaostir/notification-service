package main

import (
	"context"
	"log"
	"os"
	"time"

	apphttp "inbox-service/internal/infrastructure/http"
	"inbox-service/internal/application/queries"
	"inbox-service/internal/infrastructure/db"

	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.Background()

	dbURL := getenv("DB_URL", "postgres://inbox:inbox@localhost:5432/inbox?sslmode=disable")
	pool, err := db.NewPool(ctx, db.Config{URL: dbURL})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	feedReader := db.NewFeedReaderPG(pool)
	feedHandler := queries.NewFeedHandler(feedReader)
	handlers := apphttp.NewHandlers(feedHandler)

	e := echo.New()
	e.HideBanner = true

	apphttp.RegisterRoutes(e, handlers)

	addr := getenv("HTTP_ADDR", ":8080")
	srv := &httpServer{e: e, addr: addr}

	log.Printf("api listening on %s", addr)
	if err := srv.start(); err != nil {
		log.Fatalf("start: %v", err)
	}
}

// tiny wrapper so we can extend later with graceful shutdown
type httpServer struct {
	e    *echo.Echo
	addr string
}

func (s *httpServer) start() error {
	// Echo already handles context per request; add graceful later
	return s.e.Start(s.addr)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

var _ = time.Second // keep imported time for later shutdown improvements
