CREATE TABLE IF NOT EXISTS inbox_items (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  user_id UUID NOT NULL,

  type TEXT NOT NULL,
  status TEXT NOT NULL,

  title TEXT NOT NULL,
  body TEXT NOT NULL,
  action_url TEXT NOT NULL,

  source_event_id UUID NOT NULL,
  dedupe_key TEXT NOT NULL,

  snooze_until TIMESTAMPTZ NULL,

  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,

  version INT NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_inbox_items_tenant_dedupe
  ON inbox_items (tenant_id, dedupe_key);

    CREATE INDEX IF NOT EXISTS ix_inbox_items_feed_status
  ON inbox_items (tenant_id, user_id, status, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS ix_inbox_items_feed_all
  ON inbox_items (tenant_id, user_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS processed_events (
  tenant_id UUID NOT NULL,
  event_id UUID NOT NULL,
  processed_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (tenant_id, event_id)
);

CREATE TABLE IF NOT EXISTS outbox (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  event_type TEXT NOT NULL,
  payload_json JSONB NOT NULL,

  status TEXT NOT NULL, -- PENDING | PROCESSING | SENT | FAILED
  attempts INT NOT NULL DEFAULT 0,
  next_run_at TIMESTAMPTZ NOT NULL,

  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,

  version INT NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS ix_outbox_pending
  ON outbox (status, next_run_at, created_at);
