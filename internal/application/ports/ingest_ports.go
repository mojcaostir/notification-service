package ports

import "context"

type InboxWriter interface {
	InsertInboxItem(ctx context.Context, tx Tx, in InsertInboxItemParams) error
}

type InsertInboxItemParams struct {
	ID            string
	TenantID      string
	UserID        string
	Type          string
	Status        string
	Title         string
	Body          string
	ActionURL     string
	SourceEventID string
	DedupeKey     string
}

type EventDeduper interface {
	AlreadyProcessed(ctx context.Context, tx Tx, tenantID, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, tx Tx, tenantID, eventID string) error
}

type OutboxWriter interface {
	InsertOutbox(ctx context.Context, tx Tx, e OutboxEvent) error
}

type OutboxEvent struct {
	ID        string
	TenantID  string
	EventType string
	Payload   []byte // JSON
}
