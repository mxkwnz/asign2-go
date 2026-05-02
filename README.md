
---

### Idempotency Strategy

Each payment event published to RabbitMQ contains a unique `event_id`
field (a UUID generated at publish time, separate from the `order_id`).

When the Notification Service receives a message it follows this logic:

1. Parse the message body and extract `event_id`
2. Check an **in-memory map** (`map[string]struct{}`) for that ID
3. If the ID already exists → ACK the message silently, skip processing
4. If the ID is new → process it, add to the map, then ACK

This ensures that if RabbitMQ redelivers a message (e.g. the consumer
crashed before ACKing), the notification log is only printed once.

The map is protected by a `sync.Mutex` for safe concurrent access.

> **Trade-off**: The in-memory store is lost on service restart. For
> production, this would be replaced with a Redis SET or a `processed_events`
> DB table to survive restarts.

---

### ACK Logic

The consumer is started with `autoAck: false`:

```go
msgs, err := ch.Consume(queue, "", false, ...)
```

A message is acknowledged **only after** the notification log is
successfully printed:

```go
log.Printf("[Notification] Sent email to %s ...", event.CustomerEmail)
msg.Ack(false) // ACK only here
```

If JSON parsing fails (malformed / poison message), the message is
rejected without requeue so it does not block the queue:

```go
msg.Nack(false, false) // discard, no requeue
```

The queue is declared as **durable** and messages are published as
**Persistent** (`DeliveryMode: amqp.Persistent`), ensuring messages
survive a RabbitMQ broker restart.

**Delivery guarantee**: at-least-once. A message may be redelivered if
the consumer crashes between processing and ACKing — the idempotency
check above handles this case safely.

---

### Graceful Shutdown

All three services listen for `SIGINT` / `SIGTERM` via `os/signal`:

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
// close gRPC, HTTP, RabbitMQ connections cleanly
```

This ensures in-flight messages are not dropped when the service stops.