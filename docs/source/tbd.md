# Trunk-Based Development for Go Services

**Source**: Paul Hammant et al., [trunkbaseddevelopment.com](https://trunkbaseddevelopment.com/) — distillation of practices used at Google, Facebook, and other high-throughput engineering organizations. Reinforced by *Accelerate* (Forsgren, Humble, Kim — 2018) DORA findings: TBD correlates strongly with elite delivery performance.
**Adaptation**: Go-service specifics — module/package layout, CI gates, feature-flag patterns, release tags, monorepo and multi-repo variants. Aligned with Engineering Principles (POL-ENG-001), Twelve-Factor App, Go Server.md, Go Library.md, and Policy of Initiatives (POL-TECH-001).

---

## Philosophy

Trunk-Based Development (TBD) is a source-control branching model in which **all developers commit to a single branch** ("trunk", typically `main`) at high frequency, behind continuous integration. Long-lived branches are forbidden. Releases are cut from trunk, never developed on side branches.

**Core beliefs:**
- Long-lived branches are the single largest source of integration pain. The cost of merging grows quadratically with branch lifetime.
- Code that is not integrated does not exist. Working in isolation hides conflicts, dead code, and stale assumptions.
- The release cadence is decoupled from the merge cadence by **feature flags**, not by branches.
- A green trunk is a shared, non-negotiable invariant. If trunk is broken, everyone stops to fix it.

**The TBD Bargain:**
- You give up: long feature branches, big-bang merges, "I'll integrate later."
- You get: continuous integration in the literal sense, fast feedback, near-zero merge conflicts, releasable trunk at all times, and the ability to deploy any commit.

**When NOT to use TBD:**
- Open-source projects with untrusted contributors — use a fork-and-PR variant (still short-lived branches, still merge to trunk often).
- Regulated systems requiring formal pre-merge review of every line — combine TBD with **short-lived branches** (≤ 2 days) plus mandatory PR review.

---

## The Two Variants

| Variant | Team Size | Mechanism | When |
|---------|-----------|-----------|------|
| **Committed TBD** ("pure trunk") | < ~10 active devs per service | Direct commits to `main`. Pre-commit hooks + CI. | Small teams, mature CI, strong test culture. |
| **Short-Lived Feature Branches** | Any size | Branch from `main`, merge within 1–2 days, delete. PR review optional but common. | Default for most Go service teams. The recommended starting point. |

> Both variants are TBD. The defining property is **branch lifetime**, not whether branches exist.

**Hard rules for short-lived branches:**
- Lifetime: ≤ 2 working days. 1 day is the target.
- Diff size: ≤ 400 lines changed (excluding generated code, vendored deps, lockfiles).
- One reviewer, one round of feedback. If review takes more than a day, the change was too big.
- Rebased onto `main` immediately before merge. Squash-merge if commit history is messy; merge-commit if commits are clean and meaningful.

---

## Cadence and Rhythm

| Cadence | Target |
|---------|--------|
| Pulls from `main` | At least once per day, ideally before every push. |
| Pushes to `main` | At least once per day per active developer. |
| CI build time on trunk | < 10 minutes hard ceiling, < 5 minutes ideal. Slow CI breaks TBD. |
| Time from green local to merged | < 1 hour. |
| Trunk-broken-to-fixed | < 10 minutes. The team's first priority. |

If you cannot hit these numbers, TBD will degrade into "main is just our integration branch." Fix the bottleneck (test speed, flaky tests, slow review) before blaming the model.

---

## Branching Rules

### What's allowed on trunk
- Fully working code, even if **incomplete behind a flag**.
- Refactors that compile and pass tests.
- New packages, types, and exported APIs that aren't yet wired up — provided they don't break the build.

### What's forbidden on trunk
- Broken builds. Period.
- Failing tests (skipped tests with a tracking ticket are tolerable; commented-out tests are not).
- `// TODO: fix before release` without a flag gating the code.
- Commented-out code blocks. Delete; git remembers.
- Generated artifacts that aren't checked in by policy.

### Branch types and their lifetimes

| Branch | Purpose | Lifetime | Notes |
|--------|---------|----------|-------|
| `main` | Trunk | Forever | The only source of truth. |
| `feature/<short-name>` | Short-lived integration branch | ≤ 2 days | Optional. Delete on merge. |
| `release/v1.42` | Release stabilization | Days–weeks | Cut from `main` at a chosen commit. **Cherry-pick fixes from `main` into release; never merge release back into `main`.** |
| `hotfix/<id>` | Urgent production fix on a release line | Hours | Branch from `release/v1.42`, merge to that release, cherry-pick to `main`. |

**Never:**
- `develop` branches (GitFlow). TBD replaces GitFlow entirely.
- `personal/<name>` long-lived branches.
- `experimental/...` branches that live more than 2 days. Experiment behind a flag on trunk.

---

## Release Strategies

TBD decouples **merge** from **release**. Pick the release model that fits your service:

### Continuous Delivery (CD)
- Every green commit on `main` is automatically deployed to staging.
- Production deploys are triggered manually (button push) but use the same artifact.
- **Recommended default** for internal Go services.

### Continuous Deployment
- Every green commit on `main` is automatically deployed to production.
- Requires: progressive rollout, feature flags, automated rollback on SLO breach, mature observability.
- Right for mature platform teams. Not the starting point.

### Release Branches (long-lived support)
- Cut `release/v1.42` from `main` at a chosen commit.
- Bug fixes are committed to `main` first, then cherry-picked to the release branch.
- Tag releases on the branch: `v1.42.0`, `v1.42.1`.
- Required for: vendored Go libraries with multiple supported versions, on-prem product distributions.

**Tag scheme** (Go module convention, SemVer):
- `v1.42.0` — first release on `release/v1.42`.
- `v1.42.1` — patch (bug fix only).
- `v1.43.0` — next minor (cut a new release branch).
- `v2.0.0` — breaking change. **Requires `/v2` module path** per Go Modules SIV rule. See Go Library.md.

---

## Feature Flags: The Engine of TBD

> Feature flags are how TBD ships incomplete features safely. Without flags, TBD degrades into "everyone breaks each other's half-built features."

### Flag taxonomy

| Type | Lifetime | Purpose | Example |
|------|----------|---------|---------|
| **Release flag** | Days–weeks | Hide an incomplete feature until it ships. **Remove after launch.** | `enable_new_checkout` |
| **Experiment flag** | Weeks | A/B test or canary. Feeds a metric. **Remove when experiment concludes.** | `checkout_v2_canary_pct` |
| **Ops flag** (kill switch) | Long-lived | Disable a subsystem under load. Owned by SRE. | `disable_recommendations` |
| **Permission flag** | Long-lived | Customer/tenant entitlements. **Not really a TBD flag** — it's a domain feature. | `tenant_allows_sso` |

### Rules for release flags (the only kind TBD requires)

- **Default off** until the feature is feature-complete and tested.
- **Two code paths must both pass tests.** Don't let the off-path rot.
- **One owner, one ticket, one removal date.** A flag without a removal plan is technical debt accruing interest.
- **Remove the flag immediately after the feature is at 100%.** Open a PR with title `chore: remove flag <name>`.
- **Cap on outstanding flags.** Recommend ≤ 1 flag per active developer at any time. More than that and the codebase becomes a maze of conditionals.

### Flag implementation in Go

```go
// internal/featureflag/flag.go
package featureflag

import "context"

type Provider interface {
    Bool(ctx context.Context, name string, def bool) bool
}

// Compose with context (per-request) for tenant/user-scoped flags.
type ctxKey struct{}

func WithProvider(ctx context.Context, p Provider) context.Context {
    return context.WithValue(ctx, ctxKey{}, p)
}

func Bool(ctx context.Context, name string, def bool) bool {
    p, ok := ctx.Value(ctxKey{}).(Provider)
    if !ok {
        return def
    }
    return p.Bool(ctx, name, def)
}
```

```go
// Use site
func (s *CheckoutService) Place(ctx context.Context, o Order) error {
    if featureflag.Bool(ctx, "enable_new_checkout", false) {
        return s.placeV2(ctx, o)
    }
    return s.placeV1(ctx, o)
}
```

**Anti-patterns:**
- Reading flags from globals (`flag.Bool(...)` at init). Untestable. Couples flag lifecycle to process lifecycle.
- Flag checks scattered across 30 files for one feature. Centralize the branch point.
- Flags read inside hot loops without caching. Cache per-request.
- "Temporary" flags older than 90 days. They are not temporary.

---

## Branch by Abstraction (large refactors on trunk)

For changes too large to ship in a 2-day branch (e.g., swapping a database driver, replacing a caching layer):

1. **Introduce an abstraction** in front of the existing implementation. Land it on trunk.
2. **Add the new implementation behind the same abstraction**. Land it on trunk.
3. **Wire callers to the new implementation incrementally**, one by one. Each is a small PR.
4. **Flip the default** behind a flag. Soak.
5. **Remove the old implementation**. Remove the flag. Remove the abstraction if it has only one implementation now.

This is the TBD answer to "but my refactor is too big for a 2-day branch." It is always too big for a 2-day branch as one PR. Decompose it.

---

## CI Requirements

TBD without strong CI is reckless. The CI pipeline is the load-bearing wall:

| Stage | Required | Target Time |
|-------|----------|-------------|
| `gofmt -l` / `goimports` check | Yes | < 5s |
| `go vet` | Yes | < 30s |
| `staticcheck` (or equivalent) | Yes | < 60s |
| Unit tests `go test ./...` with `-race` | Yes | < 3 min |
| Build all binaries `go build ./cmd/...` | Yes | < 60s |
| Integration tests (real DB, real broker) | Yes for services | < 5 min |
| End-to-end / smoke tests | On trunk post-merge | < 10 min |
| Coverage gate | Optional, recommended ≥ 70% on changed lines | — |

**Pre-merge gates** (block merge):
- All required stages green.
- No lint warnings (or explicitly waived with a tracked ticket).
- Branch is rebased onto current `main` (no stale merges).

**Post-merge gates** (alert and stop the line):
- E2E suite green.
- Deploy to staging succeeded.
- Smoke health checks passed.

If post-merge fails, **the team's first priority is to get trunk green**. Revert first, fix forward later.

### Stop-the-line rule

When CI on `main` fails:
1. The breaker (or the on-call) reverts within 10 minutes — `git revert <sha>` and merge.
2. No new merges to `main` until green.
3. The original change is fixed locally and re-merged.

Do not "fix forward" on a broken trunk. Reverting is cheap; debugging compounding failures is not.

---

## Code Review in TBD

TBD does not eliminate review; it constrains it.

| Practice | TBD-aligned |
|----------|-------------|
| **Review size** | ≤ 400 changed lines. Larger reviews don't get reviewed; they get rubber-stamped. |
| **Review latency** | < 4 hours during working hours. A PR open for a day is a TBD failure. |
| **Reviewer count** | One is sufficient. Two reviewers slows merge without proportional defect reduction (per *Accelerate*). |
| **Pre-commit review** | Optional. **Pair programming** counts as review. |
| **Post-commit review** ("ship and review") | Acceptable for trusted committers on low-risk changes (formatting, comments, internal refactors). Not for public APIs. |

---

## Repository Layout (Go specifics)

### Single-service repo
```
.
├── cmd/<service>/main.go
├── internal/...
├── pkg/... (only if exposing libraries)
├── go.mod
├── go.sum
└── .github/workflows/ci.yml
```

- One `go.mod` at the root.
- All commits go to `main`. Tags `vX.Y.Z` mark releases.
- See **Go Server.md** for full layering rules.

### Monorepo (multiple Go services)
```
.
├── services/<service-a>/
├── services/<service-b>/
├── libs/<shared-lib>/
├── go.work
└── ...
```

- Use `go.work` for local development; each service/lib has its own `go.mod` for independent versioning.
- TBD applies to the **whole monorepo** — one trunk, one CI matrix.
- Releases per service: tag `services/service-a/v1.2.0` (Go submodule tag convention).

### Vendored library repo
```
.
├── pkg/<lib>/
├── go.mod
└── ...
```

- TBD on `main`.
- Maintain `release/v1`, `release/v2`, etc., for **supported major versions only**. See Go Library.md SemVer rules.

---

## Anti-Patterns

| Anti-Pattern | Why It Breaks TBD | Fix |
|--------------|-------------------|-----|
| GitFlow `develop` branch | Adds an integration layer that defeats trunk-based integration. | Delete `develop`. Merge to `main`. |
| Long-lived feature branches | Merge pain compounds; trunk diverges from reality. | Split into incremental commits behind a flag. |
| "Stabilization" branches | Implies trunk is unstable, which means TBD has already failed. | Make trunk releasable continuously. |
| PRs > 1000 LOC | Unreviewable. | Decompose. Branch by abstraction. |
| Flaky tests tolerated on trunk | Erodes the green-trunk invariant. | Quarantine immediately, fix or delete within a week. |
| Skipping CI to merge "small" changes | The exception becomes the rule. | Never. CI is the contract. |
| Flags older than 90 days | Becomes part of the architecture by accident. | Quarterly flag audit; remove or formalize as config. |
| Cherry-picking from release back to main | Risks losing fixes when paths diverge. | Always commit to `main` first, then cherry-pick to release. |
| Force-pushing to `main` | Rewrites shared history; breaks every other developer. | Forbidden. Branch protection enforces. |

---

## Adoption Path (for a team migrating from GitFlow / long branches)

1. **Week 1** — Turn on branch protection for `main`: required CI, no force-push, no direct admin override.
2. **Week 1** — Set a **branch lifetime ceiling** (start at 5 days, ratchet down weekly).
3. **Week 2** — Cap PR size at 800 LOC, then 400.
4. **Week 2–4** — Introduce a feature-flag library. First flag for the next non-trivial feature.
5. **Week 4** — Eliminate `develop`. Cut releases from `main`.
6. **Month 2** — Reduce CI runtime below 10 min. This usually requires test parallelization and removing slow integration tests from the pre-merge stage.
7. **Month 2** — Begin "branch by abstraction" for the largest pending refactor.
8. **Month 3** — Audit flags. Remove every flag whose feature shipped. Set per-developer flag cap.

Measure: branch lifetime, time-to-merge, CI runtime, time-to-restore-trunk, flag count. These are the leading indicators.

---

## TBD and the Workspace Methodologies

| Cross-cutting concern | TBD interaction |
|-----------------------|-----------------|
| **Twelve-Factor App** | Build/release/run separation aligns with TBD: trunk produces immutable artifacts; releases are flag-toggle config swaps. |
| **MDCA / DDD** | Bounded contexts → packages → independent CI subsets in monorepos. Smaller scopes → smaller PRs → easier TBD. |
| **Engineering Principles (POL-ENG-001)** | DRY and KISS support short PRs. No speculative abstraction supports decomposing changes. |
| **Policy of Initiatives (POL-TECH-001)** | An initiative's plan.md should be merged to trunk in stages, each gated by a flag where appropriate. No "initiative branch." |
| **Clean Code** | Small functions and clear boundaries are what make 400-LOC PRs feasible. |
| **SOLID** (`solid.md`) | OCP via interfaces makes "branch by abstraction" mechanical: define the interface, implement variants in parallel commits. |

---

## Quick Self-Check

Before pushing to `main` (or merging a short-lived branch):

1. Has CI run locally and passed (`go test -race ./... && go vet ./... && staticcheck ./...`)?
2. Is the diff < 400 lines of changed Go (excluding generated)?
3. Is the branch ≤ 2 days old?
4. Is incomplete behavior gated behind a flag, default-off?
5. Is the rebase onto `main` clean — no merge commits introduced?
6. Will trunk still be deployable to production after this commit?

If any answer is no, you are not doing TBD. Fix the answer, not the rule.

---

## Further Reading

- [trunkbaseddevelopment.com](https://trunkbaseddevelopment.com/) — Paul Hammant's reference site (canonical source for this document).
- *Accelerate* (Forsgren, Humble, Kim, 2018) — DORA research linking TBD to elite delivery performance.
- *Continuous Delivery* (Humble, Farley, 2010) — the deployment-pipeline foundation TBD assumes.
- Martin Fowler — "FeatureToggle" (martinfowler.com) — flag taxonomy and lifecycle.
- Cross-references in this workspace: **Twelve-Factor App.md**, **Engineering Principles.md**, **Go Server.md**, **Go Library.md**, **Policy of Initiatives.md**, **Clean Code.md**, **solid.md**.
