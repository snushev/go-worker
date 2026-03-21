# go-worker

A PostgreSQL-backed job queue worker written in Go. Jobs are claimed atomically using `SELECT FOR UPDATE SKIP LOCKED`, ensuring safe concurrent processing with zero duplicate execution.

> 🚧 **Active development** — JSON payload parsing, job execution engine, multiple concurrent workers, and a REST API layer are all coming soon.

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  go-worker                       │
│                                                  │
│   ┌──────────┐     ┌──────────┐                  │
│   │ Worker 1 │     │ Worker N │  ← coming soon   │
│   └────┬─────┘     └────┬─────┘                  │
│        │                │                        │
│        └──────┬─────────┘                        │
│               ▼                                  │
│      ┌────────────────┐                          │
│      │  pgx/v5 Pool   │                          │
│      └───────┬────────┘                          │
│              │                                   │
└──────────────┼───────────────────────────────────┘
               ▼
     ┌──────────────────┐
     │   PostgreSQL      │
     │   jobs table      │
     └──────────────────┘
```

## How It Works

The worker runs a continuous polling loop:

1. Opens a transaction
2. Selects the oldest `pending` job using `FOR UPDATE SKIP LOCKED` — this prevents two workers from grabbing the same job
3. Marks it as `processing`
4. Commits and executes

If no jobs are available, the worker sleeps for 2 seconds before retrying.

## Database Schema

```sql
CREATE TABLE jobs (
    id         SERIAL PRIMARY KEY,
    payload    JSONB          NOT NULL,
    status     TEXT           NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
```

## Getting Started

**Prerequisites:** Go 1.21+, PostgreSQL 14+

```bash
git clone https://github.com/you/go-worker.git
cd go-worker

go mod tidy
```

Set your connection string in `main.go` (or via env variable — coming soon):

```go
dsn := "postgres://postgres:password@localhost:5432/go-worker"
```

Then run:

```bash
go run .
```

Insert a test job:

```sql
INSERT INTO jobs (payload) VALUES ('{"type": "email", "to": "user@example.com"}');
```

You should see:

```
Connected to Postgres successfully
JOB LOCKED & PROCESSING: {"type": "email", "to": "user@example.com"}
```

## Project Structure

```
go-worker/
├── main.go       # Worker loop — polls, locks, and processes jobs
├── database.go   # pgxpool connection setup
└── go.mod
```

## Roadmap

- [ ] **JSON payload parsing** — typed job structs with `encoding/json`
- [ ] **Job execution engine** — dispatch to handler functions based on job type
- [ ] **Multiple concurrent workers** — configurable worker pool via goroutines
- [ ] **REST API layer** — HTTP endpoints to enqueue jobs and check status
- [ ] **Retry logic** — exponential backoff for failed jobs
- [ ] **Dead letter queue** — jobs that exceed max retries moved to `failed`
- [ ] **Config via environment variables** — DSN, worker count, poll interval
- [ ] **Graceful shutdown** — drain in-flight jobs on SIGTERM

## Dependencies

| Package | Purpose |
|---|---|
| [`pgx/v5`](https://github.com/jackc/pgx) | PostgreSQL driver and connection pool |

## Why `FOR UPDATE SKIP LOCKED`?

Standard `SELECT` + `UPDATE` leaves a gap where two workers can read the same row before either updates it. `FOR UPDATE SKIP LOCKED` solves this at the database level — any row already locked by another transaction is skipped entirely, making the operation safe without application-level coordination.

---

*Feedbacks welcome.*
