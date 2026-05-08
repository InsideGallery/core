# Arc42 Architecture Documentation

**Source**: [arc42.org](https://arc42.org) - Template Version 9.0, by Dr. Peter Hruschka & Dr. Gernot Starke
**Purpose**: Instruction for generating full arc42 architecture documentation. This is the comprehensive version (12 sections). For lightweight inception use the AIC Template.

---

## When to Use

| Document | When | Sections |
|----------|------|----------|
| **aic.md** (AIC Canvas) | Every initiative -- quick inception document | Business Case, Quality Goals, Constraints, Business Context, Hypotheses, Risks, Tasks |
| **arc42.md** (Full Arc42) | Complex/critical systems needing comprehensive architecture documentation | All 12 sections below |
| **togaf.md** (TOGAF ADM) | Enterprise-wide or compliance-heavy initiatives | See `source/TOGAF.md` |

Generate `arc42.md` when the CTO explicitly requests it, or recommend it when:
- The system has multiple interacting services/components
- Multiple teams will build/maintain the system
- Deployment is non-trivial (multi-region, hybrid, complex infrastructure)
- Long-term maintainability and onboarding matter

---

## Template Structure

### 1. Introduction and Goals

Describes the driving forces behind the architecture. Three subsections:

#### 1.1 Requirements Overview

Short description of the key functional requirements. Link to external requirements documents if they exist.

**What to write**: 5-10 most important requirements as a prioritized list or use-case table. Not the full backlog -- just the architecturally significant ones.

#### 1.2 Quality Goals

The top 3-5 quality goals for the architecture (not project goals). Use ISO 25010 categories:

| Category | Examples |
|----------|---------|
| Performance Efficiency | Response time, throughput, resource utilization |
| Reliability | Availability, fault tolerance, recoverability |
| Security | Confidentiality, integrity, authentication |
| Maintainability | Modularity, reusability, modifiability, testability |
| Portability | Adaptability, installability |
| Compatibility | Interoperability, co-existence |
| Usability | Learnability, operability, accessibility |
| Functional Suitability | Correctness, completeness, appropriateness |

**Format**: Table with quality goal + concrete scenario (measurable).

```markdown
| Quality Goal | Scenario |
|-------------|----------|
| Low latency | API responds within 50ms at p99 under 1000 RPS |
| High availability | System remains operational with 99.9% uptime (8.7h downtime/year) |
| Modifiability | New payment provider integrates in < 1 week |
```

#### 1.3 Stakeholders

Table: Role/Name, Contact, Expectations with respect to architecture.

---

### 2. Architecture Constraints

Constraints that limit architectural freedom. Three categories:

| Category | Examples |
|----------|---------|
| **Technical** | Must use Go, must run on Kubernetes, must use NATS for internal messaging, must follow Engineering Principles (POL-ENG-001) |
| **Organizational** | Team size, timeline, budget, skill availability |
| **Conventions** | API versioning `/v1/`, branch naming `epic/{TASK}_{TITLE}`, Twelve-Factor compliance |

**Format**: Simple table with constraint and explanation.

---

### 3. Context and Scope

Delimits the system from its environment. Two views:

#### 3.1 Business Context

The system as a **black box**. Shows all communication partners (users, external systems) and what data/messages flow in and out. Focus on **domain-level** inputs and outputs.

**Diagram**: C4 Context diagram (DOT format). Include:
- All human actors
- All external systems
- All data flows with labels describing the domain content

#### 3.2 Technical Context

Maps domain interfaces to technical channels: protocols (REST, gRPC, NATS), formats (JSON, Protobuf), hardware (load balancer, VPN, CDN).

**Diagram**: Deployment-aware context diagram showing protocols and channels.

---

### 4. Solution Strategy

Fundamental decisions and solution approach. Short and strategic:

- Technology decisions (language, framework, database)
- Top-level decomposition (monorepo, microservices, monolith)
- Quality goal strategies (how each goal from 1.2 is addressed)
- Organizational decisions (team structure, development process)

**Format**: 4-8 bullet points or a table mapping quality goals to solution approaches. Refer to details in later sections.

---

### 5. Building Block View

**Static decomposition** of the system. Hierarchical black-box / white-box refinement.

#### Level 1: Overall System (White Box)

- Overview diagram (C4 Container or component diagram)
- Motivation for the decomposition
- Black box descriptions of each contained building block

**Black box description** (per block):

| Field | Description |
|-------|-------------|
| Purpose/Responsibility | What it does |
| Interface(s) | APIs, NATS subjects, events consumed/produced |
| Quality/Performance | SLA, throughput, latency |
| Directory/File Location | Path in the monorepo (if applicable) |
| Open Issues/Risks | Known problems |

#### Level 2: Zoom Into Important Blocks

White-box descriptions of selected Level 1 blocks. Same structure, one level deeper.

#### Level 3+ (if needed)

Further zoom into complex subsystems. Use sparingly -- prefer relevance over completeness.

**Go-specific**: Map to monorepo structure from Go Server.md:
- Level 1 = services in `cmd/`
- Level 2 = internal packages per service in `internal/<service>/`
- Level 3 = domain modules within packages

---

### 6. Runtime View

**Dynamic behavior**: How building blocks interact at runtime. Document architecturally significant scenarios:

- Important use cases / features
- Critical external interface interactions
- Startup / shutdown sequences
- Error and exception scenarios

**Format**: Sequence diagrams (ASCII or DOT), numbered step lists, BPMN flow charts. Pick the notation that best fits the scenario.

**How many**: 3-5 key scenarios. Don't document all flows -- focus on architecturally relevant ones (complex, risky, critical-path).

---

### 7. Deployment View

**Infrastructure mapping**: Where building blocks run physically.

#### Level 1: Infrastructure Overview

- Deployment diagram showing nodes (K8s cluster, databases, message broker, CDN)
- Mapping of building blocks to infrastructure elements
- Justification for deployment structure

#### Level 2: Detailed Infrastructure (if needed)

Zoom into specific infrastructure elements (e.g., K8s namespace layout, database replication topology).

**Align with Twelve-Factor App**:
- Factor X (Dev/Prod Parity): Document all environments
- Factor V (Build, Release, Run): Show the pipeline
- Factor VII (Port Binding): Document port assignments

---

### 8. Cross-cutting Concepts

Concepts that apply across multiple building blocks. Pick ONLY the relevant ones:

| Concept | When to Include |
|---------|----------------|
| Domain model | Always for DDD systems |
| Error handling | When strategy is non-obvious |
| Logging/Monitoring | How structured logging works, which metrics |
| Security | Authentication, authorization, encryption at rest/transit |
| Persistence | ORM strategy, migration approach, connection pooling |
| Communication patterns | NATS subject conventions, REST backoff strategy |
| Testing strategy | Unit vs integration vs E2E boundaries |
| Configuration | How Twelve-Factor config is managed |
| Build/Deploy | CI/CD pipeline description |

**Format**: Sub-section per concept with explanation and examples. Cross-reference Engineering Principles and Twelve-Factor App where applicable.

---

### 9. Architecture Decisions

Important, expensive, or risky architectural decisions with rationale.

**Format**: Architecture Decision Records (ADR):

```markdown
### ADR-001: Use Counting Bloom Filter for list membership

**Status**: Accepted
**Context**: Scoring checks list membership at 500 TPS via Aerospike network calls (~2ms each).
**Decision**: Replace with in-memory Counting Bloom Filters (~850ns lookup, 0.001% FPR).
**Consequences**: +4000x faster lookups, +114.5 MB memory per instance, requires CBF sync via NATS.
```

Avoid redundancy with Section 4 (Solution Strategy). Section 4 = high-level "what"; Section 9 = detailed "why" for specific decisions.

---

### 10. Quality Requirements

Detailed quality requirements beyond the top 3-5 from Section 1.2.

#### 10.1 Quality Requirements Overview

Table or tree of all quality categories with descriptions.

#### 10.2 Quality Scenarios

Concrete, measurable acceptance criteria:

```markdown
| ID | Category | Scenario | Metric |
|----|----------|----------|--------|
| QS-1 | Performance | Under 1000 RPS, API response time | < 50ms p99 |
| QS-2 | Reliability | Single node failure | No data loss, < 30s recovery |
| QS-3 | Security | Unauthorized API access attempt | Rejected with 401, logged, alerted |
```

---

### 11. Risks and Technical Debt

Ordered list of identified risks and technical debt:

```markdown
| # | Risk / Debt | Probability | Impact | Mitigation |
|---|-------------|------------|--------|------------|
| 1 | CBF false positive rate increases with list size | Medium | Low | Monitor FPR metric, auto-rebuild at threshold |
| 2 | NATS delivery failure causes stale CBF | Low | High | Periodic full rebuild (daily), staleness alert |
```

---

### 12. Glossary

Domain and technical terms. Essential for onboarding and cross-team communication.

```markdown
| Term | Definition |
|------|-----------|
| CBF | Counting Bloom Filter -- probabilistic data structure for set membership |
| AIC | Architecture Inception Canvas -- lightweight arc42 subset for initiative initiation |
```

---

## Diagram Conventions

All diagrams live in the `images/` folder inside the initiative directory.

**Dual format** (always both):
- Raw source: `images/<name>.dot`
- Compiled image: `images/<name>.png`
- Compile: `dot -Tpng images/<name>.dot -o images/<name>.png`
- Embed in docs: `![Title](images/<name>.png)`

**Naming**: lowercase, underscores, descriptive:
- `c4_context.dot` / `.png` — system in its environment
- `c4_container.dot` / `.png` — services/modules decomposition
- `archimate_business.dot` / `.png` — ArchiMate business layer
- `deployment.dot` / `.png` — infrastructure mapping
- `sequence_<flow>.dot` / `.png` — runtime scenario (or ASCII inline for simple flows)

**Tooling**:
- **Primary**: Graphviz DOT for C4, ArchiMate, component, deployment diagrams
- **Fallback**: Markdown ASCII for simple sequence flows (inline in docs, no image file needed)

---

## Relationship: AIC vs Arc42 vs TOGAF

```
AIC (aic.md)          Arc42 (arc42.md)           TOGAF (togaf.md)
-----------           ----------------           ----------------
Lightweight           Comprehensive              Enterprise
Every initiative      Complex systems            Cross-domain / compliance

Maps to:              Maps to:                   Maps to:
- Business Case       1. Introduction & Goals     1. Architecture Vision
- Quality Goals       2. Constraints              2. Business Architecture
- Constraints         3. Context & Scope          3. IS Architecture
- Business Context    4. Solution Strategy        4. Technology Architecture
- Arch Hypotheses     5. Building Block View      5. Opportunities & Solutions
- Risks               6. Runtime View             6. Migration Planning
- Tasks               7. Deployment View          7. Implementation Governance
                      8. Cross-cutting Concepts   8. Change Management
                      9. Architecture Decisions
                      10. Quality Requirements
                      11. Risks & Tech Debt
                      12. Glossary
```
