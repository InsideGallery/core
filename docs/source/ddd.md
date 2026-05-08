# Domain-Driven Design in Go

**Source**: Eric Evans, *Domain-Driven Design: Tackling Complexity in the Heart of Software* (2003); refined by Vaughn Vernon, *Implementing Domain-Driven Design* (2013) and *Domain-Driven Design Distilled* (2016).
**Adaptation**: Go-idiomatic — plain structs, structural interfaces, package-as-bounded-context, no class hierarchies. This is the **canonical DDD reference** for the workspace; per-target specifics live in **Go Server.md**, **Go Client.md**, and **Go Library.md**. Aligned with MDCA (`mdca.md`), SOLID (`solid.md`), Clean Code, Engineering Principles (POL-ENG-001), and Twelve-Factor App.

---

## Philosophy

DDD organizes software around the **business domain**, not technical layers. The premise: most complex software fails not because of bad code but because the code does not faithfully model the business. Code that mirrors the domain is easier to change, easier to discuss with non-engineers, and easier to extend without breaking unrelated parts.

**Core beliefs:**
- The hardest part of software is **understanding the problem**. DDD invests in that understanding.
- Code is a **model** of the business. The model and the code must evolve together — when the business changes, the code does; when the code reveals a contradiction, the business model is wrong.
- The model is shared between developers and domain experts via **ubiquitous language** — the same terms in conversation, in code, in tests, in docs.
- DDD is a **toolkit, not a checklist**. Use the patterns that pay for themselves; skip the rest.

**MDCA's stance:** apply DDD **pragmatically**. The patterns that consistently earn their keep in this workspace are: bounded contexts, ubiquitous language, entities, value objects, domain events, repositories. Patterns that are more often ceremony than value: aggregates with strict invariant enforcement, specifications, factories, complex domain services. Adopt only on demand.

---

## Strategic DDD

Strategic DDD is about **what to model** and **where the boundaries go**. It is the most underused and most valuable half of DDD.

### 1. Ubiquitous Language

> The vocabulary of the domain experts becomes the vocabulary of the code.

If the business says "shift trade," the type is `ShiftTrade` — not `ScheduleSwapRequest`, not `ShiftExchangeTransaction`. If the business has two words that sound similar but mean different things ("schedule" the noun vs. "schedule" the verb), the code reflects both.

**Rules:**
- Use business terms for types, packages, methods, and tests. `order.Place(ctx, req)` not `order.Submit(ctx, req)` if the business says "place."
- One word per concept. Don't mix `Cancel` / `Void` / `Abort` for the same operation across packages.
- When the business term is awkward in code, **prefer the awkward term**. The cost of translation between conversation and code far exceeds the cost of an ugly identifier.
- When you cannot find a name in the domain, you don't understand the domain yet. Ask.
- Update the language continuously. When a domain expert corrects a term in conversation, update the code in the next PR.

**Anti-patterns:**
- Generic names: `Manager`, `Processor`, `Data`, `Info`, `Entity`. These reveal absence of domain understanding.
- Translation layers: a `BusinessOrderToInternalOrderMapper` is a smell — the model has split from the language.
- "Domain" as a module name. Every package in the domain is the domain. Name it after the bounded context.

### 2. Bounded Contexts

> A bounded context is the explicit boundary within which a particular model and its language apply consistently.

The same word can mean different things in different contexts. "Customer" in the Sales context is a prospect with discount eligibility; "Customer" in the Support context is an account with ticket history. **They are different types.** Forcing them to share a struct couples unrelated change reasons.

**Rules in MDCA:**
- One bounded context = one domain package (`domain/order`, `domain/billing`, `domain/scheduling`).
- Each context owns its types. Do not share `User`, `Product`, `Order` across contexts even when fields look identical.
- Cross-context communication: **events** (preferred, decoupled) or explicit **application service calls** (when synchronous coupling is justified).
- A bounded context is also a **team boundary candidate**. Conway's Law is real; align package boundaries with the people who own the domain.

**How to find boundaries:**
- Listen for **language shifts**. When a stakeholder says "well, in *our* world…" — that's a context boundary.
- Watch for **conflicting definitions** of the same noun. Two definitions ⇒ two contexts.
- Look at **change cadence**. Subsystems that change for the same reason at the same time belong in one context.

### 3. Context Map

> A diagram of how bounded contexts relate, and what that relationship costs.

Document context relationships explicitly in the initiative's AIC and (for bigger systems) in `togaf.md` or `arc42.md`. Common patterns:

