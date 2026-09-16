---
name: ralph
description: >-
  Ralph CLI (tireymorris/ralph): quick-start guidance for using Ralph in a repo.
  Use when the user mentions ralph, prd.json, RALPH_RUNNER, or wants to run or
  debug Ralph.
---

# Ralph

Ralph turns a natural-language goal into `prd.json`, then implements it story by story via an AI runner.

## Install

```bash
go install github.com/tireymorris/ralph@latest
```

## Use

Run from the target repo root:

```bash
ralph                 # TUI
ralph "..."          # TUI flow
ralph "..." --dry-run # PRD only
ralph --resume        # continue from existing prd.json
ralph status          # current PRD status
ralph web             # local UI
```

Set `RALPH_RUNNER` to `claude`, `cursor`, `opencode`, `pi`, or `copilot`.

## How it works

1. Clarify — runner may write `.ralph/questions.json`; Ralph reads and removes it
2. Generate/load PRD — runner writes `prd.json`
3. Review PRD — via TUI or `ralph web`
4. Implement — Ralph spawns one runner session per pending slice and marks `passes: true` when the runner exits 0
5. Cleanup — diff review + optional refactor rounds (`--skip-cleanup` to skip)

Ralph is an orchestrator; it does not write code.

## Testing

Each runner session owns testing:

- **Implementation:** red spec → green → related regressions → commit → update `slice.passes`
- **Recovery / cleanup:** run only tests needed to validate fixes or refactors

Planning should record the repo's test commands in PRD `context` for implementers.

## PRD format (`prd.json`)

The runner writes a single JSON object. Top-level fields:

| Field | Type | Notes |
|-------|------|-------|
| `version` | int | schema version (currently `1`) |
| `project_name` | string | human title for the PRD |
| `branch_name` | string | branch Ralph implements on (e.g. `feature/...`); set this so work stacks on the right base |
| `context` | string | repo/stack/conventions the runner should assume (test commands, layout, code style) |
| `test_spec` | string | the overall acceptance scenarios in prose — the bar each story is judged against |
| `stories` | array | ordered units of work (see below) |

Each **story**: `{ id (e.g. "story-1"), title, description, slices[] }`.

Each **slice** (the unit Ralph runs one runner session per):

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | e.g. `slice-1` |
| `behavior` | string | the observable behavior this slice must deliver |
| `red_hint` | string | how to write the failing test first — often the exact assertion |
| `refactor_hint` | string | implementation guidance; may be `""` |
| `passes` | bool | Ralph sets `true` when the runner session exits 0 |

## Web API

`ralph web` serves a local REST/SSE API (default `http://127.0.0.1:8080`, override with `--port`). Prefer this for programmatic/agent use (no TTY).

| Endpoint | Purpose | Request body |
|----------|---------|--------------|
| `POST /api/runs` | Start run; returns `{"id":"..."}` | `{"prompt":"...","auto_approve":false}` (`auto_approve:true` = `--yolo`: skip clarify + PRD review) |
| `GET /api/runs`, `GET /api/runs/{id}`, `GET /api/runs/{id}/prd` | List / status / PRD | — |
| `GET /api/runs/{id}/events` | SSE replay + live stream (use `curl -N`) | — |
| `POST /api/runs/{id}/clarify` | Clarification answers (when `waiting_clarify`) | `{"answers":[{"question":"...","answer":"..."}]}` |
| `POST /api/runs/{id}/review` | Approve or revise PRD (when `waiting_review`) | `{"action":"approve"}` **or** `{"action":"revise","critique":"..."}` |
| `POST /api/runs/{id}/implementation-review` | Continue after review findings (when `waiting_implementation_review`) | `{}` |
| `POST /api/runs/{id}/followup` | Send a follow-up message | `{"message":"..."}` |
| `POST /api/runs/{id}/cancel`, `POST /api/runs/{id}/resume` | Control | — |
| `GET /health` | Liveness check (200 when up) | — |
| `GET /api/version`, `POST /api/update`, `POST /api/clean` | Meta | — |

The `review` body is strict: `action` must be exactly `approve` or `revise`, and `revise` **requires** a non-empty `critique` (not `feedback`); wrong or missing fields return `{"error":"..."}` naming the expected field. The SSE stream replays `.ralph/runs/{id}/events.ndjson` then streams live events, e.g. `EventOutput` (`{payload:{Text}}`), `EventClarifyingQuestions` (`{payload:{Questions}}`), `EventPRDReview`, `EventCompleted`.

Statuses: `running`, `waiting_clarify`, `waiting_review`, `waiting_implementation_review`, `implementing`, `completed`, `failed`, `cancelled`. TUI runs use id `prd-local`.

## Key files

Gitignore these in the target repo:

- `prd.json`
- `prd.json.lock`
- `.ralph/` (run state, backups, clarify questions, review verdicts)
- `.prd.tmp.*` (top-level atomic-save temp written next to `prd.json`)

## Caveats

- `passes: true` is not proof tests passed
- `ralph status` is progress, not QA sign-off
- large PRD runs can overscope badly
- `--dry-run` may still need a real TTY in some environments
- Ralph does not load `CLAUDE.md` unless the runner does
- **No TTY (headless/agent runs):** the TUI (`ralph`, `ralph --resume`, `--yolo`) needs `/dev/tty`. From an agent/non-terminal context use `ralph --resume --headless` (resumes an existing `prd.json` with no Bubble Tea) and run it in the background.
