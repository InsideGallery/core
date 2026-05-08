# Workflow Policy (POL-ENG-002)

| Field | Value |
| :---- | :---- |
| Policy Name | Workflow Policy |
| Unique ID | POL-ENG-002 |
| Status | Draft |
| Owner | CTO / Engineering Management |
| Approver | CTO |
| Created | 2026-04-27 |
| Effective Date | After CTO approval |
| Next Review Date | 2026-10-27 |
| Related Policies | POL-GOV-001 Policy of Policies, POL-TECH-001 Policy of Initiatives, POL-ENG-001 Engineering Principles |

## Purpose

This policy defines the day-to-day product, engineering, QA, and deployment workflow for the team.
It turns the initiative lifecycle, Engineering Principles, trunk-based development, AIC/TSC documentation, Jira execution, QA, and Kubernetes deployment into one operational instruction.

The document also clarifies Product Owner responsibilities. The Product Owner owns product value, priorities, acceptance, route readiness, and
continuous clarification for the team.

## Policy Statement

All product and engineering work must flow through a visible lifecycle:

```text
Idea -> Product Discovery -> Initiative -> AIC/TSC -> Refinement -> Jira Stories
     -> Trunk-Based Development -> QA -> Deployment -> Production Validation -> Done
```

Work must not start unless the team understands the product outcome, acceptance criteria, architecture constraints, test strategy, and deployment path. Work is not done until it is deployed to production, validated by QA, and accepted against the agreed criteria.

## Scope

This policy applies to the following team members and roles:

| Name | Role |
| :---- | :---- |
| Ariel | CTO |
| Greg | DevOps Engineer |
| Maxim | Engineering Team Lead |
| Kate | Fullstack Engineer |
| Petro | AQA Engineer |
| Galya | Project Manager |
| David | Product Owner |
| Vince | Frontend Engineer |
| Mauricio | Backend Engineer |

The policy applies to:

- Product discovery and backlog work.
- Initiative creation and architectural definition.
- AIC and Tech Stack Canvas documentation.
- Jira epic and story management.
- Trunk-based development and merge request flow.
- Testing, QA, and acceptance.
- Kubernetes, ArgoCD, route migration, DNS, and production deployment.
- Production validation, rollback, and incident follow-up.

## Governing References

This policy operationalizes the following materials:

- `source/Policy of Policies.md` (POL-GOV-001) for policy lifecycle and RACI definitions.
- `source/Policy of Initiatives.md` (POL-TECH-001) for initiative lifecycle and AIC requirements.
- `source/Engineering Principles.md` (POL-ENG-001) for engineering, testing, review, deployment, branch, API, and NATS standards.
- `source/AIC Template.md` for Architecture Inception Canvas structure.
- `source/Tech Stack Canvas.md` for technology decision structure.
- `~/go/src/github.com/MetricAid/k8s/proxy.md` for current Kubernetes frontend migration and proxy deployment rules.
- Trunk-Based Development reference: https://trunkbaseddevelopment.com/

## Definitions

- **AIC**: Architecture Inception Canvas. Mandatory architecture/business framing document for
  initiatives.
- **TSC**: Tech Stack Canvas. Mandatory technology decision document for initiatives.
- **Trunk**: The main integration branch. It is the only permanent development branch.
- **Short-lived branch**: A branch created for a small task and merged back quickly, normally within
  one working day and no later than two working days unless the Engineering Team Lead approves an
  exception.
- **Feature flag**: A runtime switch that lets incomplete or risky behavior be merged without being
  exposed to users.
- **Release candidate**: A tested version that is ready for environment deployment.
- **Route takeover**: Moving a specific frontend route from legacy EC2 behavior to Kubernetes
  frontend ownership.
- **Done**: Deployed to production, QA validated, and accepted against AC.
- **Accountable (A)**: Owns the correct completion of the task.
- **Responsible (R)**: Performs the work.
- **Consulted (C)**: Provides required input before completion.
- **Informed (I)**: Must be kept up to date.

## Core Workflow Principles

1. **Performance by Design**
   - Performance is part of scope, architecture, review, testing, and deployment.
   - Each initiative must define expected sizing numbers or explicitly state why they are unknown.

