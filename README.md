# 📥 Notification Service

## Table of Contents

- [Overview](#overview)
- [What this service does](#what-this-service-does)
- [Architecture at a glance](#architecture-at-a-glance)
- [Key design principles](#key-design-principles)
- [Tech stack](#tech-stack)
- [Getting started (local development)](#getting-started-local-development)
- [Current status](#current-status)
- [What comes next](#what-comes-next)
- [Learning goals & motivation](#learning-goals--motivation)

## Overview

**Notification Service** is a backend microservice that provides a **reliable, query-optimized user inbox** for enterprise applications.

It consumes **domain/integration events** (e.g. *TaskAssignedToUser*), converts them into **Inbox Items**, and exposes APIs for clients to **query and manage a user’s inbox** (feed, unread items, etc.).

The service is designed to demonstrate **senior-level backend architecture principles**:

* Domain-Driven Design (DDD)
* CQRS (separate write and read paths)
* Transactional Outbox
* Idempotent event ingestion
* Event-driven integration
* Testability and operational readiness

---

## What this service does

### Core responsibilities

* Ingest integration events (at-least-once delivery)
* Create inbox items in a **transactionally safe** way
* Guarantee **idempotency** (no duplicate inbox items)
* Provide fast, cursor-based inbox feed queries
* Emit integration events via an **outbox** (reliable publishing)

### What it explicitly does NOT do

* Send emails or push notifications
* Own user, tenant, or task data
* Perform synchronous calls to other services

External data is handled via **event-carried state transfer** (local snapshots when needed).

---

## Architecture at a glance

```
                ┌──────────────┐
                │ Event Source │
                │ (e.g. Tasks) │
                └──────┬───────┘
                       │
                 Integration Event
                       │
              ┌────────▼────────┐
              │ Inbox Consumer  │
              └────────┬────────┘
                       │
            ┌──────────▼──────────┐
            │ Transaction (DB)    │
            │ - inbox_items       │
            │ - processed_events │
            │ - outbox            │
            └──────────┬──────────┘
                       │
         ┌─────────────▼─────────────┐
         │ HTTP Query API (CQRS Read) │
         │ /v1/inbox/feed             │
         └───────────────────────────┘
```

---

## Key design principles

### 1. CQRS

* **Write path**: event ingestion + business rules + transactions
* **Read path**: optimized SQL queries returning DTOs
* No domain logic in query handlers

### 2. Idempotency

* Every inbound event has an `event_id`
* `processed_events` table guarantees safe reprocessing
* Unique constraints protect against race conditions

### 3. Transactional Outbox

* Inbox item + processed marker + outbox event are written **in one DB transaction**
* Prevents dual-write problems
* Outbox events are published asynchronously

### 4. Testability

* Application logic depends on **ports (interfaces)**
* Infrastructure code is isolated
* Unit tests + real Postgres integration tests

---

## Tech stack

* **Language**: Go
* **HTTP**: Echo
* **Database**: PostgreSQL
* **DB Driver**: pgx
* **Containerization**: Docker / docker-compose
* **Testing**: Go test + real DB integration tests

(Event Hub and OpenTelemetry are added in later steps.)

---

## Getting started (local development)

### 1. Prerequisites

* Go 1.21+
* Docker + Docker Compose
* `psql` (optional, for debugging)

---

### 2. Start Postgres

```bash
docker-compose up -d
```

This will:

* start Postgres
* automatically create all tables via init scripts

⚠️ If schema changes don’t apply, reset volumes:

```bash
docker-compose down -v
docker-compose up -d
```

---

### 3. Run the API

```bash
go run ./cmd/api
```

The service will start on:

```
http://localhost:8080
```

---

### 4. Seed data (manual, for development)

You can insert a test inbox item via `psql`:

```bash
psql postgres://inbox:inbox@localhost:5432/inbox
```

```sql
INSERT INTO inbox_items (
  id, tenant_id, user_id,
  type, status,
  title, body, action_url,
  source_event_id, dedupe_key,
  created_at, updated_at, version
) VALUES (
  '11111111-1111-1111-1111-111111111111',
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
  'TASK_ASSIGNED',
  'UNREAD',
  'Task assigned to you',
  'Prepare quarterly report',
  'https://app.example.com/tasks/42',
  '22222222-2222-2222-2222-222222222222',
  'TASK_ASSIGNED:42:bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
  now(), now(), 1
);
```

---

### 5. Query the inbox feed

```bash
curl -H "X-Tenant-Id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" \
     -H "X-User-Id: bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" \
     http://localhost:8080/v1/inbox/feed
```

---

### 6. Run tests

Make sure Postgres is running, then:

```bash
go test ./... -count=1
```

Tests include:

* unit tests for handlers
* integration tests using real Postgres
* idempotency guarantees

---

## Current status

✅ Inbox feed (CQRS read path)

✅ Cursor-based pagination

✅ Idempotent ingest handler

✅ Transactional outbox writes

✅ Integration tests with Postgres

---

## What comes next

* Outbox dispatcher worker (publish + retry + backoff)
* Event publisher abstraction
* Azure Event Hubs integration
* OpenTelemetry tracing & metrics
* Multi-consumer safety & scaling

---

Perfect. Here is a **clean, professional README section** you can paste directly, written in a way that signals **intentional learning**, not “toy project”.

---

## 🎯 Learning goals & motivation

This service is developed **explicitly as a learning project** to deepen my understanding of **system design** and **production-grade microservice patterns**.

Rather than focusing on feature completeness, the goal is to **intentionally practice and internalize architectural principles** that are commonly expected in real-world, large-scale backend systems.

### What I wanted to learn
#### 1. How to design a microservice using DDD boundaries

* Defining a **clear bounded context**
* Separating domain concepts from infrastructure concerns
* Making business rules explicit instead of implicit SQL logic

#### 2. CQRS in practice

* Separating **write-side use cases** from **read-side queries**
* Designing read models optimized for API access

#### 3. Reliable event ingestion (at-least-once delivery)

* Handling duplicate events safely
* Designing **idempotent consumers**
* Understanding real-world failure modes of event-driven systems

#### 4. Transactional Outbox pattern

* Solving the dual-write problem (DB + message broker)
* Writing state changes and integration events **in one transaction**
* Preparing the system for reliable asynchronous publishing

#### 5. Event-driven integration without tight coupling

* Avoiding synchronous calls to other services
* Using **event-carried state transfer** instead of remote lookups
* Designing schemas and versioning strategies for integration events

#### 6. Testable backend architecture

* Designing application logic around **ports (interfaces)**
* Writing unit tests for use cases without infrastructure
* Writing **integration tests** against Postgres
* Verifying idempotency and transactional behavior with tests


---