| Pattern | Meaning | When |
|---------|---------|------|
| **Partnership** | Two contexts succeed or fail together; coordinated planning. | Tightly coupled teams, shared roadmap. |
| **Shared Kernel** | Two contexts share a small, jointly-owned model subset. | Avoid in MDCA — it couples change. Use only when duplication is genuinely worse than coupling. |
| **Customer / Supplier** | Upstream supplies, downstream consumes; downstream's needs shape upstream's API. | Most internal service relationships. |
| **Conformist** | Downstream conforms to upstream's model unchanged. | When upstream is external / unwilling to negotiate. |
| **Anticorruption Layer (ACL)** | Downstream wraps upstream behind an adapter that translates into the local language. | **Default for all external integrations.** Required when upstream's model leaks into your domain. |
| **Open Host Service / Published Language** | Upstream exposes a stable, documented protocol (REST, NATS subjects) for many consumers. | Public APIs, NATS event topics. |
| **Separate Ways** | Two contexts have no integration. | Default. Don't integrate without a reason. |

**Rule of thumb:** if you find yourself importing a third-party SDK type into your domain, you are missing an Anticorruption Layer.

### 4. Subdomains

A bounded context implements one (sometimes more) subdomain. Subdomain types help allocate engineering effort:

| Subdomain | Investment | Examples (per initiative) |
|-----------|------------|---------------------------|
| **Core** | Maximum DDD rigor. Custom-built. The thing the business is best at. | Scheduling logic in `scheduleapi`, access policy in `accessapi`. |
| **Supporting** | Custom but pragmatic. Less rigor, less depth. | Notifications, audit logging. |
| **Generic** | Buy or use a library. Wrap with an ACL. | Auth, payments (Stripe), email delivery (SES). |

> Spend modeling effort proportional to subdomain importance. A 200-line `audit` package with full DDD ceremony is wasted effort. The core scheduling rules deserve every value object you can give them.

---

## Tactical DDD

Tactical DDD is the inside of a bounded context — how the model is built. In Go, all tactical patterns reduce to **plain structs, methods, and interfaces**.

### Entities

> Identified by an ID. Mutable. Has a lifecycle (created, modified, archived).

```go
// domain/order/model.go
type Order struct {
    ID        OrderID
    UserID    UserID
    Status    OrderStatus
    Items     []OrderItem
    Total     Money
    CreatedAt time.Time
}

func (o *Order) Cancel() error {
    if o.Status != StatusPending {
        return ErrCannotCancel
    }
    o.Status = StatusCancelled
    return nil
}
```

**Rules:**
- ID is the identity. Two `Order`s with the same ID are the same order, even if other fields differ.
- Use **pointer receivers** for methods that mutate state. Use **value receivers** for read-only methods.
- ID types are themselves value objects: `type OrderID string` (or a UUID-wrapped struct). Avoid bare `string` parameters that allow mixing IDs across types.
- Methods on entities express **domain operations** (`Cancel`, `Approve`, `Reschedule`), not setters (`SetStatus`).
- Constructors enforce invariants: `NewOrder(...)` returns either a valid `Order` or an error. Never construct entities by zero-value initialization outside the package.

### Value Objects

> Identified by their attributes. Immutable. No lifecycle. Equality is structural.

```go
// domain/order/money.go
type Money struct {
    Amount   decimal.Decimal
    Currency string
}

func NewMoney(amount decimal.Decimal, currency string) (Money, error) {
    if currency == "" {
        return Money{}, ErrInvalidCurrency
    }
    return Money{Amount: amount, Currency: currency}, nil
}

func (m Money) Add(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{Amount: m.Amount.Add(other.Amount), Currency: m.Currency}, nil
}
```

**Rules:**
- All fields are read-only after construction. No setters. Mutating operations return a **new** value object.
- Use **value receivers**. The whole point is that copies are interchangeable.
- Equality is by-value (`==` or `reflect.DeepEqual` for structs with slices/maps).
- Validate at construction. If `NewMoney("", -1)` is invalid, the constructor returns an error; an invalid `Money` cannot exist.
- Prefer value objects over primitives for any domain concept: `Email`, `PhoneNumber`, `Currency`, `Percentage`, `OrderID`. Primitive obsession is a smell.

**Why value objects matter in Go:**
- Type safety: `func Charge(amount Money)` cannot accept a raw `decimal.Decimal`.
- Self-validating: by the time the value object exists, it's valid.
- Self-documenting: the type tells the reader what the value means.

