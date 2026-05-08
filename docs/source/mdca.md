# Modular Domain-Centric Architecture (MDCA)

**Source**: Internal architectural method developed for this workspace's Go services and clients. Synthesizes Domain-Driven Design (Eric Evans, 2003), Clean Architecture (Robert C. Martin, 2017), Hexagonal/Ports-and-Adapters (Alistair Cockburn, 2005), and Event-Driven Design — adapted for Go's package model and the workspace's monorepository convention.
**Adaptation**: This document is the **canonical reference** for MDCA. Per-application specifics live in **Go Server.md**, **Go Client.md**, and **Go Library.md**. Aligned with Engineering Principles (POL-ENG-001), Clean Code, SOLID (`solid.md`), Twelve-Factor App, and Policy of Initiatives (POL-TECH-001).

---

## Preamble

Why do we make an application? **To solve business needs.**

Software architecture is a tool that helps us reach those business needs. It is not about where to put files, how many interfaces to create, or how cleverly we abstract. It is about **efficiently solving a given business need**.

The [Quality Goal](https://quality.arc42.org/home-new) is one of the architecture characteristics we define for each initiative. Choosing an architecture is not a one-time decision for all applications:

- If quality goals are **efficiency, reliability, maintainability** → some coupling is acceptable.
- If quality goals are **flexibility, testability, modifiability** → coupling becomes a problem.

MDCA is about **balance and mindset**. It keeps a similar structure across services with different architectures and even different technologies, helping teams build applications faster without over-engineering.

---

## Purpose

Domain-centric architecture revolves around the **business domain** — modeling a company's actual workflows and behaviors. The architecture mirrors business operations, simplifying communication between development teams and business stakeholders.

The main principle of MDCA is to create an **efficient and maintainable system based on independent domain modules** that solve exact business needs and allow easy maintenance without compounding complexity.

MDCA separates infrastructure from domain modules **while preserving simplicity**. Some logic may live inside handlers, repositories, or other adapters when introducing an interface offers no real value. The business does not change communication parties frequently; if it does, we usually rewrite the entire service, and an additional abstraction layer does not help. Speculative abstraction costs more time ensuring compatibility than rewriting the component when the change actually arrives.

> MDCA = "domain-first, pragmatic, modular." It is not Clean Architecture maximalism. It is not framework-driven hexagonal purity. It is the **minimum structure that makes a domain testable, replaceable at the edges, and aligned with the business**.

---

## Core Philosophy

1. **Business domain centricity** — model actual business workflows, not technical layers.
2. **Domain-Driven Design at the core** — bounded contexts, ubiquitous language, entities, value objects, domain events.
3. **Add functionality / flexibility only when it is needed** — YAGNI is not optional.
4. **Do not add abstraction before you need it** — the second concrete need justifies the interface, not the first.
5. **Independent modules** — each domain module is replaceable without touching the others.
6. **Predictable structure across services** — consistency across the monorepo is a feature, not a constraint.
7. **Edges are replaceable, the center is durable** — infrastructure changes; the domain endures.

---

## Quality Goals

| Prefer | Over |
|--------|------|
| Performance efficiency | Flexibility |
| Reliability | Transferability |
| Maintainability | Compatibility |

> When a quality goal is contested, prefer the left column. If a real, present-day need shifts the balance, document it in the initiative's AIC and proceed.

---

## Key Principles (priority order)

| # | Principle | What it means in MDCA |
|---|-----------|------------------------|
| 1 | **Performance by Design** | Choose data shapes, transports, and storage that fit the workload. Don't optimize prematurely; don't pessimize structurally. |
| 2 | **KISS** | Efficient, smart, readable — *not* simplistic. Solve the actual problem. |
| 3 | **DRY** | Each piece of business knowledge has one authoritative representation. Mechanical duplication across layers (DTO/model/db) is fine; **conceptual duplication** is not. |
| 4 | **MDCA** | Modular layout around bounded contexts. See **Building Blocks** below. |
| 5 | **DDD** | Applied pragmatically — bounded contexts, entities, value objects, domain events. Skip aggregates / repositories / specifications when they add ceremony without value. |
| 6 | **Clean Code** | Small functions, intention-revealing names, no commented-out code. See **Clean Code.md**. |
| 7 | **Event-driven internal communication** | NATS for cross-service events inside the monorepo. |
| 8 | **REST + backoff for external integrations** | Synchronous, resilient, well-versioned. |

---

## Building Blocks

MDCA structures every Go application around five conceptual layers. Their physical realization differs between **server**, **client**, and **library** (see the per-target documents), but the layering is the same.

### Layer 1 — Domain (Core)

The heart. Contains:
- **Models** (entities, value objects)
- **Domain services** (business logic, use cases)
- **Ports** (interfaces describing what the domain needs from infrastructure)
- **Domain events** (facts about state changes)

**Rules:**
- Zero imports of infrastructure (no `database/sql`, no `net/http`, no broker SDKs, no logger libraries except standard `log/slog` if needed for structured fields).
- Defines the interfaces it depends on (DIP — see `solid.md`).
- Pure Go, easily unit-testable, deterministic.

### Layer 2 — Application Wiring

Composition root. A small set of structs and constructors that instantiate domain services with concrete adapters injected.
- Server: `internal/app/app.go` — DI container.
- Client: bootstrap in `cmd/<client>/main.go` plus runtime registration of plugins/systems.
- Library: not applicable (libraries don't compose; consumers do).

### Layer 3 — Adapters (Infrastructure)

Implementations of the ports declared by the domain.
- **Repositories** — persistence (Postgres, Redis, file).
- **Publishers / subscribers** — event bus (NATS).
- **External clients** — REST, gRPC.
- **Notifiers** — email, push, SMS.

**Rules:**
- Adapters import the domain (to satisfy its interfaces). The domain does **not** import adapters. Imports point inward.
- One adapter package per external concern. No `util`/`common`/`misc`.

### Layer 4 — Presentation / Transport

Translates external inputs (HTTP requests, message envelopes, UI events) into domain service calls and formats responses.
- Server: handlers + DTOs.
- Client: input systems, UI controllers.
- Library: public API surface (`pkg/<lib>`).

**Rules:**
- Handlers receive DTOs, never models.
- Repositories return models, never DTOs.
- DTOs and models are **separate types**, even when fields look identical. They evolve on different cadences.

### Layer 5 — Cross-Cutting (Shared)

Infrastructure utilities that any service may use: logger, config loader, HTTP framework wrapper, DB connection helper.
- Lives in `pkg/` for the monorepo.
- Must be genuinely reused by ≥ 2 services.
- No business logic.

---

## The Dependency Rule

> Imports flow **inward**: presentation → application → domain. Adapters import domain. Domain imports nothing from outer layers. `pkg/` is sideways and may be imported by anything except domain (ideally even domain avoids it).

```
     ┌──────────────────────────────────────────┐
     │         Presentation / Transport         │
     │   handlers, DTOs, routers, UI systems    │
     └──────────────────────────────────────────┘
                       │ depends on ↓
     ┌──────────────────────────────────────────┐
     │         Application / Wiring             │
     │   DI container, composition root         │
     └──────────────────────────────────────────┘
                       │ depends on ↓
     ┌──────────────────────────────────────────┐
     │              DOMAIN (Core)               │
     │   models, services, ports, events        │
     │              (no infra imports)          │
     └──────────────────────────────────────────┘
                       ↑ depends on
     ┌──────────────────────────────────────────┐
     │           Adapters (Infra)               │
     │   repositories, publishers, ext. clients │
     └──────────────────────────────────────────┘
```

Enforce with package-import linters (`go-arch-lint`, `depguard`) where possible.

---

## Bounded Contexts

A **bounded context** is a clearly delimited slice of the domain in which one ubiquitous language applies. In MDCA, **one bounded context = one domain package** (e.g., `domain/order`, `domain/billing`, `domain/scheduling`).

**Rules:**
- Do not share models across bounded contexts. If `Order` and `Catalog` both reference a "product," each defines its own struct (`order.OrderItem`, `catalog.Product`).
- Cross-context communication happens via:
  - **Events** (preferred, asynchronous, decoupled) — published by the source context.
  - **Application service calls** (synchronous, when latency or atomicity demands it).
- Never reach into another bounded context's repository.

---

## Modules: What Counts as a Module?

MDCA's "M" — modular — defines the unit of independence:

| Scope | Module Unit | Replaceable Without Touching |
|-------|-------------|------------------------------|
| Server | A bounded context (`domain/<feature>`) + its adapters + its handlers | Other bounded contexts in the same service |
| Server | A whole microservice in the monorepo | Other services |
| Client | A subsystem (ECS system, plugin) | Other subsystems |
| Library | A package (`pkg/<name>`) | Other packages of the library |

A module is independent when:
1. It owns its data and exposes it only via its public API or via events.
2. It can be deleted, rewritten, or replaced without touching unrelated modules.
3. Its tests run without spinning up unrelated modules.

---

## Server, Client, Library — How MDCA Specializes

| Aspect | Go Server | Go Client | Go Library |
|--------|-----------|-----------|------------|
| **Composition root** | `cmd/<svc>/main.go` + `internal/app/app.go` | `cmd/<client>/main.go` + bootstrap | None — consumer composes |
| **Domain package** | `internal/domain/<feature>/` | `internal/<subsystem>/` (often ECS-shaped) | `pkg/<concept>/` |
| **Adapters** | `internal/repository/`, `internal/handler/`, NATS publishers | Plugins, drivers (graphics, audio, network) | None — library *is* an adapter to its consumer |
| **Communication** | NATS events internally; REST externally | Internal events; networking via the server | Pure function / type API |
| **Persistence** | Postgres, Redis | Local files, IndexedDB / SQLite | None |
| **Quality goals** | Reliability + maintainability dominate | Performance dominates (frame budget, input latency) | Stability of public API + minimal deps |
| **Reference doc** | `Go Server.md` | `Go Client.md` | `Go Library.md` |

MDCA is the same architecture in all three. Only the realization changes.

---

## Pragmatic DDD: What MDCA Adopts and What It Skips

**Adopt:**
- Bounded contexts → packages.
- Ubiquitous language → use business terms in code, even when they're awkward.
- Entities and value objects → plain Go structs.
- Domain events → published when state changes succeed.
- Repository interfaces → defined by the domain, implemented by infra.

**Skip unless they pay for themselves:**
- Aggregates (large object graphs with strict consistency boundaries) — only when concurrent invariant enforcement requires it.
- Specifications (predicate objects) — usually a function literal is clearer.
- Domain Services (in the DDD sense, distinct from application services) — usually collapses into the application service.
- Hexagonal "primary vs secondary ports" terminology — call them what they are.

> The test: if a DDD pattern does not improve testability, replaceability, or domain expressiveness for *this* initiative, do not adopt it.

---

## Inter-Module Communication

Within a service:
- **Direct call** when modules are part of the same use case and atomicity matters.
- **Domain event** when modules are reacting to a fact (loose coupling).

Across services (in the monorepo):
- **NATS events** — the default. Subjects follow the standard naming convention (see Engineering Principles POL-ENG-001).
- **REST** — only for synchronous, request/response semantics (UI-driven flows, external clients).

External integrations:
- **REST + exponential backoff + idempotency keys**.
- Wrap in a dedicated adapter package; never let a third-party SDK leak into the domain.

---

## Folder Structure (server, canonical)

See **Go Server.md** for full detail. The shape is:

```
services/<svc>/
  cmd/<svc>/main.go
  internal/
    app/                     # composition root
    domain/<feature>/        # one bounded context per package
      model.go               # entities, value objects (db:"" only)
      port.go                # interfaces the domain needs
      service.go             # business logic + use cases
      service_test.go
      mocks/
    dto/<feature>/           # request.go, response.go (json:"" only)
    repository/<feature>.go  # adapter for the domain's port
    handler/<feature>.go     # HTTP/gRPC/event presentation
pkg/                         # shared infrastructure (logger, config, postgres helper, ...)
```

Single-domain services use a flat layout (no `<feature>/` subdir). Multi-domain services use the subdirectory shape above. See **Go Server.md → Folder Structure**.

---

## Anti-Patterns

| Anti-Pattern | Why it breaks MDCA | Fix |
|--------------|--------------------|-----|
| Domain package imports `database/sql`, `net/http`, or a broker SDK | Inverts the dependency rule | Define a port; move infra to an adapter package. |
| `util`, `common`, `helpers`, `shared` packages | No bounded context = no responsibility | Move each helper to the package that owns the concept. |
| Shared model used by `domain/order` and `domain/billing` | Couples bounded contexts; one change ripples | Each context defines its own struct, even if fields look identical. |
| Handler reads from a model directly (no DTO) | Couples transport to persistence | Always pass through DTO converters in the service layer. |
| Repository returns DTOs | Same problem in reverse | Repositories return domain models. |
| `internal/services/` with 15 files of mixed concerns | Loss of bounded-context boundaries | Split by domain feature, each as its own package. |
| Cross-service import of `services/svcA/internal/...` from `svcB` | Violates microservice boundary inside the monorepo | Communicate via events; if data is needed, expose a stable contract. |
| Speculative interface in front of one implementation | Abstract before need | Delete the interface; reintroduce when a second implementation appears. |
| Big-ball-of-mud `app.go` that constructs 80 services | Composition root has become its own untested layer | Group by bounded context; introduce subcontainers per area. |
| "Domain service" that opens a DB transaction directly | Mechanism leaked into policy | Move transaction control to an application service or adapter. |
| Event published before the state change is persisted | Listeners react to a fact that may roll back | Persist first, then publish (transactional outbox if guarantee is needed). |

---

## MDCA, SOLID, and Clean Code Alignment

| Principle | MDCA realization |
|-----------|-------------------|
| **SRP** | One bounded context per package; one reason to change per type. |
| **OCP** | New behavior added by registering new adapters/subscribers, not editing existing services. |
| **LSP** | All adapters honor their port's contract; conformance test suites recommended. |
| **ISP** | Ports are small, role-based, defined by the consumer (the domain). |
| **DIP** | Domain defines interfaces; adapters depend on the domain. Imports flow inward. |
| **KISS / DRY** | One authoritative model per concept *within* a bounded context; mechanical layer-to-layer duplication is acceptable. |
| **Clean Code** | Stepdown ordering inside files; small functions; no dead code; intention-revealing names match ubiquitous language. |
| **Twelve-Factor** | Config at the composition root; logs to stdout; processes are stateless; backing services are attached resources via adapters. |
| **TBD** | Bounded contexts are small enough to ship in 1–2 day branches; new behaviors land behind feature flags inside their context. |

---

## Quick Self-Check

Before merging a change in an MDCA codebase:

1. **Bounded context** — Does this change belong entirely to one domain package? If it spans two, are they communicating via events or a clearly defined application service?
2. **Imports point inward** — Does the domain import any adapter or framework? It must not.
3. **Ports** — Did I add a new dependency? It should be expressed as an interface in the domain's `port.go`, with the implementation in an adapter package.
4. **DTO vs model** — Does any DTO carry `db:""` tags or any model carry `json:""` tags? Separate them.
5. **No util/common** — Did I add a helper somewhere generic? Move it to the package that owns the concept.
6. **Event semantics** — If I'm publishing an event, does it represent a fact that already happened (past tense, e.g., `OrderPlaced`), persisted before publish?
7. **Pragmatism** — Did I introduce an abstraction with only one implementation today? If yes, delete it until the second implementation arrives.

If any answer is wrong, stop and reshape the change before merging.

---

## When to Deviate from MDCA

MDCA is the default. Deviate consciously when:
- **Throughput-critical paths** require collapsing layers (e.g., zero-copy parsing directly into a transport buffer). Document the deviation in the initiative's AIC under *Architectural hypotheses*.
- **One-off scripts / data migrations** — flat `cmd/<tool>/main.go` is fine; full layering is overkill.
- **Spike / prototype** — skip ports and adapters until the spike validates the idea. Then either rewrite to MDCA or throw it away.
- **External vendor SDK constraints** force tight coupling to a framework. Isolate the coupling at one adapter; do not let it spread.

Document any deliberate deviation in the initiative's AIC and (if applicable) in the service's `CONTEXT.md`.

---

## Further Reading

- Eric Evans — *Domain-Driven Design: Tackling Complexity in the Heart of Software* (2003).
- Vaughn Vernon — *Implementing Domain-Driven Design* (2013).
- Robert C. Martin — *Clean Architecture* (2017), Chapters 17–22.
- Alistair Cockburn — *Hexagonal Architecture* (2005, alistair.cockburn.us).
- Cross-references in this workspace: **Go Server.md**, **Go Client.md**, **Go Library.md**, **Engineering Principles.md**, **Clean Code.md**, **solid.md**, **tbd.md**, **Twelve-Factor App.md**, **AIC Template.md**.