2. **KISS**
   - Prefer the smallest solution that solves the product and operational problem.
   - Avoid unnecessary services, queues, flags, abstractions, or dependencies.

3. **DRY**
   - Reuse proven modules and patterns.
   - Do not create abstractions before repeated behavior is stable and understood.

4. **MDCA and DDD**
   - Organize work by business domain, not technical layer.
   - Keep bounded contexts clear. Modules communicate through explicit contracts.

5. **Clean Code**
   - Code must be readable, tested, reviewed, and maintainable.
   - Go code follows `gofumpt`, `golangci-lint v2`, effective Go naming, error wrapping, and
     package ownership rules from POL-ENG-001.

6. **Event-Driven Internal Communication**
   - Internal service communication uses NATS unless explicitly approved otherwise.
   - NATS subjects follow `{project}.{system}.{service}.{sync|async}.{event_description}`.

7. **REST With Backoff for External APIs**
   - External and third-party integrations use REST with retry/backoff behavior and clear failure
     handling.

8. **API Versioning**
   - New APIs start with `/v1/...`.
   - Breaking changes require a new version and an explicit deprecation plan.

9. **Kubernetes-First Deployment**
   - If a component can be deployed in Kubernetes, it must be deployed in Kubernetes.
   - All deployable components must expose health/readiness behavior suitable for the platform.

10. **Documentation Before Implementation**
    - Standard initiatives require AIC and TSC before development starts.
    - Architecture changes require AIC/TSC updates before implementation.

## Team Responsibilities

### CTO

The CTO owns technical strategy, architecture governance, and final approval for engineering policy.

The CTO is accountable for:

- Approving this workflow policy and any future policy changes.
- Setting technical direction and resolving architecture conflicts.
- Approving exceptions to Engineering Principles, initiative workflow, deployment policy, and architecture standards.
- Ensuring initiatives align with company strategy, technical risk tolerance, and long-term maintainability.
- Deciding when a solution requires deeper architecture documentation such as full arc42 or TOGAF.
- Escalating or stopping work when delivery pressure would create unacceptable product, security, reliability, or architecture risk.

The CTO must:

- Review high-risk initiatives, major architecture changes, and production-impacting exceptions.
- Support team-level decisions when normal authority is not enough to resolve a technical or delivery
  decision.
- Keep policy ownership clear and make final calls when responsibility conflicts remain unresolved.

### Product Owner

The Product Owner owns product outcomes, value, and business acceptance.

The Product Owner is accountable for:

- Defining the product problem, target users, business value, and expected outcome.
- Owning backlog priority and saying what matters first.
- Providing or approving business acceptance criteria before engineering starts.
- Explaining why a feature is needed, not only what screen or endpoint should exist.
- Clarifying edge cases, user expectations, and non-goals.
- Deciding whether a route, feature, or workflow is ready for user exposure from a product perspective.
- Accepting or rejecting completed work after QA evidence is available.
- Collecting stakeholder feedback after beta or production release.

The Product Owner must be proactive:

- Join discovery, refinement, sprint planning, demo, and release readiness discussions.
- Review Jira stories before implementation and confirm that AC represents product intent.
- Answer blocking product questions within one business day.
- Maintain a prioritized backlog with the Project Manager.
- Prepare examples, user flows, screenshots, customer notes, or business rules when AC is unclear.
- Review beta/stage behavior before production deployment when the change is user-facing.

The Product Owner must not:

- Delegate product priority to engineers.
- Approve a story without understanding the user/business outcome.
- Treat QA as the first product validation step.
- Ask for production release when AC, QA evidence, rollback, or monitoring is missing.
- Change scope during development without updating Jira, AIC/TSC if needed, and release plan.

### Project Manager

The Project Manager owns delivery flow, visibility, and coordination.

The Project Manager is responsible for:

- Maintaining Jira hygiene and ensuring each story has owner, status, AC, labels, and links.
- Scheduling discovery, refinement, grooming, QA handoff, demo, and release readiness sessions.
- Tracking dependencies, blockers, risks, and target dates.
- Ensuring the team follows this policy and escalates blocked decisions.
- Coordinating stakeholder communication before and after deployment.

