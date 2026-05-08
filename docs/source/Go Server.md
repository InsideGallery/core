# Go Application Architecture Instruction

Architectural reference for any Go service following Modular Domain-Centric Architecture (MDCA) with Domain-Driven Design, Event-Driven Design, and microservices inside a monorepository.

---

## Table of Contents

1. [Modular Domain-Centric Architecture (MDCA)](#modular-domain-centric-architecture-mdca)
2. [Folder Structure](#folder-structure)
3. [Section Responsibilities](#section-responsibilities)
4. [Domain-Driven Design (DDD) in Go](#domain-driven-design-ddd-in-go)
5. [Event-Driven Design](#event-driven-design)
6. [Microservices in a Monorepository](#microservices-in-a-monorepository)
7. [Go Best Practices](#go-best-practices)
8. [How to Start a New Service](#how-to-start-a-new-service)

---

## Modular Domain-Centric Architecture (MDCA)

### Preamble

Why do we make an application? To solve business needs.

Software architecture is a tool that helps us reach those business needs. It is not about where to put files or how many interfaces to create. It is about efficiently solving a given business need.

The [Quality Goal](https://quality.arc42.org/home-new) is one of the Software Architecture characteristics we define for each initiative. Deciding on which architecture to choose is not something we decide once and forever for all applications. If the quality goals are efficiency, reliability, and maintainability, high coupling is acceptable. If the goals are flexibility, testability, and modifiability, it becomes a problem.

MDCA is about **balance and mindset**. It helps follow principles and balance different quality goals well. It keeps a similar structure for services with different architectures and even different technologies, helping build applications faster and avoid overcomplication.

### Purpose

Domain-centric architecture revolves around the business domain, modeling a company's actual workflows and behaviors. The architecture mirrors business operations and simplifies communication between development teams and business stakeholders.

The main principle of MDCA is to create an efficient and maintainable system based on **independent domain modules**. These modules solve exact business needs and allow easy code maintenance without increasing complexity. Each service should solve the exact problem using the actual technology.

MDCA separates infrastructure from domain modules while ensuring simplicity. However, some logic may live inside handlers because moving it behind an interface does not always make sense. The business does not change communication parties frequently. If it does, we usually must rewrite the entire service, and an additional abstraction layer does not help. That abstraction can cost more time ensuring compatibility than simply rewriting the component.

### Core Philosophy

- Business domain centricity -- model actual business workflows
- Domain-Driven Design (DDD) principles at the core
- Add functionality/flexibility only when it is needed
- Do not add abstraction before you need it

### Quality Goals

**Prefer:**
- Performance efficiency
- Reliability
- Maintainability

**Over:**
- Flexibility
- Transferability
- Compatibility

### Key Principles

| Principle | Description |
|-----------|-------------|
| **KISS** | Keep It Simple, Stupid |
| **DDD** | Domain-Driven Design |
| **DRY** | Don't Repeat Yourself |

### Advantages

- Easy to learn
- Suitable for many microservices inside one monorepository
- Fast to develop
- Easy to upgrade and maintain
- Perfectly fits microservices and event-driven architectures

### Disadvantages

- Housing multiple services in a single repository risks developers incorrectly importing internal packages across service boundaries

---

## Folder Structure

The monorepository splits into three main sections: **services** (entry points + service logic), **internal** (private implementation), and **pkg** (shared infrastructure).

### Multi-Service Monorepository

```
root/
├── services/                      # Independent microservices
│   ├── serviceA/                  # One service per business capability
│   │   ├── cmd/
│   │   │   └── main.go            # Entry point: config, DI, server start
│   │   └── internal/              # Private to this service
│   │       ├── app/
│   │       │   └── app.go         # Dependency injection container
│   │       ├── domain/            # Core business logic
│   │       │   ├── featureA/      # One package per bounded context
│   │       │   │   ├── model.go         # DB entities (db:"" tags only)
│   │       │   │   ├── port.go          # Repository/dependency interfaces
│   │       │   │   ├── service.go       # Business logic + DTO↔Model converters
│   │       │   │   ├── service_test.go  # Unit tests with mocks
│   │       │   │   └── mocks/           # Auto-generated mock implementations
│   │       │   └── featureB/
│   │       │       └── ...
│   │       ├── dto/               # Data Transfer Objects (json:"" tags only)
│   │       │   ├── featureA/
│   │       │   │   ├── request.go       # Incoming request DTOs
│   │       │   │   └── response.go      # Outgoing response DTOs
│   │       │   └── featureB/
│   │       │       └── ...
│   │       ├── repository/        # Data access implementations
│   │       │   ├── featureA.go    # No _repo suffix — package provides context
│   │       │   └── featureB.go
│   │       └── handler/           # Presentation layer (HTTP, gRPC, events)
│   │           ├── featureA.go    # Request handlers (uses DTOs, never models)
│   │           ├── meta.go        # Swagger/OpenAPI annotations
│   │           └── router.go      # Route registration
│   └── serviceB/
│       └── ...
├── pkg/                           # Shared reusable packages
│   ├── auth/                      # Authentication/authorization
│   ├── httpserver/                # HTTP framework wrapper, health checks
│   ├── postgres/                  # Database connection, DSN builder
│   ├── email/                     # SMTP client
│   ├── logger/                    # Structured logging
│   ├── messaging/                 # Event bus / message queue client
│   └── config/                    # Environment variable parsing
├── migrations/                    # Database migration files
├── docs/                          # Per-service documentation
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile
└── README.md
```

### Single-Domain vs Multi-Domain Services

**Single-domain services** (e.g., accessapi with 1 domain, notifyapi with 1 domain) use a **flat** structure — no feature subfolders:

```
services/accessapi/internal/
├── domain/           # Flat — model.go, port.go, service.go directly here
│   ├── model.go
│   ├── port.go
│   ├── service.go
│   └── service_test.go
├── dto/              # Flat — request.go, response.go directly here
│   ├── request.go
│   └── response.go
├── repository/
│   └── repository.go # Single file, no suffix needed
└── handler/
    ├── handler.go
    └── router.go
```

**Multi-domain services** (e.g., scheduleapi with 11 domains) use **feature subfolders**:

```
services/scheduleapi/internal/
├── domain/
│   ├── shift/        # One package per bounded context
│   ├── trade/
│   └── timeaway/
├── dto/
│   ├── shift/
│   ├── trade/
│   └── timeaway/
├── repository/
│   ├── shift.go      # No _repo suffix — package name provides context
│   ├── trade.go
│   └── timeaway.go
└── handler/
    ├── shift.go
    ├── trade.go
    ├── timeaway.go
    └── router.go
```

### Domain Internal Structure

Each domain feature directory follows a consistent pattern:

| File | Purpose |
|------|---------|
| `model.go` | DB entities with `db:""` struct tags only. **Never** `json:""` tags — those belong in DTOs. Domain-specific value objects and enums. |
| `port.go` | Interfaces (ports) that define what the domain needs from infrastructure. Typically a Repository interface. Includes `//go:generate` directives for mock generation. |
| `service.go` | Business logic + DTO↔Model converter functions (`NewXResponse`, `NewXFromRequest`). Orchestrates calls to ports and applies domain rules. Each public method represents a use case. |
| `service_test.go` | Unit tests using generated mocks. Tests business logic in isolation from infrastructure. |
| `mocks/` | Auto-generated mock implementations (mockgen, moq, etc.). |

### DTO Structure

DTOs live in a **separate** `dto/` directory (not inside domain):

| File | Purpose |
|------|---------|
| `request.go` | Incoming request types with `json:""` struct tags only. **Never** `db:""` tags. |
| `response.go` | Outgoing response types with `json:""` struct tags only. **Never** `db:""` tags. |

**Flow:** Handler → DTO → Service (converts to Model) → Repository. Handler never sees models. Repository never sees DTOs.

---

## Section Responsibilities

### cmd/main.go -- Entry Point

The entry point is **minimal**. It wires everything together and starts the server. It must **NOT** contain business logic.

Responsibilities:
- Parse configuration from environment variables
- Initialize infrastructure (database, message broker, logger)
- Create the DI container (`app.NewServices`)
- Register HTTP/gRPC routes or event subscriptions
- Start the server and handle graceful shutdown

### internal/app/app.go -- Dependency Injection Container

A plain struct that holds all domain services. A constructor function creates repositories, injects them into services, and returns the container. No framework needed -- Go structs and constructors are sufficient.

```go
type Services struct {
    Order   *order.Service
    Payment *payment.Service
}

func NewServices(db *sqlx.DB, bus messaging.Publisher) *Services {
    orderRepo := repository.NewOrderRepo(db)
    paymentRepo := repository.NewPaymentRepo(db)
    return &Services{
        Order:   order.NewService(orderRepo, bus),
        Payment: payment.NewService(paymentRepo),
    }
}
```

### internal/domain/{feature}/ -- Domain Layer

The heart of the application. Each feature (bounded context) gets its own package containing model, port, service, and tests.

**Domain packages must have ZERO infrastructure imports** -- no database drivers, no HTTP libraries, no message queue clients. They depend only on the Go standard library and their own port interfaces.

### internal/repository/ -- Data Access Layer

Implementations of `port.Repository` interfaces. Each file corresponds to one domain feature. Repositories own SQL queries, data mapping, and database-specific concerns. They accept `context.Context` for cancellation and return domain models (not ORM objects).

### internal/handler/ -- Presentation Layer

Request handlers that translate HTTP/gRPC/event inputs into domain service calls and format responses.

We do **NOT** split handler logic from validation or response formatting because of the KISS principle -- changing the communication layer usually changes how events are processed. We should not add flexibility to systems that do not change frequently.

| File | Purpose |
|------|---------|
| `router.go` | Route registration, middleware wiring |
| `meta.go` | Swagger/OpenAPI general info annotations |
| `{feature}.go` | Handler methods grouped by domain feature |

### pkg/ -- Shared Packages

Reusable infrastructure packages shared across services. Each package owns its configuration prefix and exposes a `GetEnvConfig()` function. Prefixes are by **FUNCTIONALITY** (e.g., `POSTGRES_`, `APP_`, `AUTH_`), not by service name.

**Rules for pkg/:**
- Must be genuinely reusable across 2+ services
- Must not import from any service's `internal/` package
- Must not contain business logic -- only infrastructure concerns
- Each package should be independently testable

---

## Domain-Driven Design (DDD) in Go

DDD organizes code around the business domain rather than technical layers. In MDCA, we apply DDD **pragmatically** -- use the patterns that help, skip those that add complexity without value.

### Bounded Contexts

Each domain feature package is a bounded context. It has its own models, its own language (ubiquitous language), and its own rules. **Do not share models across bounded contexts** -- if two domains need similar data, each defines its own struct.

```go
// domain/order/model.go -- Order's view of a product
type OrderItem struct {
    ProductID   string
    Quantity    int
    UnitPrice   decimal.Decimal
}

// domain/catalog/model.go -- Catalog's view of a product
type Product struct {
    ID          string
    Name        string
    Description string
    Price       decimal.Decimal
    Stock       int
}
```

### Entities and Value Objects

**Entities** have identity (an ID field) and a lifecycle. **Value objects** are defined by their attributes and are immutable. In Go, both are plain structs. Use pointer receivers for entity methods that mutate state; use value receivers for value object methods.

```go
// Entity -- has identity
type Order struct {
    ID        string
    Status    OrderStatus
    Items     []OrderItem
    CreatedAt time.Time
}

func (o *Order) Cancel() error {
    if o.Status != StatusPending {
        return ErrCannotCancel
    }
    o.Status = StatusCancelled
    return nil
}

// Value Object -- defined by attributes, immutable
type Money struct {
    Amount   decimal.Decimal
    Currency string
}

func (m Money) Add(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{Amount: m.Amount.Add(other.Amount), Currency: m.Currency}, nil
}
```

### Ports and Adapters (Hexagonal Architecture)

**Ports** are interfaces defined in the domain layer. **Adapters** are implementations in the repository or handler layer. The domain depends on abstractions, not concrete infrastructure.

```go
// domain/order/port.go
//go:generate go tool mockgen -source=port.go -destination=mocks/mock_repository.go
type Repository interface {
    GetByID(ctx context.Context, id string) (Order, error)
    Save(ctx context.Context, order Order) error
    ListByUser(ctx context.Context, userID string, filter Filter) ([]Order, error)
}
```

```go
// repository/order.go — no _repo suffix; package name provides context
type OrderRepo struct {
    db *sqlx.DB
}

func NewOrderRepo(db *sqlx.DB) *OrderRepo {
    return &OrderRepo{db: db}
}

func (r *OrderRepo) GetByID(ctx context.Context, id string) (order.Order, error) {
    const query = `SELECT id, status, created_at FROM orders WHERE id = $1`
    var o order.Order
    if err := r.db.GetContext(ctx, &o, query, id); err != nil {
        return order.Order{}, fmt.Errorf("get order by id: %w", err)
    }
    return o, nil
}
```

### Domain Services

Domain services contain business logic that does not naturally belong to a single entity. They orchestrate repositories and apply domain rules. Each public method is a use case.

```go
// domain/order/service.go
type Service struct {
    repo      Repository
    publisher messaging.Publisher
}

func NewService(repo Repository, pub messaging.Publisher) *Service {
    return &Service{repo: repo, publisher: pub}
}

func (s *Service) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (Order, error) {
    order := NewOrder(req.UserID, req.Items)
    if err := order.Validate(); err != nil {
        return Order{}, fmt.Errorf("validate order: %w", err)
    }
    if err := s.repo.Save(ctx, order); err != nil {
        return Order{}, fmt.Errorf("save order: %w", err)
    }
    s.publisher.Publish(ctx, "order.placed", OrderPlacedEvent{OrderID: order.ID})
    return order, nil
}
```

---

## Event-Driven Design

Event-driven design decouples services by communicating through events rather than direct calls. In MDCA, events are the primary mechanism for cross-service communication within the monorepository.

### Event Types

There are three kinds of events in an event-driven system:

| Type | Description | Example |
|------|-------------|---------|
| **Domain Events** | Something happened within a bounded context. Published by domain services after a state change succeeds. | `OrderPlaced`, `PaymentCompleted` |
| **Integration Events** | Domain events translated for cross-service consumption. Carry only the data other services need, not the full domain model. | `OrderPlacedIntegration` |
| **Command Events** | Requests for another service to perform an action. Used for async task delegation. | `SendNotification`, `ProcessPayment` |

### Event Structure

Events are plain Go structs. They carry enough context for consumers to act without calling back to the producer. Include correlation IDs for tracing.

```go
type Event struct {
    ID            string    // Unique event identifier
    Type          string    // Event type (e.g., "order.placed")
    Source        string    // Producing service name
    CorrelationID string    // Trace correlation
    OccurredAt    time.Time // When the event happened
    Payload       any       // Event-specific data
}

type OrderPlacedPayload struct {
    OrderID    string
    UserID     string
    TotalPrice decimal.Decimal
    ItemCount  int
}
```

### Publisher/Subscriber Pattern

Define publisher and subscriber as interfaces in `pkg/` so domain services depend on abstractions. Implementations can use NATS, Kafka, RabbitMQ, or even in-process channels -- the domain does not care.

```go
// pkg/messaging/publisher.go
type Publisher interface {
    Publish(ctx context.Context, topic string, event any) error
}

// pkg/messaging/subscriber.go
type Handler func(ctx context.Context, event []byte) error

type Subscriber interface {
    Subscribe(ctx context.Context, topic string, handler Handler) error
    Close() error
}
```

### Event Handling in Handlers

Event subscribers are registered alongside HTTP routes in the handler layer. Each handler translates the raw event into a domain service call.

```go
func RegisterEventHandlers(sub messaging.Subscriber, svc *app.Services) {
    sub.Subscribe(ctx, "order.placed", func(ctx context.Context, data []byte) error {
        var evt OrderPlacedPayload
        if err := json.Unmarshal(data, &evt); err != nil {
            return fmt.Errorf("unmarshal order.placed: %w", err)
        }
        return svc.Notification.SendOrderConfirmation(ctx, evt.UserID, evt.OrderID)
    })
}
```

### Eventual Consistency

Event-driven systems are eventually consistent. Each service owns its data and updates it based on events it receives. Do not rely on synchronous reads across service boundaries. If a service needs data from another, it maintains its own **local projection** updated via events.

### Idempotency

Event handlers **MUST** be idempotent -- processing the same event twice must produce the same result. Use the event ID to deduplicate. Store processed event IDs in a table or use database constraints to prevent double processing.

```go
func (s *Service) HandlePaymentCompleted(ctx context.Context, evt PaymentCompletedEvent) error {
    if s.repo.EventAlreadyProcessed(ctx, evt.ID) {
        return nil // idempotent -- already handled
    }
    // ... process event ...
    return s.repo.MarkEventProcessed(ctx, evt.ID)
}
```

### Outbox Pattern

To guarantee that database changes and events are published atomically, use the **transactional outbox pattern**: write the event to an outbox table in the same database transaction as the state change. A separate process polls the outbox and publishes events to the message broker.

```go
func (r *OrderRepo) SaveWithEvent(ctx context.Context, order Order, event Event) error {
    tx, err := r.db.BeginTxx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback()

    if _, err := tx.ExecContext(ctx, `INSERT INTO orders ...`, order); err != nil {
        return fmt.Errorf("insert order: %w", err)
    }
    if _, err := tx.ExecContext(ctx, `INSERT INTO outbox ...`, event); err != nil {
        return fmt.Errorf("insert outbox: %w", err)
    }
    return tx.Commit()
}
```

---

## Microservices in a Monorepository

A monorepository hosts multiple independently deployable services that share infrastructure code via `pkg/`. Each service is a self-contained unit with its own `main.go`, `internal/` tree, and deployment configuration.

### Service Boundaries

- Each service owns exactly one business capability
- Services **MUST NOT** import another service's `internal/` packages
- Services communicate via events, HTTP APIs, or gRPC -- never via direct function calls
- Shared code lives in `pkg/` and must be infrastructure-only (no business logic)

### Independent Deployability

- Each service compiles to its own binary
- Services can be deployed, scaled, and versioned independently
- A change in serviceA should not require redeploying serviceB
- Database migrations are per-service or shared via a controlled migration directory

### Data Ownership

- Each service owns its database tables
- No service reads or writes another service's tables directly
- Cross-service data access happens through APIs or event-driven projections

---

## Go Best Practices

Conventions derived from [Effective Go](https://go.dev/doc/effective_go) that apply to all Go code.

### Formatting

Use `gofmt` (or `go fmt`) to format all code. All code must pass `gofmt` without changes. Tabs for indentation. No manual alignment of comments -- `gofmt` handles it.

Opening braces must be on the same line as the control structure:

```go
// Correct
if err != nil {
    return err
}

// Wrong -- semicolon will be inserted before the brace
if err != nil
{
    return err
}
```

### Naming

**Package names:** single lowercase word, no underscores, no mixedCaps. Short, concise, and evocative. The package name is part of the qualified name -- avoid stuttering (e.g., use `order.Service`, not `order.OrderService`).

```go
package order     // good
package orderSvc  // bad -- mixedCaps
package order_svc // bad -- underscore
```

**Exported names:** PascalCase (MixedCaps). **Unexported names:** camelCase (mixedCaps). Never use underscores in Go names.

```go
type OrderStatus string  // exported
var defaultTimeout int   // unexported
```

**Getters:** do NOT use "Get" prefix. If the field is `owner`, the getter is `Owner()`, and the setter is `SetOwner()`.

```go
func (o *Order) Status() OrderStatus    { return o.status }  // getter -- no "Get"
func (o *Order) SetStatus(s OrderStatus) { o.status = s }    // setter -- "Set" prefix
```

**Interfaces:** one-method interfaces are named by the method plus an `-er` suffix. Use canonical names: `Reader`, `Writer`, `Closer`, `Formatter`, `Stringer`.

```go
type Publisher interface {
    Publish(ctx context.Context, topic string, event any) error
}
```

### Error Handling

Always check errors. Use the multi-value return idiom. Wrap errors with context using `fmt.Errorf` and `%w` for the wrapping verb. This preserves the error chain for `errors.Is` and `errors.As`.

```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("do something: %w", err)
}
```

Define **sentinel errors** for known failure conditions. Use **custom error types** when callers need to extract structured information.

```go
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s -- %s", e.Field, e.Message)
}
```

Do not panic in library or service code. Panic is reserved for truly unrecoverable situations (e.g., programmer errors during initialization). Use `recover` only in top-level middleware to prevent one request from crashing the process.

### Control Structures

Omit unnecessary `else` when the `if` body ends in `return`, `break`, `continue`, or `goto`. This reduces nesting and improves readability (the "happy path" stays left-aligned).

```go
// Good -- early return, no else
if err != nil {
    return err
}
// continue with happy path

// Bad -- unnecessary else
if err != nil {
    return err
} else {
    // happy path buried in else
}
```

Use initialization statements in `if` for scoped variables:

```go
if err := validate(req); err != nil {
    return fmt.Errorf("validate: %w", err)
}
```

Switch on `true` for cleaner if-else-if chains. No automatic fall-through -- use comma-separated cases instead.

### Functions

Use multiple return values, especially `(result, error)`. Name return parameters only when it improves documentation -- avoid bare returns in non-trivial functions.

Use `defer` for cleanup (closing files, releasing locks, finishing spans). Deferred calls execute in LIFO order. Arguments are evaluated when `defer` executes, not when the deferred function runs.

```go
func ReadConfig(path string) (Config, error) {
    f, err := os.Open(path)
    if err != nil {
        return Config{}, fmt.Errorf("open config: %w", err)
    }
    defer f.Close()
    // ... f is guaranteed to close on all return paths
}
```

### Interfaces

**Accept interfaces, return structs.** Define interfaces at the consumer side (in domain packages), not at the implementation side (in repository packages). Keep interfaces small -- one to three methods. Large interfaces are harder to implement and mock.

```go
// Defined where it is USED (domain), not where it is IMPLEMENTED (repository)
type Repository interface {
    GetByID(ctx context.Context, id string) (Order, error)
    Save(ctx context.Context, order Order) error
}
```

Use compile-time interface checks to verify implementations:

```go
var _ order.Repository = (*OrderRepo)(nil)
```

### Concurrency

> Do not communicate by sharing memory; share memory by communicating.

Use channels for coordination between goroutines. Use `sync.Mutex` only for protecting simple shared state within a single struct.

Goroutines are cheap but not free. Always ensure goroutines can terminate -- pass `context.Context` and select on `ctx.Done()`. Never launch goroutines that run forever without a shutdown path.

```go
func (s *Scheduler) Run(ctx context.Context) error {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            s.processJobs(ctx)
        }
    }
}
```

### Data Structures

Prefer `make` for slices, maps, and channels. Use `new` sparingly -- composite literals are more idiomatic. Design structs so the **zero value is useful**.

```go
users := make([]User, 0, 100)            // slice with capacity hint
index := make(map[string]int)             // initialized map
done := make(chan struct{})                // signal channel
cfg := &Config{Timeout: 30 * time.Second} // composite literal
```

Slices are references to underlying arrays. Passing a slice to a function does not copy the data. Arrays are values and are rarely used directly -- prefer slices.

### Commentary

Document all exported types, functions, and methods with doc comments. The comment begins with the name of the thing being documented. Line comments (`//`) are the norm.

```go
// Order represents a customer purchase with one or more line items.
type Order struct { ... }

// PlaceOrder validates and persists a new order, then publishes an OrderPlaced event.
func (s *Service) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (Order, error) {
```

Do not comment obvious code. Comments should explain **WHY**, not **WHAT**.

### Constants and Enums

Use `iota` for enumerated constants. Start with an explicit zero value or use blank identifier to skip zero.

```go
type OrderStatus string

const (
    StatusPending   OrderStatus = "pending"
    StatusConfirmed OrderStatus = "confirmed"
    StatusShipped   OrderStatus = "shipped"
    StatusCancelled OrderStatus = "cancelled"
)
```

### Init Functions

Use `init()` sparingly -- only for verifying program state or registering side effects (e.g., database drivers). Do not use `init()` for complex initialization that can fail at runtime. Prefer explicit initialization in `main()`.

### Embedding

Use struct embedding for **composition**, not inheritance. Embedded types promote their methods to the outer type. The receiver remains the inner type.

```go
type Base struct {
    ID        string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Order struct {
    Base
    Status OrderStatus
    Items  []OrderItem
}
```

### Testing

Test files live next to the code they test: `service_test.go` alongside `service.go`. Use **table-driven tests** for multiple scenarios. Use generated mocks for interface dependencies. Tag integration tests with build constraints.

```go
//go:build integration

func TestOrderRepo_GetByID(t *testing.T) { ... }
```

Unit tests must not depend on external systems (databases, APIs, message brokers). Integration tests verify adapter implementations against real infrastructure.

---

## How to Start a New Service

1. **Domain Analysis** -- identify the bounded context and its core use cases.
2. **Define Module Boundaries** -- one package per feature under `domain/`.
3. **Establish Domain Models** -- define DB entities in `domain/{feature}/model.go` with `db:""` tags only.
4. **Define DTOs** -- request/response types in `dto/{feature}/request.go` and `response.go` with `json:""` tags only.
5. **Define Ports** -- write Repository and other interfaces in `domain/{feature}/port.go`.
6. **Implement Domain Service** -- business logic + DTO↔Model converters in `service.go`, tests in `service_test.go`.
7. **Implement Repository** -- data access in `repository/{feature}.go` (no `_repo` suffix — package name provides context).
8. **Implement Handlers** -- HTTP/event handlers in `handler/{feature}.go` (bind DTOs, call service, return DTOs).
9. **Wire Everything** -- DI in `app/app.go`, routes in `handler/router.go`.
10. **Create Entry Point** -- minimal `main.go`: config -> infra -> DI -> serve.

### Real-World Implementation Tips

- **Start small** -- apply the architecture to a bounded area first
- **Evolve gradually** -- refactor existing systems incrementally
- **Focus on business value** -- prioritize domains most critical to the business
- **Continuous refactoring** -- revisit boundaries as domain understanding evolves
- **Do not add abstraction before you need it**
- **Keep the architecture solution behind the scenes** -- we should not always use an interface or follow SOLID if it will not improve the specific situation
- **Do not use an interface when the cost of abstraction exceeds the cost of rewriting**

### Example of High Interface Price

Consider a dynamic subscriber that entirely relies on a specific message queue implementation (e.g., NATS). Separating the subscriber from NATS to make it generic costs ~8 hours. But if we decide to use Kafka instead of NATS, rewriting the subscriber directly takes ~3 hours. The abstraction costs more than the rewrite -- avoid it.
