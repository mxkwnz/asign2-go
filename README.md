# Clean Architecture Microservices (Order & Payment)

## 1. Architecture Decisions

This system is designed using **Clean Architecture** combined with **Microservices principles**.

### Why Clean Architecture?

Clean Architecture was chosen to:

* Separate business logic from infrastructure
* Improve maintainability and testability
* Allow independent evolution of components

Each service is structured into layers:

* **Domain** – core entities (Order, Payment)
* **Use Case** – business rules and workflows
* **Repository** – database access
* **Transport** – HTTP handlers
* **Main** – dependency injection

Dependency rule:

* Outer layers depend on inner layers
* Domain layer has no external dependencies

---

### Why Microservices?

The system is split into two independent services:

* Order Service
* Payment Service

This allows:

* Independent deployment
* Loose coupling
* Clear responsibility separation
* Scalability

---

### Why REST Communication?

REST was chosen because:

* Simple to implement
* Easy to debug (Postman)
* Suitable for synchronous workflows

---

### Why Database per Service?

Each service has its own database:

* Prevents tight coupling
* Ensures data ownership
* Allows independent scaling

---

## 2. Bounded Contexts

The system is divided into two **bounded contexts**:

### Order Context

Responsible for:

* Creating orders
* Managing order lifecycle
* Handling cancellation
* Communicating with Payment Service

Order states:

* Pending
* Paid
* Failed
* Cancelled

---

### Payment Context

Responsible for:

* Processing payments
* Validating payment limits
* Generating transaction IDs
* Storing payment results

Payment states:

* Authorized
* Declined

---

### Context Separation

* Order Service does NOT access Payment database
* Payment Service does NOT access Order database
* Communication only via HTTP

This ensures:

* Clear boundaries
* No shared models
* Proper microservice design

---

## 3. Failure Handling

The system is designed to handle failures gracefully.

### Scenario: Payment Service Unavailable

When Order Service calls Payment Service:

* A **timeout of 2 seconds** is applied
* If Payment Service does not respond:

    * Request is cancelled
    * Order is marked as **Failed**
    * HTTP **503 Service Unavailable** is returned

---

### Why this approach?

* Prevents system hanging
* Ensures fast failure detection
* Maintains system responsiveness

---

### Design Decision

Order is marked as **Failed** instead of staying Pending because:

* It avoids uncertainty
* Makes system state explicit
* Simplifies client-side handling

---

## 4. Idempotency

To prevent duplicate orders:

* The system uses `Idempotency-Key` header
* Requests with the same key return the same result
* Ensures safe retries in distributed systems

---

## 5. Architecture Diagram

```
Client
   ↓ HTTP
Order Service (Port 8080)
   - Domain (Order)
   - Use Case (Business Logic)
   - Repository (order_db)
   ↓ HTTP (REST)
Payment Service (Port 8081)
   - Domain (Payment)
   - Use Case (Payment Logic)
   - Repository (payment_db)
   ↓
PostgreSQL Databases
   - order_db
   - payment_db
```

---

## 6. Dependency Flow

```
Handler → UseCase → Repository → Database
             ↓
        Payment Client → HTTP → Payment Service
```

* Handlers are thin (only HTTP logic)
* Use cases contain business rules
* Repository handles persistence
* External communication goes through interfaces

---

## 7. Conclusion

This system demonstrates:

* Clean Architecture implementation
* Proper microservice decomposition
* Database per service pattern
* Reliable inter-service communication
* Robust failure handling
* Idempotent request processing

The design ensures scalability, maintainability, and clear separation of concerns.
