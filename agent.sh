#!/usr/bin/env bash
set -euo pipefail

# === Configuration — adjust per project ===
PLAN="docs/aics/plan.md"          # Master action plan (new TODO-block format)
LOG="agent.log"          # Detailed execution log
OUT="agent.out"          # Metrics summary (cost, time, tokens)
ERR="agent.err"          # Captured Codex stderr stream

# Codex CLI configuration.
# Override via environment if your local install uses a different binary or flags.
CODEX_BIN="${CODEX_BIN:-codex}"
CODEX_EXTRA_ARGS="${CODEX_EXTRA_ARGS:---dangerously-bypass-approvals-and-sandbox}"
CODEX_RETRY_EXIT_101="${CODEX_RETRY_EXIT_101:-2}"
CODEX_RETRY_BACKOFF_SEC="${CODEX_RETRY_BACKOFF_SEC:-2}"

# --- Metrics tracking ---
TOTAL_TASKS=0
TOTAL_COST=0
TOTAL_DURATION=0
TOTAL_INPUT_TOKENS=0
TOTAL_OUTPUT_TOKENS=0
SESSION_START=$(date +%s)
PLAN_TASKS_DONE=0

RUNNER_MODE=""
RUNNER_JSON=0
RUNNER_HELP=""
TMPJSON=""
TMPRAW=""
TMPERR=""

init_metrics() {
  cat > "$OUT" <<HEADER
# Agent Codex Metrics — $(date '+%Y-%m-%d %H:%M:%S')
# Format: task_num | timestamp | task_name | duration_s | cost_usd | input_tokens | output_tokens | cumulative_cost

HEADER
}

log() {
  echo "=== $(date '+%Y-%m-%d %H:%M:%S') — $1 ===" | tee -a "$LOG"
}