### Engineering Team Lead

The Engineering Team Lead owns engineering execution quality.

The Engineering Team Lead is accountable for:

- Technical decomposition and implementation readiness.
- Enforcing Engineering Principles, Clean Code, API versioning, tests, review discipline, and trunk-based development.
- Ensuring stories are small enough for short-lived branches.
- Selecting reviewers and ensuring review feedback is resolved within two business days.
- Confirming technical readiness for deployment.
- Coordinating architecture questions with the CTO/Architect when decisions exceed team authority.

### AQA Engineer

The AQA Engineer owns QA strategy, automation, and release validation evidence.

The AQA Engineer is accountable for:

- Defining how the feature will be tested before implementation starts.
- Creating or updating automated tests where practical.
- Validating AC on dev/stage/beta/prod as appropriate.
- Defining regression scope and risk-based test coverage.
- Confirming production behavior after deployment.
- Reporting quality risks before release approval.

### DevOps Engineer

The DevOps Engineer owns platform, deployment mechanics, and operational readiness.

The DevOps Engineer is accountable for:

- Kubernetes, ArgoCD, ingress, ALB, Route53, certificates, and environment readiness.
- Deployment automation and rollback mechanics.
- Health checks, observability, logs, metrics, and alert readiness.
- Validating infrastructure changes before DNS or production cutover.
- Supporting incident response and infrastructure postmortems.

### Fullstack Engineer

The Fullstack Engineer combines the responsibilities of the Frontend Engineer and Backend Engineer
for assigned stories. When working fullstack, this role owns the end-to-end behavior across UI, API,
domain logic, integration contracts, tests, and release support.

### Frontend Engineer

The Frontend Engineer is responsible for frontend implementation, route behavior, UI state, API integration, and frontend test coverage.

The Frontend Engineer must:

- Keep route ownership explicit during migration.
- Avoid exposing unfinished behavior without feature flags.
- Validate auth/session/cookie behavior and browser regressions with the AQA Engineer.
- Keep frontend changes compatible with backend contracts and API versioning.

### Backend Engineer

The Backend Engineer is responsible for backend implementation, domain logic, APIs, NATS integrations, persistence, and backend test coverage.

The Backend Engineer must:

- Follow MDCA flow: Handler -> DTO -> Service -> Model -> Repository.
- Keep APIs versioned from day one.
- Use NATS for internal service communication and REST with backoff for external integrations.
- Add unit, integration, or flow tests based on risk and contract boundaries.

## RACI Matrix - End-to-End Workflow

Legend: A = Accountable, R = Responsible, C = Consulted, I = Informed.

| Workflow activity | CTO | DevOps Engineer | Engineering Team Lead | Fullstack Engineer | AQA Engineer | Project Manager | Product Owner | Frontend Engineer | Backend Engineer |
| :---- | :----: | :----: | :----: | :----: | :----: | :----: | :----: | :----: | :----: |
| Identify product problem and business value | C | I | C | I | C | C | A/R | I | I |
| Maintain product priority and roadmap order | C | I | C | I | C | R | A | I | I |
| Create initiative README business sections | C | I | C | C | C | R | A | C | C |
| Answer AIC/TSC product gaps | I | I | C | C | C | R | A/R | C | C |
| Define technical constraints and hypotheses | A | C | R | C | C | I | C | C | C |
| Complete AIC and TSC for standard initiative | A | C | R | C | C | R | C | C | C |
| Review architecture before implementation | A | C | R | C | C | I | C | C | C |
| Approve architecture or Engineering Principles exception | A/R | C | C | I | I | I | I | I | I |
| Decompose approved initiative into Jira epics/stories | C | I | A | C | C | R | C | C | C |
| Define business acceptance criteria | I | I | C | C | C | R | A/R | C | C |
| Define test approach and QA evidence | I | C | C | C | A/R | C | C | C | C |
| Confirm story is Ready for Development | I | I | A | C | C | R | C | C | C |
| Implement fullstack stories | I | I | A | R | C | I | C | C | C |
| Implement frontend stories | I | I | A | C | C | I | C | R | C |
| Implement backend stories | I | I | A | C | C | I | C | C | R |
| Maintain trunk-based branch discipline | C | I | A/R | R | C | I | I | R | R |
| Perform merge request review | C | C | A | R | C | I | I | R | R |
| Run QA and regression checks | I | C | C | C | A/R | I | C | C | C |
| Approve business acceptance | I | I | C | I | C | R | A | I | I |
| Prepare deployment and rollback plan | I | A/R | C | C | C | I | I | C | C |
| Approve technical release readiness | C | C | A/R | C | C | I | I | C | C |
| Approve product go/no-go | I | I | C | I | C | R | A | I | I |
| Approve high-risk production release exception | A/R | C | C | I | C | I | C | I | I |
| Deploy to dev/stage/beta/prod | I | A/R | C | I | C | I | I | I | I |
| Validate production after deployment | I | R | A | C | R | I | C | C | C |
| Communicate release status to stakeholders | I | I | C | I | C | R | A | I | I |
| Lead incident response for application issue | I | C | A/R | C | C | R | I | C | C |
| Lead incident response for infrastructure issue | I | A/R | C | I | C | R | I | I | I |
| Complete postmortem and follow-up tasks | I | C | A | C | C | R | C | C | C |
| Approve workflow policy changes | A/R | I | C | I | I | C | C | I | I |

