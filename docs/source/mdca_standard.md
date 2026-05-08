# Modular Domain-Centric Architecture (MDCA) — Standard

**Version:** 1.0
**Status:** Draft
**Editor:** Maksym Tkach (Frogo Team)
**Last updated:** 2026-04-29

---

## Abstract

This document defines **Modular Domain-Centric Architecture (MDCA)** — a software architecture standard for building business applications, optimized for Go services and clients but applicable to any modern statically-typed language. MDCA is a synthesis and refinement of three established schools of thought: **Domain-Driven Design (DDD)**, **Hexagonal Architecture (Ports and Adapters)**, and **Clean Architecture**. It distills their durable contributions, removes ceremony that does not pay for itself in practice, and prescribes a consistent module layout that scales across teams and microservices.

MDCA does not invent new ideas. It standardizes a pragmatic application of existing ones.

---

## 1. Status of This Document

This document is a working standard maintained by the Frogo Team. It uses the key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** as defined in [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119), to indicate requirement levels.

A codebase is **MDCA-conformant** if it satisfies all **MUST** clauses in §6 and §7. Codebases that satisfy the **SHOULD** clauses additionally are **MDCA-recommended**.

---

## 2. Scope

This standard applies to:
- Backend services written in statically-typed languages (primary target: Go).
- Client-side applications with a clear domain model (e.g., game clients, desktop apps, mobile apps with significant local logic).
- Reusable libraries that encapsulate a domain.

Out of scope:
- Pure data pipelines / ETL with no behavioral domain.
- One-off scripts and prototypes.
- UI-only frontends without local business logic.

---

## 3. Terminology

| Term | Definition |
|------|------------|
| **Domain** | The problem space — the business operations the software exists to support. |
| **Bounded Context** | A delimited region of the domain in which a single ubiquitous language and model apply consistently. |
| **Ubiquitous Language** | The vocabulary shared by domain experts and engineers, expressed identically in conversation, code, and documentation. |
| **Module** | An independently replaceable unit of code that owns its data and exposes a stable contract. In MDCA, the canonical module is one bounded context. |
| **Port** | An interface declared by the domain that names a capability it requires from infrastructure. |
| **Adapter** | A concrete implementation of a port; lives outside the domain. |
| **Composition Root** | The single location where concrete adapters are instantiated and injected into the domain. |
| **Domain Event** | A past-tense fact about something that has occurred in the domain, immutable after publication. |
| **Application Service** | A coordinator that translates external requests into domain operations and manages the boundaries of a single use case. |

---

## 4. Design Goals

MDCA-conformant systems prioritize, in order:

1. **Performance efficiency** — the architecture imposes no overhead the workload does not require.
2. **Reliability** — the system behaves predictably under stress and failure.
3. **Maintainability** — change in one place does not ripple into unrelated places.

When tension arises between these goals and *flexibility*, *transferability*, or *compatibility*, the goals above prevail unless an architectural exception is documented in the initiative's design record.

---

## 5. Relationship to Prior Art

MDCA inherits selectively from three sources. This section is informative.

### 5.1 Domain-Driven Design (Evans, 2003; Vernon, 2013)

**Adopted from DDD:**
- Strategic patterns: bounded contexts, ubiquitous language, context maps, subdomain classification (core / supporting / generic).
- Tactical patterns: entities, value objects, repositories, domain events.
- The premise that code is a model of the business and that the model must evolve with the business.

**Not adopted from DDD by default:**
- Aggregates with strict invariant enforcement (used **only** when concurrent invariant violation is a real risk).
- Specifications as object types.
- Factories as object types (constructor functions suffice).
- Domain services as a routine pattern (used only when an operation has a domain name and spans entities).

### 5.2 Hexagonal Architecture (Cockburn, 2005)

**Adopted from Hexagonal:**
- The principle that the domain depends on abstractions, never on infrastructure.
- Ports and adapters as the mechanism for that inversion.
- The composition root as the single wiring point.

**Refined:**
- MDCA does not distinguish "primary" from "secondary" ports terminologically; ports are simply interfaces declared by the domain.
- MDCA prescribes a concrete folder layout (§7.1) rather than leaving it implicit.

### 5.3 Clean Architecture (Martin, 2017)

