# Assignment 2 - gRPC Migration & Contract-First Development (Order & Payment)

**Course:** Advanced Programming 2  
**Student:** Muhammedali  
**GitHub:** mxkwnz  
**Deadline:** 12.04.2026

---

## Repository Links

| Repository | URL |
|---|---|
| Proto Files (Repo A) | https://github.com/mxkwnz/ap2-protos |
| Generated Code (Repo B) | https://github.com/mxkwnz/ap2-generated |
| Main Project | https://github.com/mxkwnz/Microservices-GO |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        CLIENT                               │
│                    (HTTP / Postman)                         │
└─────────────────────┬───────────────────────────────────────┘
                      │ REST (JSON)
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              ORDER SERVICE (port 8080)                      │
│                                                             │
│  REST Handler (Gin) → Use Case → Repository → order_db     │
│                                                             │
│  gRPC Server: SubscribeToOrderUpdates (stream) :9090        │
│  gRPC Client: calls Payment Service :9091                   │
└──────────────┬──────────────────────────┬───────────────────┘
               │ gRPC (ProcessPayment)    │ gRPC stream
               ▼                          ▼
┌──────────────────────────┐   ┌─────────────────────────────┐
│  PAYMENT SERVICE         │   │  SUBSCRIBER / FRONTEND      │
│  (port 8081 / 9091)      │   │  (stream-client)            │
│                          │   │                             │
│  gRPC Server             │   │  Receives real-time order   │
│  REST GET /payments/:id  │   │  status updates             │
│  Repository → payment_db │   └─────────────────────────────┘
└──────────────────────────┘
```

---

## Contract-First Flow

```
┌─────────────────┐     push      ┌──────────────────────────┐
│   Repo A        │ ──────────►  │   GitHub Actions          │
│   ap2-protos    │               │   (generate.yml)          │
│                 │               │   runs protoc             │
│  proto/         │               └──────────┬───────────────┘
│   payment/v1/   │                          │ auto-push
│   order/v1/     │                          ▼
└─────────────────┘               ┌──────────────────────────┐
                                  │   Repo B                  │
                                  │   ap2-generated           │
                                  │                           │
                                  │  payment/v1/*.pb.go       │
                                  │  order/v1/*.pb.go         │
                                  │  → tagged v1.0.2          │
                                  └──────────┬───────────────┘
                                             │ go get @v1.0.2
                                             ▼
                                  ┌──────────────────────────┐
                                  │  order-service           │
                                  │  payment-service         │
                                  └──────────────────────────┘
```

---

## What Changed from Assignment 1

| Component | Assignment 1 | Assignment 2 |
|---|---|---|
| Order → Payment | REST HTTP client | gRPC `ProcessPayment` |
| External API | REST (Gin) | REST (Gin) — unchanged |
| Streaming | None | `SubscribeToOrderUpdates` |
| Schema management | None | `.proto` + GitHub Actions |
| Configuration | Hardcoded strings | `.env` / environment variables |
| PaymentClient | HTTP client | gRPC client |

---

## Clean Architecture — Unchanged Layers

```
Handler (transport)   ← UPDATED: added gRPC transport
      ↓
Use Case              ← UNCHANGED
      ↓
Repository            ← UNCHANGED
      ↓
Domain                ← UNCHANGED
```

Only the **transport/ports layer** was updated. Business logic remains identical.

---

## How to Run

### Prerequisites
- Go 1.21+
- PostgreSQL running locally
- Databases: `order_db` and `payment_db`

### Step 1 — Run migrations

**order_db:**
```sql
CREATE TABLE orders (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    item_name TEXT NOT NULL,
    amount BIGINT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    idempotency_key TEXT UNIQUE
);
```

**payment_db:**
```sql
CREATE TABLE payments (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    transaction_id TEXT,
    amount BIGINT NOT NULL,
    status TEXT NOT NULL
);
```

### Step 2 — Configure environment variables

**payment-service/.env:**
```
DATABASE_URL=postgres://postgres:password@localhost:5432/payment_db?sslmode=disable
GRPC_ADDR=:9091
HTTP_ADDR=:8081
```

**order-service/.env:**
```
DATABASE_URL=postgres://postgres:password@localhost:5432/order_db?sslmode=disable
PAYMENT_GRPC_ADDR=localhost:9091
HTTP_ADDR=:8080
ORDER_GRPC_ADDR=:9090
```

### Step 3 — Start Payment Service first
```bash
cd payment-service
go run cmd/payment-service/main.go
```

### Step 4 — Start Order Service
```bash
cd order-service
go run cmd/order-service/main.go
```

---

## API Endpoints

### Order Service (REST) — port 8080
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/orders` | Create new order (triggers gRPC payment) |
| GET | `/orders/:id` | Get order by ID |
| PATCH | `/orders/:id/cancel` | Cancel order |
| GET | `/orders/revenue?customer_id=x` | Get revenue by customer |

### Payment Service (REST) — port 8081
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/payments/:order_id` | Get payment by order ID |

### gRPC Endpoints
| Service | Method | Type | Port |
|---------|--------|------|------|
| PaymentService | ProcessPayment | Unary | 9091 |
| OrderService | SubscribeToOrderUpdates | Server Streaming | 9090 |

---

## Testing

### Create an order
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "c-1", "item_name": "Laptop", "amount": 5000}'
```

Expected response:
```json
{
    "ID": "...",
    "CustomerID": "c-1",
    "ItemName": "Laptop",
    "Amount": 5000,
    "Status": "Paid",
    "CreatedAt": "..."
}
```

### Subscribe to real-time order updates (streaming)
```bash
cd stream-client
go run main.go <order-id>
```

Expected output:
```
Subscribed to order: <order-id>
[UPDATE] Status: Paid | Message: Status updated to Paid
```

---

## Proto Definitions

### PaymentService
```protobuf
service PaymentService {
  rpc ProcessPayment(PaymentRequest) returns (PaymentResponse);
}

message PaymentRequest {
  string order_id = 1;
  int64  amount   = 2;
}

message PaymentResponse {
  string transaction_id = 1;
  string status         = 2;
  google.protobuf.Timestamp processed_at = 3;
}
```

### OrderService (Streaming)
```protobuf
service OrderService {
  rpc SubscribeToOrderUpdates(OrderRequest) returns (stream OrderStatusUpdate);
}

message OrderRequest {
  string order_id = 1;
}

message OrderStatusUpdate {
  string order_id = 1;
  string status   = 2;
  string message  = 3;
}
```

---

## Generated Code Versions

| Version | Changes |
|---------|---------|
| v1.0.0 | Initial generation |
| v1.0.1 | Fixed go.mod go version |
| v1.0.2 | Final version with correct grpc compatibility |

---

## Git History Summary

- **Assignment 1 commits:** REST-based microservices with Clean Architecture
- **Assignment 2 commits:** gRPC migration, proto repos, streaming, .env config

---

## Evidence

1. GitHub Actions successful run — `ap2-generated` Actions tab
2. Postman `201 Created` — Order created with `Status: Paid` via gRPC
3. Stream client output — Real-time `[UPDATE] Status: Paid` from DB
4. Both services running — Payment gRPC on :9091, Order gRPC on :9090