## Workflow Stages

### 1. Product Discovery

Lead: Product Owner.

Trigger: new product idea, customer request, operational pain, incident follow-up, compliance need,
or technical risk.

Required output:

- Problem statement.
- User or stakeholder group.
- Business value.
- Initial scope and non-goals.
- Success criteria.
- Known constraints.
- Open questions.

Rules:

- The Product Owner owns the "why" and priority.
- The Project Manager owns visibility and follow-up.
- The Engineering Team Lead confirms whether the work needs architecture review; the CTO makes the
  final call for strategic or high-risk architecture questions.
- No Jira development story is created until the problem and expected outcome are clear.

### 2. Initiative Creation

Lead: Engineering Team Lead.

Standard initiatives must follow POL-TECH-001 and use AIC plus TSC.

Required AIC inputs:

- Business Case.
- Functional Overview.
- Quality Goals from ISO-25010.
- Organizational Constraints.
- Technical Constraints, including POL-ENG-001 references.
- Business Context.
- Architectural Hypotheses.
- Technical Challenges and Risks.
- Tasks after refinement.

Required TSC inputs:

- Business Feature Description.
- System Name and Description.
- Sizing Numbers.
- Major Quality Attributes.
- Services.
- Architecture Representation.
- Sequence Flow when useful.
- Frontend, backend, and data technology choices.
- APIs and integrations.
- Security and compliance.
- Testing and QA.
- Infrastructure and deployment.
- Monitoring and analytics.
- Development workflow and collaboration.

Rules:

- Missing information must be asked directly. Do not guess.
- AIC/TSC must be updated when decisions change.
- Diagrams must live in the initiative `images/` folder as `.dot` source plus `.png` render where
  Graphviz is used.
- Technical initiatives may use the simplified AIC path, but must still include risks and tasks.

### 3. Refinement and Jira Decomposition

Lead: Project Manager.

Trigger: AIC/TSC are reviewed enough to plan implementation.

Definition of Ready for a Jira story:

- Story has a clear user/business or technical outcome.
- AC is written and testable.
- Design/architecture link is attached when relevant.
- Dependencies and affected services are known.
- Test approach is known.
- Deployment impact is known.
- Story is small enough for a short-lived branch.
- Owner is assigned.

Jira rules:

- Epics represent delivery milestones or initiatives.
- Stories represent one sprint-sized unit of work.
- Story summaries use a layer prefix such as `[Frontend]`, `[Backend]`, `[Platform]`, or `[Cross]`.
- Story descriptions include Acceptance Criteria.
- Story points use Fibonacci sizing and must not exceed 8. Larger work must be split.
- Stories link back to AIC/TSC or `plan.md` when those artifacts exist.
- The Project Manager keeps status current; the Engineering Team Lead keeps technical scope current;
  the Product Owner keeps priority current.