**Adopted from Clean Architecture:**
- The Dependency Rule: source-code dependencies point inward toward the domain.
- The separation of policy (domain) from mechanism (infrastructure).
- The expectation that frameworks live at the edges, not at the center.

**Not adopted from Clean Architecture:**
- The four-ring diagram as a literal package layout. MDCA collapses the rings into three layers (Domain, Application, Infrastructure) and ties them to language-native units (Go packages).
- Use-case interactor classes as a mandatory layer. MDCA permits application services to be plain methods on a service struct when ceremony is unwarranted.
- Mandatory DTO-per-layer ("input boundary" / "output boundary"). MDCA requires DTO/model separation only between transport and domain.

### 5.4 Differential Summary

| Concern | DDD | Hexagonal | Clean Arch | **MDCA** |
|---------|-----|-----------|------------|----------|
| Strategic boundaries | Bounded contexts | Implicit | Implicit | Bounded contexts (mandatory) |
| Ubiquitous language | Central | Not addressed | Not addressed | Central |
| Dependency direction | Implied | Inward (ports) | Inward (rings) | Inward (explicit, lint-enforced) |
| Aggregates | Mandatory | Not addressed | Not addressed | **Optional**, justified by invariants |
| Use-case classes | N/A | N/A | Mandatory | **Optional**; methods on service structs default |
| DTO/model separation | Recommended | Recommended | Mandatory per ring | **Mandatory** between transport and domain only |
| Folder layout | Free | Free | Conceptual rings | **Prescribed** (§7.1) |
| Module unit | Bounded context | Hexagon | Component | Package = bounded context |
| External integration | ACL | Adapter | Gateway | **ACL adapter package** (mandatory for external SDKs) |
| Ceremony default | High | Medium | High | **Low** — patterns earn their place |

---

## 6. Principles (Normative)

A conformant codebase **MUST** satisfy all principles in this section.

### P1 — Domain Centricity

Code **MUST** be organized around business capabilities, not technical layers. Top-level groupings such as `controllers/`, `models/`, `services/` (where `services` is a flat dump) **MUST NOT** be the primary organizing structure. Instead, the primary organizing unit **MUST** be the bounded context.

### P2 — Ubiquitous Language

Types, methods, packages, and tests **MUST** use the vocabulary of domain experts. When the business term is awkward in code, the awkward term **MUST** be preferred over a translated euphemism. Generic names such as `Manager`, `Processor`, `Data`, `Info`, `Util` **MUST NOT** be used as type names within the domain.

### P3 — Dependency Inversion

Source-code dependencies **MUST** flow inward: presentation depends on application, application depends on domain. The domain layer **MUST NOT** import any infrastructure (database drivers, HTTP libraries, message brokers, vendor SDKs). Adapters **MUST** import the domain to satisfy its ports; the reverse is forbidden.

### P4 — Modular Independence

Each bounded context **MUST** be replaceable without modification to other bounded contexts. Cross-context type sharing **MUST NOT** occur — when two contexts model "the same" concept, each defines its own struct.

### P5 — Pragmatic Abstraction

Abstractions **MUST NOT** be introduced before a second concrete need exists. An interface with a single implementation **SHOULD** be deleted, with the concrete type used directly, until a second implementation appears.

### P6 — Explicit Composition

Concrete adapters **MUST** be instantiated in a single composition root (in Go services: `cmd/<svc>/main.go` and `internal/app/`). Service locators, global containers, and `init()`-based dependency wiring **MUST NOT** be used.

### P7 — Event-Driven Internal Communication

Within a system of services, cross-context communication **SHOULD** occur via domain events on a message broker (NATS in the reference implementation). Synchronous cross-context calls **MAY** be used where latency or atomicity demands them, but **MUST** be documented in the initiative's design record.

### P8 — Anticorruption at External Boundaries

External integrations (third-party SDKs, vendor APIs) **MUST** be encapsulated in adapter packages that translate between vendor types and domain types. Vendor types **MUST NOT** appear in domain signatures.

### P9 — DTO/Model Separation

Types crossing the transport boundary (HTTP, gRPC, events) **MUST** be distinct from domain entities and value objects. DTOs **MUST NOT** carry persistence tags (`db:""`); domain models **MUST NOT** carry transport tags (`json:""`, protobuf tags).

### P10 — Quality-Goal Discipline