record_task() {
  local task_name="$1"
  local json_file="$2"

  if [ ! -f "$json_file" ]; then
    echo "$(( ++TOTAL_TASKS )) | $(date '+%H:%M:%S') | ${task_name:0:80} | ? | ? | ? | ? | $TOTAL_COST" >> "$OUT"
    return
  fi

  local metrics cost duration input_tokens output_tokens

  metrics=$(
    jq -Rrs '
      [split("\n")[] | select(length > 0) | fromjson?] as $events |
      {
        cost: (
          $events
          | map(.total_cost_usd // .cost.total_usd // .usage.total_cost_usd // empty)
          | last // 0
        ),
        duration_ms: (
          $events
          | map(.duration_ms // .timing.duration_ms // .metrics.duration_ms // empty)
          | last // 0
        ),
        input_tokens: (
          $events
          | map(
              if .type == "turn.completed" then
                (.usage.input_tokens // 0) + (.usage.cached_input_tokens // 0)
              else
                (.usage.input_tokens // .usage.prompt_tokens // 0) +
                (.usage.cache_creation_input_tokens // 0) +
                (.usage.cache_read_input_tokens // 0)
              end
            )
          | last // 0
        ),
        output_tokens: (
          $events
          | map(.usage.output_tokens // .usage.completion_tokens // 0)
          | last // 0
        )
      } |
      [.cost, .duration_ms, .input_tokens, .output_tokens] |
      @tsv
    ' "$json_file" 2>/dev/null || printf '0\t0\t0\t0\n'
  )

  IFS=$'\t' read -r cost duration input_tokens output_tokens <<< "$metrics"

  local duration_s
  duration_s=$(awk "BEGIN {printf \"%.1f\", $duration / 1000}")

  TOTAL_COST=$(awk "BEGIN {printf \"%.4f\", $TOTAL_COST + $cost}")
  TOTAL_DURATION=$(awk "BEGIN {printf \"%.0f\", $TOTAL_DURATION + $duration}")
  TOTAL_INPUT_TOKENS=$(( TOTAL_INPUT_TOKENS + input_tokens ))
  TOTAL_OUTPUT_TOKENS=$(( TOTAL_OUTPUT_TOKENS + output_tokens ))
  TOTAL_TASKS=$(( TOTAL_TASKS + 1 ))

  printf "%d | %s | %-80s | %7s s | \$%8s | %8d in | %7d out | \$%s cumul\n" \
    "$TOTAL_TASKS" "$(date '+%H:%M:%S')" "${task_name:0:80}" \
    "$duration_s" "$cost" "$input_tokens" "$output_tokens" "$TOTAL_COST" >> "$OUT"
}

write_summary() {
  local elapsed=$(( $(date +%s) - SESSION_START ))
  local hours=$(( elapsed / 3600 ))
  local mins=$(( (elapsed % 3600) / 60 ))

  cat >> "$OUT" <<SUMMARY

# === SESSION SUMMARY ===
# Tasks completed:  $TOTAL_TASKS
# Total cost:       \$$TOTAL_COST
# Total duration:   ${hours}h ${mins}m (wall clock)
# Input tokens:     $TOTAL_INPUT_TOKENS
# Output tokens:    $TOTAL_OUTPUT_TOKENS
# Avg cost/task:    \$$(awk "BEGIN {if ($TOTAL_TASKS>0) printf \"%.4f\", $TOTAL_COST/$TOTAL_TASKS; else print 0}")
# Avg time/task:    $(awk "BEGIN {if ($TOTAL_TASKS>0) printf \"%.0f\", $TOTAL_DURATION/1000/$TOTAL_TASKS; else print 0}")s
SUMMARY

  log "Session summary written to $OUT"
}

cleanup() {
  rm -f "$TMPJSON" "$TMPRAW" "$TMPERR"
  write_summary
}

trap cleanup EXIT

split_extra_args() {
  local raw="$1"
  local -n out_ref="$2"

  out_ref=()

  if [ -n "$raw" ]; then
    # shellcheck disable=SC2206
    out_ref=($raw)
  fi
}

detect_runner() {
  if ! command -v "$CODEX_BIN" >/dev/null 2>&1; then
    log "Codex binary not found: $CODEX_BIN"
    log "Set CODEX_BIN=/path/to/codex if it is installed under a different name."
    exit 1
  fi

  if "$CODEX_BIN" exec --help >/dev/null 2>&1; then
    RUNNER_MODE="exec"
    RUNNER_HELP=$("$CODEX_BIN" exec --help 2>&1 || true)
  else
    RUNNER_MODE="plain"
    RUNNER_HELP=$("$CODEX_BIN" --help 2>&1 || true)
  fi

  if printf '%s' "$RUNNER_HELP" | grep -q -- '--json'; then
    RUNNER_JSON=1
  fi

  log "Using Codex runner: mode=$RUNNER_MODE json=$RUNNER_JSON bin=$CODEX_BIN"
}

runner_supports_prompt_flag() {
  printf '%s' "$RUNNER_HELP" | grep -q -- '--prompt'
}

emit_runner_output() {
  local file="$1"

  if [ $RUNNER_JSON -eq 1 ]; then
    local rendered
    rendered=$(
      jq -Rr '
        fromjson? |
        (
          .result //
          .output_text //
          .assistant //
          (
            if .type == "item.completed" and .item.type == "agent_message" then
              .item.text
            else
              empty
            end
          )
        ) |
        select(type == "string" and length > 0)
      ' "$file" 2>/dev/null || true
    )

    if [ -n "$rendered" ]; then
      printf '%s\n' "$rendered" | tee -a "$LOG"
      return
    fi

    cat "$file" | tee -a "$LOG"
    return
  fi

  cat "$file" | tee -a "$LOG"
}

emit_runner_stderr() {
  local file="$1"
  local attempt="$2"

  if [ ! -s "$file" ]; then
    return
  fi

  {
    printf '=== %s — Codex stderr (attempt %s) ===\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$attempt"
    cat "$file"
    printf '\n'
  } | tee -a "$ERR" >> "$LOG"
}

run_codex() {
  local prompt="$1"
  local task_label="${2:-unknown}"
  local -a cmd extra_args
  local max_attempts attempt rc target_file

  log "Prompt: ${prompt:0:120}..."

  split_extra_args "$CODEX_EXTRA_ARGS" extra_args

  cmd=("$CODEX_BIN")
  if [ "$RUNNER_MODE" = "exec" ]; then
    cmd+=("exec")
  fi
  if [ $RUNNER_JSON -eq 1 ]; then
    cmd+=("--json")
  fi
  if [ ${#extra_args[@]} -gt 0 ]; then
    cmd+=("${extra_args[@]}")
  fi

  if runner_supports_prompt_flag; then
    cmd+=("--prompt" "$prompt")
  else
    cmd+=("$prompt")
  fi

  if [ $RUNNER_JSON -eq 1 ]; then
    target_file="$TMPJSON"
  else
    target_file="$TMPRAW"
  fi

  max_attempts=$(( CODEX_RETRY_EXIT_101 + 1 ))
  attempt=1

  while true; do
    : > "$target_file"
    : > "$TMPERR"

    set +e
    "${cmd[@]}" > "$target_file" 2> "$TMPERR"
    rc=$?
    set -e

    emit_runner_output "$target_file"
    emit_runner_stderr "$TMPERR" "$attempt"

    if [ $rc -eq 101 ] && [ $attempt -lt $max_attempts ]; then
      log "Codex exited with code 101 on attempt $attempt/$max_attempts; retrying in ${CODEX_RETRY_BACKOFF_SEC}s"
      attempt=$(( attempt + 1 ))
      sleep "$CODEX_RETRY_BACKOFF_SEC"
      continue
    fi

    break
  done

  if [ $RUNNER_JSON -eq 1 ]; then
    record_task "$task_label" "$target_file"
  else
    TOTAL_TASKS=$(( TOTAL_TASKS + 1 ))
    echo "$TOTAL_TASKS | $(date '+%H:%M:%S') | ${task_label:0:80} | ? | ? | ? | ? | $TOTAL_COST" >> "$OUT"
  fi

  if [ $rc -ne 0 ]; then
    log "Codex exited with code $rc"
  fi

  return $rc
}

# --- plan.md TODO-block helpers ---
# Required block format:
#   ### TODO <task-id> ...
#   ...
#   - Backward compatibility: Yes|No ...
#   - Status: TODO|TAKEN|DONE|BLOCKED.
# Task boundaries are from one `### TODO` heading to the next `### TODO` heading.
# Script processes ONLY blocks where Backward compatibility=Yes and Status=TODO.
count_todo_tasks() {
  grep -c '^### TODO ' "$PLAN" 2>/dev/null || echo 0
}

count_eligible_todo_tasks() {
  awk '
    BEGIN { in_task = 0; bc_yes = 0; task_status = ""; total = 0 }

    function commit_task() {
      if (!in_task) {
        return
      }

      if (bc_yes && toupper(task_status) == "TODO") {
        total++
      }
    }

    /^### TODO / {
      commit_task()
      in_task = 1
      bc_yes = 0
      task_status = ""
      next
    }

    in_task && /^- Backward compatibility:[[:space:]]*/ {
      line = tolower($0)
      bc_yes = (line ~ /^- backward compatibility:[[:space:]]*yes/)
    }

    in_task && /^- Status:[[:space:]]*/ {
      line = $0
      sub(/^- Status:[[:space:]]*/, "", line)
      sub(/[[:space:]]*\.$/, "", line)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)
      task_status = toupper(line)
    }

    END {
      commit_task()
      print total
    }
  ' "$PLAN" 2>/dev/null || echo 0
}

find_next_eligible_task() {
  awk '
    BEGIN { in_task = 0; bc_yes = 0; task_status = ""; found = 0 }

    function emit_task(next_start,   end_line) {
      if (!in_task || found) {
        return
      }

      if (bc_yes && toupper(task_status) == "TODO") {
        end_line = next_start - 1
        if (end_line < start_line) {
          end_line = start_line
        }
        print task_id "\t" start_line "\t" end_line "\t" task_header
        found = 1
      }
    }

    /^### TODO / {
      emit_task(NR)
      if (found) {
        exit
      }

      in_task = 1
      bc_yes = 0
      task_status = ""
      task_id = $3
      start_line = NR
      task_header = $0
      next
    }

    in_task && /^- Backward compatibility:[[:space:]]*/ {
      line = tolower($0)
      bc_yes = (line ~ /^- backward compatibility:[[:space:]]*yes/)
    }

    in_task && /^- Status:[[:space:]]*/ {
      line = $0
      sub(/^- Status:[[:space:]]*/, "", line)
      sub(/[[:space:]]*\.$/, "", line)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)
      task_status = toupper(line)
    }

    END {
      emit_task(NR + 1)
    }
  ' "$PLAN" 2>/dev/null
}

update_task_status() {
  local task_id="$1"
  local to_status="$2"
  local tmp

  tmp=$(mktemp /tmp/agent-plan-XXXXXX)

  if ! awk -v task_id="$task_id" -v to_status="$to_status" '
    BEGIN { in_target = 0; changed = 0 }

    /^### TODO / {
      in_target = ($0 ~ "^### TODO[[:space:]]+" task_id "([[:space:]]|$)")
    }

    in_target && /^- Status:[[:space:]]*/ {
      $0 = "- Status: " to_status "."
      changed = 1
      in_target = 0
    }

    {
      print
    }

    END {
      if (!changed) {
        exit 2
      }
    }
  ' "$PLAN" > "$tmp"; then
    rm -f "$tmp"
    return 1
  fi

  mv "$tmp" "$PLAN"
}

get_task_status() {
  local task_id="$1"

  awk -v task_id="$task_id" '
    BEGIN { in_target = 0 }

    /^### TODO / {
      if (in_target) {
        exit
      }
      in_target = ($0 ~ "^### TODO[[:space:]]+" task_id "([[:space:]]|$)")
    }

    in_target && /^- Status:[[:space:]]*/ {
      line = $0
      sub(/^- Status:[[:space:]]*/, "", line)
      sub(/[[:space:]]*\.$/, "", line)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)
      print toupper(line)
      exit
    }
  ' "$PLAN" 2>/dev/null
}

extract_task_block() {
  local start_line="$1"
  local end_line="$2"

  sed -n "${start_line},${end_line}p" "$PLAN"
}

do_plan_task() {
  local next_task task_id start_line end_line task_header task_name task_block prompt rc status_after

  next_task=$(find_next_eligible_task)
  [ -n "$next_task" ] || return 1

  IFS=$'\t' read -r task_id start_line end_line task_header <<< "$next_task"
  task_name=$(printf '%s\n' "$task_header" | sed -E 's/^### TODO[[:space:]]+//')
  task_block=$(extract_task_block "$start_line" "$end_line")

  log "Plan task found (eligible): $task_header"

  if ! update_task_status "$task_id" "TAKEN"; then
    log "Failed to mark task status as TAKEN: $task_id"
    return 1
  fi

  prompt=$(cat <<PROMPT
Read and execute exactly one task from $PLAN.

Task metadata:
- Task ID: $task_id
- Block lines: $start_line-$end_line
- Eligibility selected by runner: Backward compatibility=Yes and Status=TODO.

Execution requirements:
1. Implement ONLY this task using WHAT, WHERE, WHY, and References.
2. Keep all changes backward compatible.
3. When fully done, update ONLY this task status line in $PLAN:
   - Status: TAKEN.
   to
   - Status: DONE.
4. Do not modify status of any other task.
5. Do not commit and do not git add.
6. Run relevant tests/lint for touched areas and summarize validation.

Task block:
$task_block
PROMPT
)

  set +e
  run_codex "$prompt" "plan: $task_name"
  rc=$?
  set -e

  if [ $rc -ne 0 ]; then
    log "Task failed in Codex, reverting status TAKEN -> TODO for $task_id"
    if ! update_task_status "$task_id" "TODO"; then
      log "Failed to revert task status for $task_id"
    fi

    return "$rc"
  fi

  status_after=$(get_task_status "$task_id")
  if [ "$status_after" != "DONE" ]; then
    log "Task status after Codex is '$status_after'; runner will enforce DONE for $task_id"
    if ! update_task_status "$task_id" "DONE"; then
      log "Task completed but failed to mark status DONE: $task_id"
      return 1
    fi
  fi

  PLAN_TASKS_DONE=$(( PLAN_TASKS_DONE + 1 ))
  log "Task marked DONE: $task_id"
}

# --- Phase 3: Verify tests + lint ---
do_verify() {
  log "All tasks done — running tests..."
  if ! make test 2>&1 | tee -a "$LOG"; then
    log "Tests FAILED — asking Codex to fix"
    if ! run_codex "Run 'make test'. Some tests are failing. Read the test output, diagnose the failures, and fix them. Do NOT commit or git add any changes — leave everything unstaged for manual review. Follow all architecture rules from AGENTS.md and any repository Codex instructions. Output what you fixed." \
      "fix: test failures"; then
      log "Codex failed while fixing test failures"
      return 2
    fi
    return 1
  fi
  log "Tests passed"

  log "Running linter..."
  if ! make lint 2>&1 | tee -a "$LOG"; then
    log "Lint FAILED — asking Codex to fix"
    if ! run_codex "Run 'make lint'. There are linter errors. Read the lint output and fix all issues. Do NOT commit or git add any changes — leave everything unstaged for manual review. Follow all architecture rules from AGENTS.md and any repository Codex instructions. Output what you fixed." \
      "fix: lint errors"; then
      log "Codex failed while fixing lint errors"
      return 2
    fi
    return 1
  fi
  log "Lint passed"

  return 0
}

# === Main ===
TMPJSON=$(mktemp /tmp/agent-json-XXXXXX)
TMPRAW=$(mktemp /tmp/agent-raw-XXXXXX)
TMPERR=$(mktemp /tmp/agent-stderr-XXXXXX)

init_metrics
detect_runner

total_todo=$(count_todo_tasks)
eligible_todo=$(count_eligible_todo_tasks)
non_eligible_todo=$(( total_todo - eligible_todo ))

log "Session started — total TODO blocks: $total_todo, eligible TODO blocks: $eligible_todo, skipped blocks: $non_eligible_todo"

while true; do
  if [ -n "$(find_next_eligible_task)" ]; then
    log "Phase 1: plan.md eligible TODO blocks"
    if ! do_plan_task; then
      log "Plan task execution failed. Stopping."
      exit 1
    fi

    sleep 1
    continue
  fi

  log "No eligible TODO blocks left in $PLAN"
  break
done

if [ "$PLAN_TASKS_DONE" -gt 0 ]; then
  log "Phase 2: Verification"
  while true; do
    verify_rc=0
    if do_verify; then
      log "ALL DONE — eligible tasks complete, tests pass, lint clean. Stopping."
      exit 0
    fi

    verify_rc=$?
    if [ $verify_rc -gt 1 ]; then
      log "Verification failed due to Codex runner error. Stopping."
      exit "$verify_rc"
    fi

    log "Verification had failures — fixed, re-checking everything"
    sleep 1
  done
fi

log "No tasks were completed in this run (no eligible TODO blocks were found)."
