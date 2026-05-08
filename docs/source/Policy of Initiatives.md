# Policy of Initiatives (POL-TECH-001)

## Purpose

The purpose of this policy is to establish a standardized, transparent, and efficient framework for proposing, defining, architecting, and managing all new business and technical initiatives. By mandating the use of the Architecture Inception Canvas (AIC) template, this policy ensures that every initiative is consistently documented, architecturally sound, and aligned with strategic goals before development work begins. This process is designed to foster collaboration, mitigate risk, and ensure clarity and accountability throughout the initiative lifecycle.

---

## Policy Statement

All new initiatives, regardless of their origin (Business or Technical), must be documented using the official AIC Template. No initiative shall be considered valid or approved for development until it has progressed through the procedures outlined in this document. This policy is the single source of truth for the initiative management lifecycle and supersedes any previous informal processes.

---

## Scope

This policy applies to all full-time employees across all departments and locations involved in the conception, design, approval, and implementation of new features, systems, or technical improvements. It governs the entire lifecycle of all documents related to initiatives, including the mandatory AIC and any subsequent proof-of-concept (PoC) or design artifacts.

---

## Definitions

* **Initiative**: A formal proposal for creating a new feature, system, or significant technical change. All initiatives are categorized as either a "Standard Initiative" or a "Technical Initiative."
* **Architecture Inception Canvas (AIC)**: The mandatory document, based on the Arc42 framework, used to describe and define an initiative.
* **Standard Initiative**: An initiative with a direct business impact, originating from either Business or Technical teams. It requires the full AIC documentation.
* **Technical Initiative**: A small-scale initiative with no direct business impact (e.g., refactoring, service renaming). It follows a simplified documentation process.
* **Initiator**: The individual who first proposes an initiative and is responsible for its initial documentation.
* **Architect**: The individual(s) responsible for defining the architectural solution. The CTO, Software Architect, or Cloud Architect fulfills this role.
* **Initiative Stages**:
  * **Genesis**: The initial phase, covering the creation of the AIC and an optional Proof of Concept (PoC).
  * **Custom**: The development of core, production-grade functionality (MVP).
  * **Product**: The implementation of the full-featured initiative as designed.
* **Business Context**: A schema definition of which part of the system will be affected by changes. Can include text description and schema/diagram. More details on how to write Business Context can be found in the [arc42 official documentation](https://docs.arc42.org/section-3/).

---

## Procedures

### Stage 1: Initiative Initiation

1. **Standard Initiative Initiation**:
   * An initiative can be proposed by authorized personnel from either the Business or Technical teams.
     * **Business Initiators**: Engineering Manager, Product Owner, Delivery Manager, System Analyst, Business Analyst.
     * **Technical Initiators**: CTO, Product Owner, Software Architect, Cloud Architect, Team Lead, and Engineering Manager.
   * The Initiator must create a new initiative document using the official AIC Template.
   * The Initiator is responsible for completing the "Collect Business Requirements" sections, including the **Business Case**, **Functional Overview**, and **Quality Goals**. Quality Goals must be selected from the official Arc42 list.

2. **Technical Initiative Initiation**:
   * Any team member may propose a Technical Initiative.
   * The Initiator must use the AIC Template, but is only required to complete: **Header (Meta)**, **Business Context**, **Architectural hypotheses**, and **Tasks**.
   * The Business Case section must be filled with: "*No impact*."

### Stage 2: Architectural Definition

1. Upon initiation, the initiative is assigned to an Architect (CTO, Software Architect, or Cloud Architect).
2. The Architect is responsible for completing the technical sections of the AIC, with a primary focus on **Architectural Hypotheses** and **Technical Challenges & Risks**.
3. The Architect must supplement the 'Architectural Hypotheses' section with relevant diagrams. It is strongly recommended to use standard notations, such as the [C4 model](https://c4model.com/diagrams), [BPMN](https://en.wikipedia.org/wiki/Business_Process_Model_and_Notation), [UML](https://en.wikipedia.org/wiki/Unified_Modeling_Language), or [flowcharts](https://en.wikipedia.org/wiki/Flowchart), where applicable.
4. The Initiator may define the vision and initial proposal, but the final architectural decision rests with the assigned Architect.

### Stage 3: Review and Refinement

1. The assigned Architect or Engineering Manager can schedule and lead a **Refinement** to gather additional information or a pre-preview of the initiative.
2. The assigned Architect or Engineering Manager schedules and leads a **Grooming meeting** to present the completed AIC to the relevant development team(s). At least one representative from each chapter related to the initiative's context must be invited.
3. The purpose of this meeting is to present the proposed architecture, gather feedback, and identify potential issues.
4. For a Technical Initiative, the AIC must be reviewed and approved by at least one of: CTO, Software Architect, or Cloud Architect. This can occur asynchronously or in a meeting.

### Stage 4: Task Separation and Implementation

1. Following the refinement meeting and incorporation of feedback, the initiative is considered architecturally approved.
2. A Product Manager, Engineering Manager, or Delivery Manager is responsible for breaking down the approved initiative into detailed epics and user stories. Every Epic must have Acceptance Criteria.
3. A Product Manager, Engineering Manager, or Delivery Manager must communicate and present epics to QA engineers.
4. The work then proceeds through the defined Initiative Stages (Genesis, Custom, Product).

---

## Roles & Responsibilities

| Rule/Step RACI | A | R | C | I |
| :---- | :---- | :---- | :---- | :---- |
| Identifying the need for an initiative and creating the initial AIC document with all the required business context. | Sys. Analyst | Sys. Analyst | CTO | Eng Manag. |
| Overall integrity of the initiative process and the quality of all architectural solutions. | CEO | CTO | Cloud Arch. / Software Arch. | DM / Eng Manag. / Prod. Manag. |
| Defining the technical solution and completing the architectural sections of the AIC. | CTO | Software Arch. / Cloud Arch. / CTO | CTO | Engineer |
| Presenting it for review. | CTO | Eng Manag. / Software Arch. / Cloud Arch. / CTO | Team Lead | Engineer |
| Decomposing the approved initiative into actionable development tasks (epics, stories). | CTO | Prod. Manag. / DM / Eng Manag. | Team Lead | Engineer |
| Describe "How to test". | DM | Team Lead | Engineer | Engineer |

---

## Allowed & Prohibited Conduct

### Allowed

* Always use the official AIC Template for any new initiative.
* Follow all defined stages of the initiative lifecycle, from initiation to task separation.
* Support architectural proposals with clear diagrams (e.g., C4, BPMN, Flowcharts) to enhance understanding.
* For Standard Initiatives, select Quality Goals exclusively from the [predefined Arc42 list](https://docs.arc42.org/section-1/#12-quality-goals).
* Use the simplified "Technical Initiative" process for changes with no business impact.

### Prohibited

* Do not begin development work on any initiative that does not have a formally reviewed AIC document.
* Do not bypass the Architectural Definition or Review and Refinement stages.
* Do not use custom or undefined Quality Goals in an AIC document.
* Do not create, enforce, or distribute an initiative document using an outdated or unofficial template.

---

## Consequences of Non-Compliance

Failure to adhere to this policy may result in misaligned development efforts, increased technical debt, project delays, and wasted company resources. Violations are a serious matter and may subject the responsible employee(s) to disciplinary action, ranging from coaching to formal measures.