### 4. Trunk-Based Development

Lead: Engineering Team Lead.

The team uses trunk-based development.

Rules:

- `main` or `master` is the trunk and the permanent integration branch.
- Long-lived feature branches are prohibited.
- Short-lived branches are allowed for review, but must be merged quickly.
- Engineers integrate small slices frequently.
- Incomplete behavior must be protected by feature flags, configuration flags, branch by
  abstraction, or hidden routes.
- CI must pass before merge.
- Code must be reviewed before merge unless an emergency hotfix exception is declared.
- Release branches are allowed only for stabilization or emergency patching and must not become a
  place for new feature development.

Branch naming:

- Current POL-ENG-001 epic branch format is `epic/{TASK}_{TITLE}`.
- To keep this compatible with trunk-based development, epic branches must be short-lived
  integration branches, not long-running parallel trunks.
- Subtask branches must be small and merge into trunk or the approved epic integration branch
  quickly.
- If a repo can support direct short-lived task branches to trunk, that is preferred.

Merge request rules:

- The author is responsible for finding a reviewer.
- Reviewers must respond within two business days.
- Review must check logic, AC, test coverage, Engineering Principles, API versioning, observability,
  and deployment risk.
- Test coverage must not decrease.
- Subtask work may be reviewed incrementally, but the final epic/bug solution must be coherent and
  traceable.

### 5. Engineering Implementation Rules

Lead: Engineering Team Lead.

General rules from POL-ENG-001:

- Use Clean Code principles.
- Avoid `fmt.Print*`, `log.Print*`, and bare `print*` in application code. Use structured logging.
- Never ignore errors.
- Error strings are lowercase and wrap context with `%w`.
- Prefer early returns and left-aligned happy path.
- Keep functions focused; split functions that grow beyond clear readability.
- Avoid flag arguments that make one function do two different things.
- Use full-word naming and correct acronym casing, such as `userID` and `HTTPClient`.
- No `Get` prefix for simple accessors.
- Use compile-time interface checks where appropriate.
- Each package owns its env config prefix.
- Logs go to stdout only.
- Services shut down gracefully.
- Jobs must be idempotent.

Go/MDCA rules:

- Domain models have storage tags only.
- DTOs have JSON tags only.
- Handler never sees domain storage models.
- Repository never sees DTOs.
- Consumer-side packages define ports/interfaces where they are used.
- Cross-service internal communication uses events, not service-to-service HTTP.

Testing rules:

- Unit tests cover business logic not covered by integration tests.
- Integration tests are required for service and database communication.
- Handler tests use framework test helpers and mocked services.
- Flow tests validate cross-service behavior against a running stack where required.
- QA automation should be added for stable product paths and critical regressions.

### 6. QA and Acceptance

Lead: AQA Engineer.

QA starts before coding.

Rules:

- The AQA Engineer defines test strategy during refinement.
- Each story must have "how to test" or equivalent QA evidence.
- Engineers support QA with test data, feature flag instructions, environment notes, and logs.
- The Product Owner validates product behavior; the AQA Engineer validates quality and regression
  risk.
- A story cannot be moved to Done only because code was merged.

Minimum QA evidence:

- AC pass/fail result.
- Environment tested.
- Build or commit reference.
- Regression scope.
- Known issues or explicit "none known".
- Product acceptance result for user-facing changes.

### 7. Deployment Readiness

Lead: DevOps Engineer.

Before any non-trivial deployment, the team must confirm:

- AIC/TSC or story context is current.
- CI passes.
- MR is approved and merged according to branch policy.
- Database migrations are reviewed and reversible or forward-safe.
- Feature flags are configured.
- Rollback plan exists.
- Monitoring and logs are available.
- QA sign-off exists for the target environment.
- Product go/no-go is explicit for user-facing changes.
- The DevOps Engineer confirms platform readiness for infrastructure-affecting changes.
- The Engineering Team Lead confirms technical readiness.
- The CTO is consulted for high-risk, architecture-impacting, or exception-based production
  releases.

Deployment environments:

