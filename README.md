
# go-worker

A PostgreSQL-backed concurrent job queue system written in Go.

Jobs are processed asynchronously by multiple workers using `SELECT FOR UPDATE SKIP LOCKED`, ensuring safe concurrent execution with zero duplication.

---

## 🚀 Features

- Concurrent workers (goroutines)
- PostgreSQL-backed durable queue
- Atomic job locking (`FOR UPDATE SKIP LOCKED`)
- JSON-based job payloads
- Execution engine (print, sleep, extendable)
- REST API to enqueue jobs
- Clean, modular structure

---

## 🧠 Architecture

Client → HTTP API → PostgreSQL → Worker Pool → Execution

---

## ⚙️ How It Works

1. API receives a job request
2. Job is inserted into PostgreSQL (`pending`)
3. Workers poll database
4. Each worker:
   - Locks a job
   - Marks it as `processing`
   - Executes logic
   - Marks it as `done` or `failed`

---

## 🧩 Example Job

```json
{
  "task": "sleep",
  "seconds": 3
}
```

---

## 🌐 API

### POST /job

Create a job:

```json
{
  "task": "print",
  "value": "kobra"
}
```

---

### GET /jobs

Returns pending and failed jobs.

---

## 🧵 Concurrency

Workers run as goroutines:

- Parallel execution
- No duplicate jobs
- Safe locking via PostgreSQL

---

## 🗄 Database Schema

```sql
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## 📁 Project Structure

```
go-worker/
├── main.go        # entry point + worker pool
├── database.go    # DB connection
├── handlers.go    # HTTP API
├── models.go      # payload structs
```

---

## 🛣 Roadmap

- Retry logic
- Dead letter queue
- Config via ENV
- Graceful shutdown
- Metrics / logging

---

## 💡 Why This Project Matters

This is NOT a CRUD app.

It demonstrates:

- async processing
- concurrency
- system design
- real-world backend patterns

---

Built for learning, extensibility, and performance.

