# Engineering Principles (POL-ENG-001)

## Purpose

The purpose of this policy is to establish a unified and consistent framework of engineering principles for the design, development, testing, and deployment of all company software. By adhering to these principles, we will ensure our systems are simple, high-performing, maintainable, and aligned with strategic goals. This policy fosters clarity, accountability, and technical excellence across the engineering department.

---

## Policy Statement

All engineering, product, and technical personnel must create, review, and manage software initiatives and respond to operational failures in accordance with the framework established in this document. This "Engineering Principles" Policy is the single source of truth for the company's software development lifecycle, standards, and incident analysis process. No software initiative is considered complete, and no incident response is finalized, until it fully complies with the requirements outlined herein.

---

## Scope

This policy applies to all full-time employees and contractors involved in the software development lifecycle, including but not limited to roles such as Engineer, Team Lead, Tech Lead, System Analyst, Product Manager, and Engineering Manager. It governs all activities related to software architecture, coding, testing, review, deployment, and incident analysis.

This policy does not supersede external laws or regulations. All software and systems developed must comply with applicable legal and regulatory frameworks.

---

## Definitions

* **Acceptance Criteria (AC)**: A set of predefined requirements, constraints, and edge cases that must be met for a task to be considered complete.
* **Incident**: An unplanned interruption to an IT service or a reduction in the quality of an IT service.
* **System Down-time**: Any period during which a system is unavailable or fails to perform its primary functions, impacting users or business operations.
* **Post-mortem**: A formal document analyzing an incident, its impact, the actions taken to resolve it, the root cause, and the follow-up actions required to prevent recurrence.
* **Root Cause**: The fundamental issue which, if addressed, will prevent the incident from recurring.
* **KISS (Keep It Simple and Smart)**: A design principle that states systems work best if they are kept simple rather than made complex.
* **DRY (Don't Repeat Yourself)**: A principle aimed at reducing repetition of software patterns.
* **Accountable (A)**: The person ultimately answerable for completing the correct and thorough task.
* **Responsible (R)**: The individual(s) who perform the work to complete the task.
* **Consulted (C)**: Individuals whose input and opinions are sought.
* **Informed (I)**: Individuals who are kept up to date on progress.

---

## Procedures

### Architectural & Design Principles

* **Communication**: All internal service communication must use an event-driven approach. REST communication with a backoff mechanism must be used for all third-party integrations.
* **Documentation**: All external communication and APIs must be documented in Swagger.
* **Core Principles**: All design must prioritize **Performance by Design**. Solutions must adhere to **KISS** to avoid unnecessary complexity and **DRY** to prevent code repetition.
* **Architecture Changes**: Any changes to architecture require discussion and documentation via an Architecture Inception Canvas or Architecture Communication Canvas before implementation.

### Development & Testing Procedures

* **Code Quality**: Developers must follow **Clean Code** principles.
* **Test Coverage**: Test coverage must be improved over time and is not permitted to decrease in any merge request. Reviewers must check test coverage.
* **Test Types**: Integration tests are required for communication with other services and databases. Unit tests are required for functionality not covered by integration tests.
* **Task Management**: A task is considered "Done" only after it has been deployed to production and tested by QA.

### Code Review & Deployment Procedures

* **Reviewer Responsibility**: Reviewers must provide feedback within two business days. The review must focus on logic, algorithms, adherence to Engineering Principles, and fulfillment of the Acceptance Criteria.
* **Developer Responsibility**: The developer who authors the code is responsible for finding a reviewer, ensuring all pipeline checks pass, merging approved code, and moving the task to "Done" upon release.
* **Deployment**: All infrastructure must be deployed in a unified way. If a component is deployable in Kubernetes, it must be deployed in Kubernetes.

### Branch Naming Convention

* **Epic Branches**: All branches related to an epic must follow this naming structure:
  * **Format**: `epic/{TASK}_{TITLE}`
  * **{TASK}**: The full Task ID (e.g., `DAF-100`).
  * **{TITLE}**: A short, descriptive title, with words separated by a dash (`-`).
  * **Example**: `epic/DAF-123-user-authentication-flow`

### Merge Request (MR) Submission

* **Decomposition**: An epic may be broken down into multiple subtasks. Developers are permitted to create separate branches and MRs for subtasks.
* **Final Merge**: The final solution for a single epic or bug must be consolidated into **one final Merge Request**. All subtask branches must be merged into the primary epic branch before final review and merge.

### API Development & Versioning

* **Mandatory Versioning**: Each API endpoint must be deployed with a specific version in the URL path (e.g., `/v1/users`). The initial version of any new service must be `v1`.
* **Backward Incompatibility**: Each change that is not backward-compatible must be deployed to a new version number higher than the previous one (e.g., `v2`).
* **Deprecation**: When a new API version is released, partners must be informed how long the previous version will be supported. By default, a previous version must exist for at least one month.
* **Legacy Endpoints**: All legacy endpoints without a version must have a corresponding versioned endpoint created (starting from `v1`). A migration plan with a due date must be defined.
* **Exceptions**: Adding a "non-versioned" API must be discussed individually and requires explicit written approval from the CTO and the BE Lead.

### NATS Communication and Subject Naming

* **Naming Convention**: The name of a NATS subject must follow the pattern: `{project}.{system}.{service}.{kind of event (sync or async)}.{what is the message about}`.
  * Example: `[company].scoring.api-server.async.scoring_event`.
* **Responsibility**: The service that writes/produces the message is responsible for defining the subject names and their data structures.
* **Versioning**: To create a new version of an existing subject, a version number (e.g., `v2`, `v3`) may be added to the end of the subject name. The initial version does not use a `v1` suffix.
* **Renaming Subjects**: To rename a subject, a new service must be created to publish events with the new topic. After new versions of all consumers are prepared and deployed, the old services can be removed once 100% of traffic is using the new version.
* **Request/Reply**: Services that communicate using a request/reply pattern must have only one consumer. Every NATS request-reply handler must send a reply back to the client, even if an error occurs.

---

## Roles & Responsibilities

| Rule/Step RACI | A | R | C | I |
| :---- | :---- | :---- | :---- | :---- |
| Writing high-quality, tested code; creating and following AC; finding a reviewer and seeing a task through to deployment. | Team Lead | Engineer | Software Arch. / CTO / Eng Manag. | Team Lead |
| Providing timely and thorough reviews of Merge Requests. | Team Lead | Engineer | Software Arch. / CTO / Eng Manag. | Team Lead |
| Adherence to Engineering Principles | CTO / Eng Manag. | Team Lead | CTO / Eng Manag. | Engineer |
| Ensuring API Development & Versioning Standards are met. | CTO / Team Lead | Engineer | Cloud Arch. / Software Arch. | Eng Manag. |

---

## Allowed & Prohibited Conduct

### Allowed

* Following an event-driven approach for internal communication and REST for external communication.
* Writing simple, clean, and non-repetitive code (KISS, DRY, Clean Code, MDCA).
* Improving test coverage in all merge requests.
* Proposing architectural changes by filling out an Architecture Inception Canvas.
* Deploying infrastructure in a unified way using Kubernetes where applicable.
* Always using the `epic/{TASK}_{TITLE}` format for branches related to an epic.
* Creating sub-branches from an epic branch for individual subtasks.
* Consolidating all work for an epic into a single, final Merge Request for review.
* Deploying all new API endpoints with a specific version (e.g., v1).
* Following the standard naming convention for all NATS subjects.

### Prohibited

* Starting development on a task that does not have clear Acceptance Criteria.
* Decreasing test coverage in a merge request.
* Making architectural changes without prior discussion and approval.
* Over-engineering solutions or introducing unnecessary complexity.
* A reviewer failing to provide feedback within the required two-day timeframe without communication.
* Deploying a non-backward-compatible API change without incrementing the version number.
* Releasing a "non-versioned" API without explicit approval from the CTO and BE Lead.
* Renaming an existing NATS subject without following the new service creation and migration procedure.

---

## Consequences of Non-Compliance

Failure to adhere to this Engineering Principles Policy can expose the company to significant technical debt, security vulnerabilities, and operational risks. Any violation is considered a serious matter. Employees who fail to comply may be subject to disciplinary action, ranging from coaching and re-training up to and including formal disciplinary measures.