- `dev`: early engineering validation.
- `stage`: integrated QA and product review.
- `beta`: controlled production-like exposure.
- `prod`: production users.

Done means:

- Production deployed.
- QA validated production or the agreed production-safe smoke scope.
- Product acceptance completed when user-facing.
- Jira updated.
- Monitoring checked after release.

## Deployment Policy

### Kubernetes and ArgoCD

- Deployable infrastructure and services must be deployed through Kubernetes where possible.
- ArgoCD is the expected synchronization mechanism for Kubernetes application state.
- Runtime configuration must be environment-specific and reviewed before sync.
- Health, readiness, and liveness behavior must be available for deployable services.
- Logs must go to stdout and be visible in the standard observability stack.

### Multi-Environment Frontend Migration

The frontend migration pattern separates DNS cutover from application cutover.

Environment mapping:

| Environment | Domain | Cluster |
| :---- | :---- | :---- |
| dev | `app.dev.metricaid.ca` | dev EKS cluster, `dev` namespace |
| stage | `app.stg.metricaid.ca` | dev EKS cluster, `stage` namespace |
| beta | `app.beta.metricaid.com` | prod EKS cluster, `beta` namespace |
| prod | `app.metricaid.com` | prod EKS cluster, `prod` namespace |

Legacy upstream mapping:

| Environment | Legacy upstream IP |
| :---- | :---- |
| dev | `<IP-HERE>` until the DevOps Engineer provides the real IP |
| stage | `<IP-HERE>` until the DevOps Engineer provides the real IP |
| beta | `52.60.108.117` |
| prod | `52.60.174.103` |

Policy:

- Default behavior sends all traffic to the legacy host through `legacy-proxy`.
- Kubernetes frontend ownership is explicit per route.
- No frontend service may claim `/` until final full cutover is approved.
- Fast rollback must remain possible by removing the route from the Kubernetes-owned route list and
  syncing ArgoCD.

Target shape:

- ALB receives internet traffic.
- Unmatched/default traffic routes to `legacy-proxy`.
- `legacy-proxy` forwards to the legacy upstream IP for the environment.
- Explicit ready routes route directly to `andromeda-v2`.
- `legacy-proxy` and `andromeda-v2` must share the same environment host and ALB IngressGroup.

Repository implementation rules:

- `services/legacy-proxy/` must exist as a standalone chart because the shared service chart assumes
  Ptolemy binaries.
- `legacy-proxy` requires Deployment, ConfigMap, Service, and Ingress resources.
- `legacy-proxy` service name is `legacy-proxy-sv` unless the chart explicitly documents another
  name.
- `legacy-proxy` ingress owns `/` catch-all.
- `legacy-proxy` nginx/envoy config must preserve `Host`, `X-Forwarded-For`, and
  `X-Forwarded-Proto`.
- Websocket/SSE routes require upgrade headers and appropriate timeouts.
- If the legacy upstream requires HTTPS, the proxy config must use HTTPS upstream behavior and
  validate SNI/server name needs.
- `services/andromeda-v2/templates/ingress.yaml` must support `ingress.paths` as a list.
- `ingress.path` fallback is allowed only for backward compatibility.
- Environment values must not default `andromeda-v2` to `/` before final cutover.

Required ingress behavior:

- `legacy-proxy` owns catch-all `/` fallback.
- `andromeda-v2` owns only explicit `ingress.paths` until final cutover.
- Suggested group order:
  - `andromeda-v2`: `100`
  - `legacy-proxy`: `1000`
- Duplicate `/` ownership by `andromeda-v2` is prohibited before final cutover.

### Route Ownership Process

Route list ownership:

- `services/andromeda-v2/values.<env>.yaml` `ingress.paths` is the source of truth for Kubernetes
  frontend route ownership.
- Adding a route means Kubernetes serves that route after ArgoCD sync.
- Removing a route means traffic falls back to legacy after ArgoCD sync.

Route takeover flow:

