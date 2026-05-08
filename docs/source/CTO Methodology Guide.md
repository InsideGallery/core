# CTO Methodology Guide

**Source**: Will Larson, "The Engineering Executive's Primer" (2024)
**Purpose**: Extracted methodologies, frameworks, and decision models for use in initiative planning and engineering leadership.

---

## 1. Engineering Strategy (Rumelt Framework)

**When**: Within first 6 months of role, update annually.

A strategy has three parts:

| Part | Purpose | Key Questions |
|------|---------|---------------|
| **Diagnosis** | Theory describing the root cause of the problem | What is actually happening? What data confirms it? |
| **Guiding Policy** | Approach to address diagnosed problems, with explicit trade-offs | How are resources allocated? What rules must all teams follow? How are decisions made? |
| **Coherent Actions** | Concrete steps to implement the policy | What enforcement mechanisms exist? What transformations are needed? |

**Quality tests for Guiding Policy**:
- **Applicability**: Can it navigate real-world trade-offs?
- **Enforceability**: Will teams actually be held to it?
- **Effectiveness**: Does it create a multiplier effect?

**Process** (10 steps):
1. Write it yourself (do not delegate)
2. Write for engineering leaders (staff+ engineers, senior managers)
3. Define stakeholder list, select 3-5 for a working group
4. Draft diagnosis first, iterate one-on-one with working group
5. Draft guiding policy, get feedback from 2-3 external peers
6. Share diagnosis + policy with all stakeholders
7. Draft coherent actions, iterate with working group
8. Talk individually to likely dissenters
9. Present to full engineering org, 1-week feedback window
10. Finalize, promise to evaluate in 2 months, update annually

---

## 2. Three-Phase Planning

**When**: Annual and quarterly planning cycles.

| Phase | Scope | Key Rule |
|-------|-------|----------|
| **1. Financial Plan** | P&L, budget, headcount | Keep fixed for the year; frequent revisions destroy accountability |
| **2. Resource Allocation** | Distribute capacity (e.g., product 63%, infra 25%, DevEx 12%) | Keep stable; frequent changes destroy morale |
| **3. Roadmap Alignment** | Cross-functional agreement on deliverables | Separate proven (80%) from unproven (20%) projects |

**Org design math**: Teams of ~8, groups of 4-6 teams, continue until 5-7 groups report to you.

**Anti-patterns**:
- Planning as checkbox exercise (plans never revisited)
- Headcount as universal cure
- Only "exciting" work gets prioritized
- Over-detailed plans removing team autonomy

---

## 3. Three Management Styles

Every engineering executive must master all three and switch between them:

### Policy-Based
**When**: Recurring decisions (reviews, promotions, vendor selection).
1. Identify frequently-made decisions
2. Study how the best performers make them
3. Document methodology
4. Test on small group, then roll out
5. Schedule revision

### Consensus-Based
**When**: Multi-stakeholder decisions with distributed context.
1. Confirm no policy exists, multiple stakeholders hold context
2. Apply "will I remember this in 6 months?" test
3. Identify all stakeholders, create shared channel
4. Draft framing document
5. Follow the most invested party's lead on reaching agreement

### Conviction-Based
**When**: High-stakes, non-recurring decisions with deep uncertainty.
1. Talk to domain experts to build your mental model
2. Write a decision document
3. Discuss with trusted advisors (including external network)
4. Announce preliminary decision with ~1 week comment period
5. Make final decision, document reasoning, execute

**Micromanagement test** (2 questions):
- "Do I consider opinions of those most involved?" (Should be YES)
- "Am I adding complexity for a team that would have decided similarly?" (Should be NO)

---

## 4. Trust, But Verify

**Principle**: Blind trust is not a management technique. It prevents distinguishing good processes with bad outcomes from bad processes with good outcomes.

**Four verification instruments**:

| Instrument | Method |
|-----------|--------|
| **Review meetings** | Weekly/monthly metric reviews (quarterly is too infrequent) |
| **Primary source investigation** | Talk directly to people doing the work, bypass summary layers |
| **Direct data work** | Maintain your own small data sources, cross-reference |
| **Intolerance for inconsistencies** | When something doesn't add up, dig in immediately |

**Cycle**: Trust -> Verify -> Return with findings -> Solve together.

---

## 5. Engineering Metrics (Four Segments)

