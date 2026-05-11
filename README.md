
## Assignment 4 — Caching, Background Jobs & External Integrations

### Invalidation Strategy

Cache-aside pattern is implemented in the Order Service using `go-redis`.

- **Read path**: `GET /orders/:id` checks Redis first. On a hit, the cached order is returned immediately without a DB query. On a miss, the DB is queried, and the result is stored in Redis with a 5-minute TTL (`CACHE_TTL_SECONDS=300`).
- **Write path (atomic invalidation)**: After any status update — payment completion, cancellation — the corresponding Redis key is deleted immediately. This ensures the next read always fetches fresh data from the DB and re-populates the cache.
- **Cache key format**: `order:<uuid>`

### Retry & Exponential Backoff

The `NotificationWorker` retries failed email sends with exponential backoff:

- Attempt 1 fails → wait **2s** → retry
- Attempt 2 fails → wait **4s** → retry
- Attempt 3 fails → wait **8s** → give up, NACK message for requeue

Backoff formula: `2^attempt seconds`. Configurable via `MAX_RETRIES` env var.

### Idempotency (Redis-backed)

Before sending a notification, the worker checks Redis for key `notification:sent:<event_id>`. If it exists with value `"sent"`, the event is skipped and the message is ACKed. On success, the key is written with a 24-hour TTL. This prevents duplicate emails even if RabbitMQ redelivers a message.

### Email Provider Adapter

The `EmailSender` interface decouples business logic from vendor implementation. The factory reads `PROVIDER_MODE`:

- `SIMULATED`: Logs the send, adds random latency (100–500ms), and randomly fails 30% of the time to test retry logic.
- `REAL`: Uses standard SMTP (configure via `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM`).

### Bonus: Redis Rate Limiter

A middleware on the Order Service uses Redis `INCR` + `EXPIRE` to count requests per client IP within a sliding window. Returns HTTP 429 when the limit is exceeded. Configurable via `RATE_LIMIT_REQUESTS` and `RATE_LIMIT_WINDOW_SECONDS`.    