When a design choice trades off the goals in §4 against flexibility/transferability/compatibility, the §4 goals **MUST** prevail unless a deliberate exception is recorded in the initiative's Architecture Inception Canvas.

---

## 7. Layering and Layout (Normative)

### 7.1 Layers

A conformant codebase **MUST** distinguish the following layers:

| Layer | Contents | Imports allowed |
|-------|----------|-----------------|
| **Domain** | Entities, value objects, ports, domain events, domain services | Standard library only; no infrastructure |
| **Application** | Composition root, dependency wiring, use-case orchestration when not collapsed onto domain methods | Domain, infrastructure adapters |
| **Infrastructure** | Repository implementations, brokers, external clients, ACLs | Domain (to satisfy ports), shared infra packages |
| **Presentation** | Handlers, DTOs, routers, UI input systems | Application, domain (read-only) |
| **Shared Infrastructure** | Cross-cutting utilities (logger, config, HTTP framework wrapper) | No business logic; no service-internal imports |

### 7.2 Reference Layout (Go Server)

```
services/<svc>/
  cmd/<svc>/main.go               # composition root entry
  internal/
    app/                          # DI container
    domain/<feature>/             # one bounded context per package
      model.go                    # entities, value objects (db:"" only)
      port.go                     # interfaces the domain requires
      service.go                  # use cases
      service_test.go
      events.go                   # domain events
    dto/<feature>/                # request.go, response.go (json:"" only)
    repository/<feature>.go       # adapter implementing port
    handler/<feature>.go          # presentation
pkg/                              # shared infrastructure
```

Single-domain services **MAY** flatten the `<feature>/` subdirectory.

### 7.3 Module Granularity

The canonical module unit is **one bounded context per package**. A bounded context **MUST**:

1. Own its data; expose access only via its public API or domain events.
2. Be deletable, rewritable, or replaceable without touching unrelated contexts.
3. Be testable without instantiating unrelated contexts.

A package that fails any of these tests is not a valid bounded context.

---

## 8. Tactical Rules (Normative)

### 8.1 Entities

- An entity **MUST** be identified by a typed ID field (e.g., `OrderID`, not bare `string`).
- Methods that mutate state **MUST** use pointer receivers; methods that read state **SHOULD** use value receivers.
- Entity methods **MUST** express domain operations (`Cancel`, `Approve`, `Reschedule`); naive setters (`SetStatus`) **MUST NOT** be exposed when they bypass invariants.
- Construction outside the entity's package **MUST** go through a constructor (`New<Type>`) that enforces all invariants.

### 8.2 Value Objects

- Value objects **MUST** be immutable after construction.
- Value objects **MUST** validate at construction; an invalid value object **MUST NOT** be representable.
- Domain concepts represented as primitives in transport (Money, Email, ID) **SHOULD** be wrapped in value objects within the domain.

### 8.3 Aggregates

- Aggregates **SHOULD NOT** be introduced unless a cross-entity invariant exists that can be violated under concurrent updates.
- When used, the aggregate **MUST** have one root entity, and external code **MUST** access inner entities only via the root.
- Aggregates **MUST** reference other aggregates by ID, not by pointer.

### 8.4 Repositories

- Repository interfaces **MUST** be declared in the domain package alongside the type they persist.
- Repositories **MUST** accept and return domain types. ORM rows, DB driver types, and DTOs **MUST NOT** appear in repository signatures.
- One repository **SHOULD** correspond to one aggregate root (or one entity, if no aggregate exists).
- Generic `Save(any)` repositories **MUST NOT** be used.

### 8.5 Domain Events

- Domain events **MUST** be named in the past tense (`OrderPlaced`, `PaymentCompleted`).
- Events **MUST** be published only after the corresponding state change is durably persisted. The transactional outbox pattern **SHOULD** be used when delivery guarantees matter.
- Event payloads **MUST** carry the minimum data consumers need; entire entities **MUST NOT** be embedded.
- Breaking changes to event shape **MUST** be expressed as new event types (`OrderPlacedV2`).

### 8.6 Application Services

- Application services **MAY** be implemented as methods on a domain service struct when the use-case has a single coordinator.
- Application services **MUST** be the only place where transactions are opened.
- Application services **MUST NOT** contain business invariants that belong on entities.

### 8.7 Anticorruption Layers