| Segment | Purpose | Examples |
|---------|---------|---------|
| **Plan** | Track engineering's business impact | Completed projects, impact per project |
| **Operate** | Monitor system health | Incidents, latency, cost per metric, app ratings |
| **Improve** | Measure developer productivity | DORA/SPACE: deploy frequency, lead time, change failure rate, MTTR |
| **Inspire** | Showcase innovation | List of technical achievements that turned impossible into obvious |

**Stakeholder-specific**:
- **CEO/Board**: Planning or operational metrics (never optimization metrics)
- **Finance**: Headcount vs budget, vendor cost, CapEx/OpEx
- **Product/Sales**: Planning metrics showing business impact
- **Support/Legal**: SLA-type metrics (ticket rate, resolution time)

**Anti-patterns**: Metrics as substitute for trust; chasing ideal measurement; evaluating individuals instead of teams.

---

## 6. Meeting Calendar Framework

For organizations of 20-200 engineers:

### Weekly
| Meeting | Attendees | Format |
|---------|-----------|--------|
| **Engineering Leadership** | Direct reports + key partners (recruiting, HR, finance) | Shared editable agenda, drive alignment |
| **Technical Review** | Engineers + architects | Written specs read before meeting, discuss not present |
| **Incident Review** | On-call engineers + managers | Written documents first, cancellable in quiet weeks |

### Monthly
| Meeting | Attendees | Format |
|---------|-----------|--------|
| **Engineering Managers** | All eng managers | 15min round-robin + 30min development topic + 15min Q&A |
| **Staff Engineers** | Staff+ engineers (without managers) | Same format, development from "Staff Engineer's Path" |
| **Engineering Q&A** | All engineers | New hires intro, key messages, open floor, anonymous questions |

---

## 7. Internal Communication (5 Habits)

| Habit | Description |
|-------|-------------|
| **Drip communication** | Weekly update email: human opener, reminders, 2-3 topic paragraphs, highlights, invite questions (~20 min/week) |
| **Pre-flight check** | Before any mass communication, have at least one person review it |
| **Full package** | Every message needs: TL;DR, link to source of truth, where to ask questions |
| **Keep it short** | Edit relentlessly for brevity |
| **Multi-channel** | Distribute through ALL channels: email, chat, meetings, notes, decision logs |

**"Look deeper" rule**: When a CEO makes a specific suggestion, the real message is about a deeper concern. Address the root, not the surface.

---

## 8. Standards Calibration (Show. Document. Share.)

**When**: Your standards exceed those of peers and you want to influence without authority.

1. **Show** -- Demonstrate the desired standard yourself through several iterations
2. **Document** -- Create a clear document explaining how to replicate, framing benefits from the audience's perspective
3. **Share** -- Send to teams you want to influence; engage the interested, don't pressure the rest

**Key insight**: Telling the CEO about a peer's low standards is really telling them they failed to solve a known problem. Escalate cautiously.

---

## 9. Hiring System (9 Components)

| Component | Purpose |
|-----------|---------|
| ATS | Candidate tracking and coordination |
| Interview scripts + rubrics | Standardized questions with evaluation criteria |
| Interview loop documentation | Cycle, procedures, prepared interviewers |
| Scorecards | Formalized candidate evaluation |
| RACI | Who decides on offer, who approves compensation |
| Job description templates | Reusable, standardized |
| Interviewer training | Shadow -> Reverse-shadow -> Independent |
| Level framework | Pre-level after phone screen, confirm before offer |
| Compensation bands | Compa-ratio target 0.95 for new hires, cap at 1.1 |

**Warning signs of over-optimization**: Recruiters hire <5/quarter; >2 weeks from first interview to offer.

---

## 10. Onboarding (Four Roles)

| Role | Responsibility |
|------|---------------|
| **Sponsor** (you) | Choose orchestrator, track monthly, meet outlier new hires |
| **Orchestrator** | Design curriculum, evolve based on feedback, 20+ hours/month |
| **Manager** | Personalized onboarding doc, weekly 1:1s, assigns buddy, selects first project |
| **Buddy** | Daily 15-30 min meetings first weeks, takes to lunches/meetings |

**Starter curriculum**: Engineering values/strategy, technical architecture, dev environment setup.

---

## 11. Performance and Compensation

**Review calibration** (for several hundred engineers):
1. Managers present preliminary ratings + promotion nominations
2. Group of 5-8 managers reviews each nomination (3-5 hours)
3. Direct reports to eng leader review with leader
4. Eng leader reviews with HR, aligns with peers, checks budget