1. The Product Owner marks the route product-ready.
2. The Frontend Engineer and/or Fullstack Engineer confirm frontend readiness.
3. The Backend Engineer confirms backend/API readiness if the route depends on backend behavior.
4. The AQA Engineer confirms QA evidence for the route.
5. The DevOps Engineer confirms ingress, ALB, certificate, and environment readiness.
6. The Engineering Team Lead confirms technical release readiness.
7. PR adds the route to `ingress.paths`.
8. ArgoCD sync applies the change.
9. The AQA Engineer validates the route in the target environment.
10. The Product Owner accepts the product behavior.

Rollback:

1. Remove the route from `ingress.paths`.
2. Sync ArgoCD.
3. Confirm traffic falls back to legacy behavior.
4. Record root cause and follow-up work in Jira.

### DNS Cutover Rules

Before Route53 DNS changes:

- `legacy-proxy` is deployed and healthy.
- ALB health checks are green.
- `legacy-proxy` can reach the environment legacy IP.
- ALB TLS certificate is valid for the target domain.
- Host header, auth, cookies, sessions, websocket/SSE needs, and CORS behavior are tested.
- Rollback plan is written and understood by the DevOps Engineer and Engineering Team Lead.

DNS change:

- Move the environment domain from direct legacy A record to Alias A pointing to ALB DNS.
- Do not combine DNS cutover with broad route takeover unless approved by the Engineering Team Lead
  and Product Owner.
- Keep legacy fallback as default after DNS switch.

After DNS change:

- Smoke-test legacy fallback through ALB.
- Validate auth/session/cookie behavior.
- Validate websocket/SSE behavior where applicable.
- Check mixed-content and CORS behavior.
- Monitor error rate, latency, and user-impacting logs.
- Keep route takeover incremental.

### Final Full Cutover

Final cutover is allowed only when:

- All required frontend routes are owned by Kubernetes.
- Legacy fallback has not been used for the migrated user journey during the agreed soak period.
- QA regression scope is complete.
- The Product Owner accepts product behavior.
- The Engineering Team Lead approves technical readiness.
- The DevOps Engineer approves platform readiness.
- The CTO approves the final cutover when it affects production architecture, platform risk, or
  strategic migration scope.
- Rollback and decommission plans are written.

Final cutover steps:

1. Add `/` to `andromeda-v2` only after final approval.
2. Sync ArgoCD.
3. Validate all core journeys.
4. Monitor production.
5. Keep `legacy-proxy` during the agreed soak period.
6. Decommission `legacy-proxy` only after explicit approval.

## Product Owner Operating Guide

The Product Owner must use this checklist to stay ahead of the team.

Before refinement:

- What user or business problem are we solving?
- Who needs this and why now?
- What is the expected measurable outcome?
- What is explicitly out of scope?
- What are the top edge cases?
- What must be true for the user to accept this?
- Which routes, pages, APIs, or workflows are impacted?

During refinement:

- Confirm that every story has testable AC.
- Confirm that engineering understands priority and tradeoffs.
- Decide what can be released behind a flag and what cannot.
- Identify stakeholders who must review beta/stage behavior.
- Make unresolved product questions visible immediately.

During development:

- Answer blocking questions within one business day.
- Review intermediate demos or screenshots.
- Do not add scope silently. If scope changes, update Jira and affected docs.
- Keep priority stable unless business urgency changes.

Before release:

- Review QA evidence.
- Validate the feature on stage or beta for user-facing changes.
- Give explicit product go/no-go.
- Confirm communication needs with the Project Manager.

After release:

- Review production behavior and stakeholder feedback.
- Confirm whether the outcome was achieved.
- Create follow-up backlog items for gaps.
- Close the loop with users or stakeholders.

## Meeting and Communication Cadence