- Every external SDK or vendor API **MUST** be wrapped in an adapter package that exposes only domain types.
- The adapter package **MUST** be the only place that imports the vendor SDK.

---

## 9. Examples (Informative)

### 9.1 Entity with Invariants (Go)

```go
// domain/order/model.go
type OrderID string

type Order struct {
    ID        OrderID
    UserID    UserID
    Status    OrderStatus
    items     []OrderItem // unexported — only the entity mutates
    Total     Money
    PlacedAt  time.Time
}

func NewOrder(userID UserID, items []OrderItem) (Order, error) {
    if len(items) == 0 {
        return Order{}, ErrEmptyOrder
    }
    total, err := sum(items)
    if err != nil {
        return Order{}, err
    }
    return Order{
        ID:       newOrderID(),
        UserID:   userID,
        Status:   StatusPending,
        items:    items,
        Total:    total,
        PlacedAt: time.Now(),
    }, nil
}

func (o *Order) Cancel() error {
    if o.Status != StatusPending {
        return ErrCannotCancel
    }
    o.Status = StatusCancelled
    return nil
}
```

### 9.2 Port and Adapter (Go)

```go
// domain/order/port.go
type Repository interface {
    Get(ctx context.Context, id OrderID) (Order, error)
    Save(ctx context.Context, o Order) error
}

var ErrNotFound = errors.New("order not found")
```

```go
// internal/repository/order.go
type OrderRepo struct{ db *sqlx.DB }

func (r *OrderRepo) Get(ctx context.Context, id order.OrderID) (order.Order, error) {
    var o order.Order
    err := r.db.GetContext(ctx, &o, `SELECT ... WHERE id=$1`, id)
    if errors.Is(err, sql.ErrNoRows) {
        return order.Order{}, order.ErrNotFound
    }
    if err != nil {
        return order.Order{}, fmt.Errorf("get order %s: %w", id, err)
    }
    return o, nil
}
```

### 9.3 Composition Root (Go)

```go
// cmd/orderapi/main.go
func main() {
    cfg := config.Load()
    db := postgres.MustOpen(cfg.PG)
    bus := nats.MustConnect(cfg.NATS)

    repo := repository.NewOrderRepo(db)
    pub  := messaging.NewPublisher(bus)
    svc  := order.NewService(repo, pub)

    h := handler.NewOrder(svc)
    httpserver.Run(cfg.HTTP, h.Routes())
}
```

### 9.4 Domain Event with Outbox

```go
// domain/order/service.go
func (s *Service) Place(ctx context.Context, req PlaceRequest) (Order, error) {
    o, err := NewOrder(req.UserID, req.Items)
    if err != nil {
        return Order{}, err
    }
    err = s.tx.Run(ctx, func(ctx context.Context) error {
        if err := s.repo.Save(ctx, o); err != nil { return err }
        return s.outbox.Append(ctx, OrderPlaced{
            OrderID: o.ID, UserID: o.UserID, Total: o.Total,
        })
    })
    if err != nil {
        return Order{}, fmt.Errorf("place order: %w", err)
    }
    return o, nil
}
```

---

## 10. How to Adopt MDCA (Informative)

A team migrating to MDCA from a layered ("controllers/services/repositories") codebase **SHOULD** follow this sequence:

1. **Identify bounded contexts.** Listen for language shifts in stakeholder conversations. Group existing types by which stakeholder cares about them. Each group is a candidate context.
2. **Carve packages by context.** Move types into `domain/<context>/` packages. Resist the urge to share types across new packages.
3. **Extract ports.** For each external dependency the domain has (DB, broker, vendor SDK), declare an interface in the domain package.
4. **Move infrastructure outward.** Implement each port in a peer package (`internal/repository/`, `internal/messaging/`). Replace direct imports of `database/sql` etc. from the domain.
5. **Establish a composition root.** Centralize all `New*` calls in `cmd/<svc>/main.go` plus `internal/app/`.
6. **Separate DTOs from models.** Wherever a model is serialized for transport, introduce a DTO and a converter.
7. **Add a dependency-direction lint** (`go-arch-lint`, `depguard`) to prevent regression.
8. **Wrap external SDKs in ACLs.** No third-party type may appear in a domain signature.
9. **Audit ceremony.** Remove abstractions with one implementation. Remove specifications and factories that have no real load. Collapse use-case interactors onto service methods where appropriate.

---

