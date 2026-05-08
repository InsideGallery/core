# Plan Task Authoring Guide (agent)

This guide defines the `docs/aics/plan.md` task format consumed by `./agent.sh`.

Keep format rules in this file. `docs/aics/plan.md` should only contain the plan content and a short link back to this
guide.

## Required Task Block Format

```md
### TODO <TASK-ID> (<SOURCE-STATUS>): <Short Title>

- WHAT: <Clear implementation scope.>
- WHERE:
  Layer `domain`: <files/modules>
  Layer `repository`: <files/modules>
  Layer `handler`: <files/modules>
  Tests: <files/modules>
  Swagger/docs: <files/modules>
- WHY: <Why this is required / risk if skipped.>
- References: `<doc/path-1>`, `<doc/path-2>`, `<doc/path-3>`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: TODO.
```

## Status Lifecycle (In-Block)

Status is tracked by the `- Status:` line inside each TODO block (not in heading text).

- `- Status: TODO.` means queued.
- `- Status: TAKEN.` means currently being executed.
- `- Status: DONE.` means completed.
- Optional manual status: `- Status: BLOCKED.`

`agent.sh` transitions status as follows:

1. Finds the first block where:
   - Heading starts with `### TODO `
   - `Backward compatibility` is `Yes`
   - `Status` is `TODO`
2. Changes status to `TAKEN` before execution.
3. Runs Codex for that one task block.
4. Expects Codex to update status to `DONE` when finished.
5. If Codex exits with failure, runner reverts status back to `TODO`.
6. If Codex succeeds but leaves status not `DONE`, runner enforces `DONE`.

## Backward Compatibility Rule

- Runner executes only TODO blocks with `- Backward compatibility: Yes ...`.
- Blocks with `- Backward compatibility: No ...` are skipped.
- Keep this line explicit and mandatory in every TODO block.

## Common Skip Cases

`### TODO ...` in the heading is not enough. A block is skipped when any of these are true:

- Missing `- Backward compatibility: ...`.
- Missing `- Status: ...`.
- `Backward compatibility` is not `Yes`.
- `Status` is not exactly `TODO` after trimming spaces and a trailing period.
- The heading does not start with `### TODO `.

When the runner finishes immediately, check the startup log line:

```text
Session started — total TODO blocks: <n>, eligible TODO blocks: <m>, skipped blocks: <k>
```

If `eligible TODO blocks` is `0`, normalize the metadata lines before changing the runner.

## Gap Register Sync

When `docs/aics/gap.md` is used as an input source, every `### TODO ...` gap must have a matching task block in
`docs/aics/plan.md`.

The `plan.md` copy must preserve:

- The same task ID and heading.
- The same `WHAT`, `WHERE`, and `WHY` content.
- Runner metadata: `References`, `Backward compatibility`, and `Status`.

Use `docs/aics/gap.md` as the evidence register and `docs/aics/plan.md` as the executable queue.

## Authoring Rules

1. Use unique `<TASK-ID>` values (for example: `WA240-VER-03`).
2. Keep exactly one task per `### TODO` block.
3. Keep headings at level `###`.
4. Keep section keys exact: `WHAT`, `WHERE`, `WHY`, `References`, `Backward compatibility`, `Status`.
5. Keep `Status` values uppercase: `TODO`, `TAKEN`, `DONE`, `BLOCKED`.
6. In `WHERE`, list exact files/modules expected to change.
7. In `References`, include all source docs required for implementation.
8. Prefer ending `Status` with a trailing period for consistency (`- Status: TODO.`).
9. Do not place runner instructions inside `plan.md`; update this file instead.

## Minimal Example

```md
### TODO WA999-VER-01 (MISSING): Add endpoint parity for sample flow

- WHAT: Add `/v2/sampleapi/items/{id}` read endpoint with legacy-compatible response.
- WHERE:
  Layer `handler`: `services/sampleapi/internal/handler/item.go`, `router.go`.
  Layer `domain`: `services/sampleapi/internal/domain/item/{port,service}.go`.
  Layer `repository`: `services/sampleapi/internal/repository/item.go`.
  Tests: `services/sampleapi/internal/{handler,domain,repository}/item_test.go`.
  Swagger/docs: `services/sampleapi/docs/swagger.json`.
- WHY: Migration parity and consumer cutover are blocked without this route.
- References: `docs/epics/initiative-99/aic.md`, `docs/epics/initiative-99/arc42.md`.
- Backward compatibility: Yes (additive/parity-preserving change).
- Status: TODO.
```
