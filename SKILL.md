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
2. Generate/load PRD — runner writes `prd.json` (test commands go in `context`, not a separate orchestrator field)
3. Review PRD — via TUI or `ralph web`
4. Implement — Ralph spawns one runner session per pending slice; the runner runs targeted tests for that slice before updating `slice.passes`
5. Cleanup — diff review + optional refactor rounds (`--skip-cleanup` to skip)

Ralph is an orchestrator; it does not write code or run a project-wide test command.

## Testing

Ralph does **not** auto-detect or shell out test commands (no `RALPH_TEST_COMMAND`, no post-story test gate). Each runner session owns testing:

- **Implementation:** red spec → green → related regressions → commit → update `slice.passes`
- **Recovery / cleanup:** run only tests needed to validate fixes or refactors

Planning should record observed test commands in PRD `context` for implementers. `passes: true` means the runner exited 0 — not that Ralph verified a suite.

## PRD format (`prd.json`)

| Field | Type | Notes |
|-------|------|-------|
| `version` | int | schema version |
| `project_name` | string | human title |
| `branch_name` | string | branch Ralph implements on |
| `context` | string | stack, layout, conventions, **test commands** |
| `test_spec` | string | holistic acceptance scenarios in prose |
| `stories` | array | ordered units of work |

Each **slice**: `{ id, behavior, red_hint, refactor_hint?, passes }` — Ralph runs one session per pending slice.

## Web API

`ralph web` — default `http://127.0.0.1:8080`. See repo `README.md` for full endpoint table.

Statuses: `running`, `waiting_clarify`, `waiting_review`, `waiting_implementation_review`, `implementing`, `completed`, `failed`, `cancelled`.

## Key files

Gitignore in the target repo: `prd.json`, `prd.json.lock`, `.ralph/`, `.prd.tmp.*`

## Caveats

- `passes: true` is not proof tests passed — runner must run targeted tests during the session
- Ralph does not run a project-wide test gate
- `ralph status` is progress, not QA sign-off
- large PRD runs can overscope badly
- `--dry-run` may still need a real TTY in some environments
- Ralph does not load `CLAUDE.md` unless the runner does
- **Telegram + runners:** see repo `README.md` / `telegram` skill for bridge gotchas with `RALPH_RUNNER=claude` or `cursor`
- **Headless:** use `ralph --resume --headless` from non-TTY contexts