### Aggregates (use sparingly)

> A cluster of entities and value objects treated as a single unit for consistency. The aggregate has one **root entity** that controls all access.

In MDCA, aggregates are the most frequently misapplied DDD pattern. Use them **only when**:
1. There is a real invariant that must hold across multiple entities atomically.
2. Concurrent updates can violate that invariant if checked separately.

If neither holds, just use entities and value objects directly.

```go
// domain/order/order.go — Order is the aggregate root
type Order struct {
    ID    OrderID
    items []OrderItem // unexported — only the root mutates
    total Money
}

func (o *Order) AddItem(productID ProductID, qty int, price Money) error {
    if len(o.items) >= 100 {
        return ErrTooManyItems // invariant: max 100 items per order
    }
    o.items = append(o.items, OrderItem{ProductID: productID, Quantity: qty, Price: price})
    o.recalculateTotal()
    return nil
}

// External code never accesses o.items directly. Read access via methods:
func (o *Order) Items() []OrderItem { return slices.Clone(o.items) }
```

**Rules:**
- One root per aggregate. External references point only to the root, never to inner entities.
- The root enforces all invariants. Inner entities are mutated only via root methods.
- Persistence loads and saves the **whole aggregate** as a unit. Repositories operate on aggregate roots.
- Keep aggregates small. A 12-entity aggregate is almost always wrong; split it.
- Reference other aggregates by ID, not by pointer. `Order.UserID UserID`, not `Order.User *User`.

**When NOT to use aggregates:**
- Read-only views — use a query model, not an aggregate.
- Entities with no cross-entity invariants — a plain entity is enough.
- "Just to organize the package" — that's what packages are for.

### Domain Services

> Stateless operations that do not naturally belong to a single entity or value object.

```go
// domain/pricing/calculator.go
type DiscountCalculator struct {
    rules []DiscountRule
}

func (c *DiscountCalculator) Apply(order order.Order, customer customer.Customer) Money {
    // logic that requires both Order and Customer; doesn't belong to either alone
}
```

**Rules:**
- Stateless. No persistence, no I/O.
- Named after a domain concept (`DiscountCalculator`, `RouteSelector`, `EligibilityChecker`).
- Distinct from **application services** (`order.Service`, `internal/app/...`), which orchestrate use cases, manage transactions, and coordinate adapters.

**MDCA note:** in most Go services, application service + entity methods are sufficient. A standalone domain service is justified only when the operation truly spans entities and has its own domain meaning.

### Repositories

> An interface for persistence. Defined by the domain, implemented by infrastructure.

```go
// domain/order/port.go
//go:generate mockgen -source=port.go -destination=mocks/mock_repository.go
type Repository interface {
    Get(ctx context.Context, id OrderID) (Order, error)
    Save(ctx context.Context, o Order) error
    ListByUser(ctx context.Context, userID UserID, f Filter) ([]Order, error)
}
```

**Rules:**
- Defined as an interface in the domain package.
- Returns and accepts **domain types**, never DTOs and never DB rows.
- Implemented in `internal/repository/` (server) or equivalent adapter package.
- One repository per **aggregate root**, not per table. If `Order` has line items, both load and save through the order repository.
- Hide query language. Domain code calls `repo.ListByUser(ctx, id, filter)`, not `repo.Query("SELECT ...")`.

**Sentinel errors at the boundary:**
```go
var ErrNotFound = errors.New("order not found")

// Implementations translate driver errors into domain errors.
func (r *OrderRepo) Get(ctx context.Context, id OrderID) (Order, error) {
    var o Order
    err := r.db.GetContext(ctx, &o, query, id)
    if errors.Is(err, sql.ErrNoRows) {
        return Order{}, ErrNotFound
    }
    if err != nil {
        return Order{}, fmt.Errorf("get order %s: %w", id, err)
    }
    return o, nil
}
```

### Domain Events

> A fact about something that happened in the domain. Past tense. Immutable.

```go
// domain/order/events.go
type OrderPlaced struct {
    OrderID    OrderID
    UserID     UserID
    Total      Money
    PlacedAt   time.Time
}
```