**Compensation**: Benchmark via agencies, use compa-ratio, geographic tiers, budget always wins over optimization.
**Frequency**: Semi-annual reviews, annual compensation/promotion decisions.

---

## 12. Culture Surveys (10-Step Protocol)

**When**: Organization reaches ~150 people. Run twice a year.

1. Get full access to company-wide results
2. Create personal analysis document
3. Assess sample sizes
4. Group into: celebrate / proactively fix / monitor
5. Check if previous improvements actually improved
6. Focus on highest/lowest absolute scores
7. Focus on fastest-changing metrics
8. Analyze differences by group (manager, tenure, location)
9. Read every comment
10. Discuss 1 hour with a peer leader

**Action**: Pick 2-3 areas with actionable solutions. Track monthly. Reference achievements when launching next survey.

---

## 13. Technology Standardization (Standardize-by-Default)

**Rule**: Standardize by default. Research only when the improvement is **at least 10x** on one dimension without degrading others.

- Fewer technologies = deeper investment in each
- Cap concurrent research experiments (2-3 for ~1000 engineers)
- Binary outcome per research: Continue (migrate from old) or Stop (abandon cleanly)
- Over-standardization creates dead ends; over-research creates abandoned half-migrations

---

## 14. Priority Framework (Company, Team, Me)

**Default**: Prioritize company > team > personal interests.
**Evolution**: "Eventual Quid Pro Quo" -- mostly follow the hierarchy, but periodically prioritize energy-restoring work. If depletion persists >1 year, the role needs changing.

**Test**: Energy-restoring work should be "orthogonal but not opposite" to company needs (1-2 conference talks/year is fine; 8-10 is too many).

---

## 15. Corporate Values (Quality Test)

A useful value must pass three tests:

| Test | Question |
|------|----------|
| **Has an alternative** | Can you invert it and still get something reasonable? |
| **Practical significance** | Can you use it to navigate real trade-offs? |
| **Actually practiced** | Does it describe how people actually behave? |

**Recommended engineering values**:
- "Create new opportunities rather than fight over existing ones"
- "Always go to vendors unless it's a core competency"
- "Follow existing patterns unless there's a 10x improvement"
- "Be curious in conflict resolution"

**Process**: Wait 6+ months before introducing. First change behavior, then document as a value.

---

## 16. First 90 Days Framework

**Focus areas** (in order):
1. Learn the business (revenue model, culture, decision-making)
2. Build trust ("audio tour" -- meet everyone, weekly status emails)
3. Build external support network (peer communities, coach)
4. Understand org processes (document existing, change max 1-2)
5. Evaluate hiring (funnel metrics, attend interviews, max 3 key hires)
6. Understand technology (review strategy, make a small code change, attend incident reviews)

**Anti-patterns**: Rushing changes before understanding; judging without context; constantly referencing previous employer.

---

## 17. Executive Departure (Structured Process)

**Decision heuristics** (ask yourself):
1. Has learning speed declined significantly?
2. Are you consistently drained? (Keep energy journal for one quarter)
3. Do you still believe what you say when selling candidates on joining?
4. Will departure be more disruptive in 6 months?

**Continuous preparation**: Identify succession gaps during reviews; delegate meetings quarterly; take 2-week vacations annually with full delegation.

---

## 18. Hub Office Launch (5 Pillars)

| Pillar | Requirement |
|--------|-------------|
| **Mission** | End-to-end ownership of a business line; no mission = no office |
| **Leadership** | Appoint hub director early; visit quarterly for 1-2 years; start with ~12 people |
| **Predictability** | One sprint year, then ~50% annual growth; no freeze/unfreeze |
| **Integration** | Send 3-5 HQ veterans for 1-2 years; equal access to Staff+ roles |
| **Naming** | Never "remote office" -- use "hub" or "representation" |

---

## Key References

| Book | Author | Focus |
|------|--------|-------|
| Good Strategy Bad Strategy | Richard Rumelt | Strategy formulation |
| The First 90 Days | Michael Watkins | Leadership transitions |
| Accelerate | Forsgren, Humble, Kim | DevOps metrics (DORA) |
| High Output Management | Andrew Grove | Management fundamentals |
| The Manager's Path | Camille Fournier | Tech leadership growth |
| Thinking in Systems | Donella Meadows | Systems thinking |
| An Elegant Puzzle | Will Larson | Engineering management systems |
| INSPIRED / EMPOWERED | Marty Cagan | Product management |
| The Staff Engineer's Path | Tanya Reilly | Staff+ career growth |