## 11. Compliance Checklist

A reviewer **MUST** verify each item before declaring a change MDCA-conformant:

- [ ] All new types and methods use ubiquitous language. (P2)
- [ ] No domain package imports infrastructure or vendor SDKs. (P3, P8)
- [ ] No bounded context imports types from another bounded context. (P4)
- [ ] No interface introduced has only one implementation without a documented second use case on the horizon. (P5)
- [ ] All adapter wiring lives in the composition root. (P6)
- [ ] Cross-context communication uses events unless an exception is documented. (P7)
- [ ] Every external SDK is wrapped in an ACL package. (P8)
- [ ] DTOs carry only transport tags; models carry only persistence tags. (P9)
- [ ] Entity invariants are enforced at construction and via domain methods, not setters. (§8.1)
- [ ] Repository signatures use domain types, not driver or DTO types. (§8.4)
- [ ] Domain events are past-tense, persisted-then-published, with minimal payload. (§8.5)

---

## 12. Anti-Patterns (Informative)

The following patterns are non-conformant. Each row identifies the violated principle and the corrective action.

| Anti-Pattern | Violates | Fix |
|--------------|----------|-----|
| `domain/common/` package | P4 | Distribute types into the contexts that own them. |
| Domain importing `database/sql` | P3 | Define a port; move SQL to a repository. |
| Vendor SDK type in domain method signature | P8 | Wrap in ACL adapter; expose domain type. |
| `Manager`, `Processor`, `Util` type names in domain | P2 | Rename in business vocabulary. |
| Anemic entity (struct + getters/setters, no behavior) | P1 | Move behavior onto the entity; replace setters with operations. |
| Repository per database table | §8.4 | Repository per aggregate root. |
| Generic `Save(any)` repository | §8.4 | Type-specific repositories. |
| Event named `PlaceOrder` (imperative) | §8.5 | Rename to `OrderPlaced` (past tense). |
| Event published before persistence commits | §8.5 | Persist-then-publish; outbox pattern if guaranteed delivery is required. |
| Service locator / global container | P6 | Constructor injection from composition root. |
| Speculative interface with one implementation | P5 | Delete; reintroduce when a second implementation exists. |
| Model carrying both `db:""` and `json:""` tags | P9 | Split into model and DTO. |

---

## 13. Exceptions

A codebase **MAY** deviate from a **SHOULD** clause when a documented quality goal demands it. Deviations **MUST** be recorded in the initiative's Architecture Inception Canvas under *Architectural hypotheses* and reviewed at the next architecture review.

A codebase **MUST NOT** deviate from a **MUST** clause without explicit standard amendment.

---

## 14. References

- Eric Evans. *Domain-Driven Design: Tackling Complexity in the Heart of Software.* Addison-Wesley, 2003.
- Vaughn Vernon. *Implementing Domain-Driven Design.* Addison-Wesley, 2013.
- Alistair Cockburn. *Hexagonal Architecture.* alistair.cockburn.us, 2005.
- Robert C. Martin. *Clean Architecture: A Craftsman's Guide to Software Structure and Design.* Prentice Hall, 2017.
- Robert C. Martin. *Agile Software Development: Principles, Patterns, and Practices.* Prentice Hall, 2002. (SOLID origin.)
- IETF RFC 2119. *Key words for use in RFCs to Indicate Requirement Levels.* 1997.

---

## Appendix A — Glossary of Layer Imports (Go)

| From → To | Allowed? |
|-----------|----------|
| `cmd/<svc>` → `internal/app` | Yes |
| `cmd/<svc>` → `internal/repository` | Yes (composition only) |
| `internal/app` → `internal/domain/...` | Yes |
| `internal/handler` → `internal/domain/...` | Yes (read-only or via app) |
| `internal/repository` → `internal/domain/...` | Yes (to implement ports) |
| `internal/domain/...` → `internal/repository` | **No** |
| `internal/domain/...` → `internal/handler` | **No** |
| `internal/domain/<a>` → `internal/domain/<b>` | **No** (cross-context) |
| `internal/domain/...` → `database/sql`, `net/http`, vendor SDK | **No** |
| `internal/domain/...` → `pkg/...` (infra) | Discouraged; allowed only for pure utilities (e.g., `pkg/clock`) |
| `internal/repository` → `internal/handler` | **No** |