| Meeting | Owner | Required participants | Purpose | Output |
| :---- | :---- | :---- | :---- | :---- |
| Product Discovery | Product Owner | Project Manager, Engineering Team Lead, affected engineers as needed | Define problem, value, scope | Discovery notes or initiative README |
| Architecture Review | Engineering Team Lead | CTO as needed, DevOps Engineer, engineers, AQA Engineer | Validate AIC/TSC and risks | Approved architecture direction |
| Refinement/Grooming | Project Manager | Product Owner, Engineering Team Lead, AQA Engineer, assigned engineers | Prepare stories for development | Ready Jira stories |
| Daily Workflow Check | Project Manager | Team | Surface blockers and status | Updated blockers/status |
| QA Handoff | AQA Engineer | Assigned engineers, Product Owner as needed | Confirm test scope and evidence | QA plan and test data needs |
| Release Readiness | Project Manager | Product Owner, Engineering Team Lead, AQA Engineer, DevOps Engineer, assigned engineers, CTO for high-risk releases | Go/no-go and rollback review | Release decision |
| Demo/Acceptance | Product Owner | Project Manager, AQA Engineer, assigned engineers | Product acceptance | Accepted or follow-up stories |
| Postmortem | Engineering Team Lead or DevOps Engineer | Impacted participants | Learn from incidents | Root cause and follow-up tasks |

Communication rules:

- Decisions that affect scope, architecture, deployment, or acceptance must be written in Jira,
  AIC/TSC, `plan.md`, or release notes.
- Slack/chat is acceptable for coordination, not as the only place where decisions live.
- Blockers older than one business day must be escalated by the Project Manager.
- Product questions blocking implementation must be answered by the Product Owner or escalated.

## Allowed and Prohibited Conduct

### Allowed

- Start discovery before all answers are known.
- Use AIC/TSC to expose gaps and risks.
- Merge small, safe slices frequently.
- Hide incomplete work behind feature flags.
- Deploy route-by-route through explicit `ingress.paths`.
- Roll back by removing route ownership from Kubernetes.
- Ask for CTO/Architect review when architecture risk is high.

### Prohibited

- Starting development without clear AC.
- Starting standard initiative implementation without AIC and TSC.
- Keeping long-lived feature branches.
- Decreasing test coverage in a merge request.
- Deploying a breaking API change without version bump.
- Using service-to-service HTTP for internal communication without explicit approval.
- Claiming `/` in `andromeda-v2` before final cutover.
- Changing DNS before legacy fallback, ALB health, TLS, and rollback are validated.
- Moving Jira to Done before production deployment and QA validation.
- Treating Product Owner approval as optional for user-facing production changes.

## Compliance and Exceptions

Exceptions require explicit written approval:

- Architecture exceptions: CTO or assigned Architect.
- Branching or review exceptions: Engineering Team Lead.
- Production release exceptions: Engineering Team Lead plus Product Owner; CTO for high-risk or
  architecture-impacting releases.
- Infrastructure or DNS exceptions: DevOps Engineer plus Engineering Team Lead.
- Product acceptance exceptions: Product Owner.
- Workflow policy exceptions or policy changes: CTO.

Emergency hotfix process:

1. Declare emergency in the team channel.
2. Assign one accountable engineer and one reviewer.
3. Keep the branch short-lived and scope minimal.
4. Run the fastest relevant CI/test path.
5. Deploy with rollback ready.
6. Validate production.
7. Create follow-up Jira items for missing tests, documentation, or cleanup within one business day.

## Consequences of Non-Compliance

Failure to follow this workflow can cause unclear ownership, missed requirements, release defects,
production incidents, and unplanned technical debt. Non-compliance must be corrected through
coaching, process correction, or formal escalation depending on severity and repetition.

## Revision History

| Version | Date | Author | Changes |
| :---- | :---- | :---- | :---- |
| 0.1 | 2026-04-27 | Codex | Initial draft combining initiative lifecycle, policy governance, Engineering Principles, trunk-based development, PO responsibilities, RACI, and Kubernetes deployment workflow. |
| 0.2 | 2026-04-27 | Codex | Added CTO role responsibilities, RACI ownership, architecture governance, release exception, and policy approval responsibilities. |
| 0.3 | 2026-04-27 | Codex | Converted operational instructions, RACI, meetings, approvals, and exceptions from person names to role names; kept names only in the people-to-roles mapping. |
| 0.4 | 2026-04-27 | Codex | Simplified Fullstack Engineer responsibilities by defining the role as Frontend Engineer plus Backend Engineer responsibilities. |
| 0.5 | 2026-04-27 | Codex | Removed agent-specific process reference from governing references. |
| 0.6 | 2026-04-27 | Codex | Added explicit lead role for each workflow stage. |