**Rules:**
- **Past tense** names: `OrderPlaced`, `PaymentCompleted`, `ShiftTraded`. Never `PlaceOrder` (that's a command).
- Published **after** the state change is persisted. Listeners react to facts that have already happened.
- Carry just enough data for consumers to act without round-tripping back. Resist the urge to dump entire entities.
- Events cross **bounded contexts** via NATS subjects. Use the standard subject naming convention (see Engineering Principles POL-ENG-001).
- Versioned at the wire format level. A breaking change is a new event (`OrderPlacedV2`), not a silent reshape.

**Outbox pattern (when delivery guarantees matter):**
- Persist the event to an `outbox` table in the same DB transaction as the state change.
- A separate publisher reads the outbox and publishes to NATS, marking rows shipped.
- Prevents "state change committed, event lost" and "event published, state change rolled back."

### Factories (rarely needed in Go)

In Go, a constructor function (`NewOrder`, `NewMoney`) is the factory. A separate `OrderFactory` type is justified only when:
1. Construction requires significant logic (lookups, multi-step assembly).
2. Multiple alternative construction paths benefit from grouping.

Otherwise: just write a `New...` function in the package.

### Specifications (almost never needed in Go)

A predicate object that encapsulates a business rule. In Go, a function literal or a small interface is almost always clearer:

```go
// Don't do this:
type EligibleForDiscountSpec struct{ /* ... */ }
func (s EligibleForDiscountSpec) IsSatisfiedBy(o Order) bool { /* ... */ }

// Do this:
func IsEligibleForDiscount(o Order) bool { /* ... */ }
```

Reach for specifications only if the predicate composes (`AND`, `OR`, `NOT`) with many others in a query DSL — and even then, prefer the database to express it.

---

## Layering: Where DDD Patterns Live

| Pattern | Layer (MDCA) | Go realization |
|---------|--------------|-----------------|
| Entity | Domain | `domain/<feature>/model.go` (struct + methods) |
| Value Object | Domain | `domain/<feature>/model.go` or own file |
| Aggregate Root | Domain | Same as entity, with unexported inner state |
| Domain Service | Domain | Stateless struct in `domain/<feature>/` |
| Repository (interface) | Domain | `domain/<feature>/port.go` |
| Repository (impl) | Infrastructure | `internal/repository/<feature>.go` |
| Domain Event | Domain | `domain/<feature>/events.go` |
| Application Service | Application | `domain/<feature>/service.go` (use cases) |
| Factory | Domain | `New<Type>` function in `domain/<feature>/` |
| ACL | Infrastructure | Adapter package wrapping external SDK |

> Domain code imports nothing from infrastructure. Infrastructure imports domain. Imports flow inward (see `mdca.md` Dependency Rule and `solid.md` DIP).

---

## DDD in Go vs. DDD in Java/C#

| Convention | Java/C# DDD | Go MDCA-DDD |
|------------|-------------|--------------|
| Class hierarchies | Common (abstract entity, base aggregate) | Avoided. Plain structs. Composition via embedding. |
| Interface position | Often in domain, with implementations elsewhere | Same — interfaces in domain, but **defined by the consumer** (small, role-based) |
| Annotations / decorators | Heavy (`@Entity`, `@Repository`) | None. Struct tags for marshaling only. |
| Reflection-driven mappers | ORMs everywhere | Explicit mapping in repositories. `sqlx`, `pgx`, hand-written. |
| Aggregate enforcement | Often via framework | Manual. Unexported fields + receiver methods. |
| Event dispatching | Spring events / MediatR | Explicit publisher injection; NATS subjects. |
| DDD ceremony | High | Low. Pattern earns its place. |

---

## Pragmatic Adoption Checklist

For any new bounded context, work through this list. Stop when patterns stop adding value.

1. ✅ **Ubiquitous language** — types and methods named in business terms.
2. ✅ **Bounded context** — one package, no cross-context type sharing.
3. ✅ **Entities + value objects** — primitives wrapped, IDs typed.
4. ✅ **Repository interface** — defined in domain, implemented in infra.
5. ✅ **Domain events** — published when state changes that other contexts care about.
6. ⚠️ **Aggregates** — only if there is a real cross-entity invariant under concurrent access.
7. ⚠️ **Anticorruption Layer** — required for any external SDK; don't let third-party types leak in.
8. ❓ **Domain service** — only if the operation has a name in the business but spans entities.
9. ❌ **Specifications** — almost never. Use functions.
10. ❌ **Factories as types** — almost never. Use `New...` constructors.

---

## Smells and Anti-Patterns

| Smell | Why it's wrong | Fix |
|-------|----------------|-----|
| Anemic domain model | Structs with public fields and zero methods; logic lives in services | Move behavior onto the entity. If the type has no methods, the model is anemic. |
| Primitive obsession | `func Transfer(from, to string, amount float64)` | Wrap in value objects: `AccountID`, `Money`. |
| Setters everywhere | `o.SetStatus(s)` invites invariant violations | Replace with domain operations: `o.Cancel()`, `o.Approve()`. |
| Shared types across contexts | Coupled change reasons; ripple on edit | Each context defines its own struct. |
| `domain/common/` package | Loss of bounded context boundaries | Delete; move types to the context that owns them. |
| Repository returning DTOs | Couples persistence to transport | Repositories return domain models. |
| Repository per table | Loses aggregate atomicity | One repository per aggregate root. |
| Generic `Save(any)` repository | Erases types and invariants | Type-specific repositories. |
| Domain importing `database/sql` | Inverted dependency | Define a port; move SQL to an adapter. |
| Events named in present tense (`PlaceOrder`) | Confuses commands with facts | Past tense (`OrderPlaced`). |
| Events carrying entire entities | Couples consumers to internal model | Carry only what consumers need. |
| Aggregate with 10+ entity types | Too coarse; transactions become contention | Split into smaller aggregates. |
| God service (`OrderService.DoEverything`) | SRP violation; tests need 8+ mocks | Split by use case; multiple narrow services or methods on the entity. |
| External SDK types in domain signatures | Missing ACL | Wrap upstream in an adapter that translates to local types. |

---

## DDD and the Workspace Methodologies

| Methodology | Interaction with DDD |
|-------------|----------------------|
| **MDCA** (`mdca.md`) | DDD provides the domain modeling vocabulary; MDCA defines the layering and module structure. They are co-designed in this workspace. |
| **SOLID** (`solid.md`) | DDD's repository pattern is DIP. ISP keeps repository interfaces small. SRP keeps bounded contexts focused. |
| **Clean Code** | Ubiquitous language is the foundation of intention-revealing names. |
| **Twelve-Factor App** | Domain stays unaware of config, backing services, and processes — those live at the edges. |
| **TBD** (`tbd.md`) | Bounded contexts are small enough to ship in 1–2 day branches; new behaviors land within their context behind feature flags. |
| **Event-Driven Design** | Domain events are the primary cross-context communication. Subjects follow naming convention (Engineering Principles). |
| **Arc42 / TOGAF** | Bounded contexts and context maps are first-class artifacts in the architecture documentation. |

---

## Quick Self-Check

Before merging a change to a domain package:

1. **Language** — Do all new types and methods use business terms a domain expert would recognize?
2. **Boundary** — Does this change stay within one bounded context, or does it leak across?
3. **Behavior on the model** — Do entities expose domain operations, or only data + getters/setters?
4. **Value objects** — Are domain concepts (Money, Email, IDs) typed, or are they raw primitives?
5. **Invariants at construction** — Can an invalid entity or value object be constructed? It must not.
6. **Repository scope** — Is the repository per aggregate root, returning domain models?
7. **Events** — Are events past-tense facts, published after persistence, carrying minimal payload?
8. **No leakage** — Does the domain import any infrastructure type? It must not.

If any answer is wrong, reshape the change before merging.

---

## When to Step Back from DDD

DDD is overhead. Skip it when:
- The "domain" is a thin wrapper over a CRUD form (admin tools, configuration UIs). Plain structs + handlers are fine.
- The subdomain is **generic** (auth, email, payments) — wrap a vendor with an ACL; don't model the vendor's domain yourself.
- The code is a **spike or prototype** — model just enough to validate the idea. Then either rewrite to DDD or throw it away.
- The "business rules" are entirely the database's (e.g., a reporting service that aggregates other services' data) — favor query models and read-side projections.

Document any deliberate skip in the initiative's AIC under *Architectural hypotheses* and (if applicable) the service's `CONTEXT.md`.

---

## Further Reading

- Eric Evans — *Domain-Driven Design: Tackling Complexity in the Heart of Software* (2003) — the original "blue book."
- Vaughn Vernon — *Implementing Domain-Driven Design* (2013) — the practical "red book."
- Vaughn Vernon — *Domain-Driven Design Distilled* (2016) — short overview.
- Eric Evans — *DDD Reference* (free PDF, domainlanguage.com) — pattern summary.
- Cross-references in this workspace: **mdca.md**, **solid.md**, **Go Server.md**, **Go Client.md**, **Go Library.md**, **Clean Code.md**, **Engineering Principles.md**, **Twelve-Factor App.md**, **tbd.md**.
