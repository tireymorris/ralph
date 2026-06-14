---
name: ralph
description: >-
  Ralph CLI (tireymorris/ralph): quick-start guidance for using Ralph in a repo.
  Use when the user mentions ralph, prd.json, RALPH_RUNNER, or wants to run or
  debug Ralph.
---

# Ralph

Ralph turns a goal into `prd.json`, then implements it story-by-story via an AI runner. Ralph is an orchestrator; it does not write code.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/tireymorris/ralph/main/scripts/install.sh | bash
```

Run from the target repo root (git required for implementation). Upgrade: `ralph update`. From a clone: `go install .` or `scripts/build.sh -o ralph`.

## CLI

```bash
ralph "build a feature"          # TUI (needs a terminal)
ralph "build a feature" --dry-run
ralph --resume
ralph status
ralph clean
ralph web [--port N]             # local UI (default http://127.0.0.1:8080)
```

Flags: `--dry-run`, `--resume`, `--skip-cleanup`, `--yolo`, `--verbose`.

Env:
- `RALPH_RUNNER` — `claude` (default), `opencode`, `pi`, `cursor`, or `copilot`
- `RALPH_YOLO=1` — skip clarify and PRD approval gates
- `RALPH_RUNNER_TIMEOUT` — per-session timeout, e.g. `30m`
- `RALPH_REPO` — git URL for `ralph update`

## Runners

| `RALPH_RUNNER` | Binary on `PATH` |
|----------------|------------------|
| `claude` (default) | `claude` |
| `opencode` | `opencode` |
| `pi` | `pi` |
| `cursor` | `cursor-agent` |
| `copilot` | `copilot` |

Ralph does not handle runner auth. For Copilot: `copilot login`, or set `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, or `GITHUB_TOKEN`.

## Workflow

1. Clarify — runner may write `.ralph/questions.json`; Ralph reads and removes it
2. Generate/load PRD — runner writes `prd.json`
3. Review PRD — approve or revise (skipped with `--yolo` / auto-approve; agent self-reviews instead)
4. Implement — one runner session per ready story; `passes: true` when the runner exits 0
5. Implementation review — critical diff review after each story; may pause on findings
6. Cleanup — optional final diff pass (skip with `--skip-cleanup`)

## Web API

Prefer the web API when driving Ralph programmatically (no TTY needed).

**Runs**
- `POST /api/runs` — `{ "prompt": "...", "auto_approve": true }` (same as `--yolo`)
- `GET /api/runs` — list runs
- `GET /api/runs/{id}` — run status
- `GET /api/runs/{id}/prd` — PRD JSON
- `GET /api/runs/{id}/events` — SSE replay, then live events
- `POST /api/runs/{id}/clarify` — submit clarification answers
- `POST /api/runs/{id}/review` — `{ "action": "approve" }` or `{ "action": "revise", "critique": "..." }`
- `POST /api/runs/{id}/implementation-review` — continue after review findings
- `POST /api/runs/{id}/cancel` — cancel run
- `POST /api/runs/{id}/resume` — force resume from checkpoint
- `POST /api/runs/{id}/followup` — follow-up on a terminal run

**Other**
- `GET /api/version`
- `POST /api/update`
- `POST /api/clean`

**Run statuses:** `running`, `waiting_clarify`, `waiting_review`, `waiting_implementation_review`, `implementing`, `completed`, `failed`, `cancelled`.

Notes:
- TUI PRD-backed runs appear in the API as id `prd-local`.
- Implementation review requires git and a runner that emits `===ralph-findings===` JSON.

## Key files

Gitignore in the target repo:

- `prd.json` / `prd.json.lock`
- `.ralph/` — questions, self-review verdict, `prd.tmp.*`, `runs/<id>/` metadata and events, review transcripts, backups

`ralph clean` removes Ralph state in cwd. For a full repo cleanup with PRD backup, use the **cleanup-local-repo** skill.

## Caveats

- `passes: true` is not proof tests passed
- `ralph status` is progress, not QA sign-off
- large PRD runs can overscope badly
- `--dry-run` may still need a real TTY in some environments
