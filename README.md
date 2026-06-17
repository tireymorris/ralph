# Ralph

Turn a goal into `prd.json`, then implement it slice-by-slice via an AI coding CLI. Ralph orchestrates; the runner writes code.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/tireymorris/ralph/main/scripts/install.sh | bash
```

**Requires:** Go 1.24.0+, Git, and one runner on `PATH`.

Upgrade: `ralph update`. From a clone: `go install .` or `scripts/build.sh -o ralph`.

## Usage

```bash
ralph "build a feature"          # TUI (needs a terminal)
ralph "build a feature" --dry-run
ralph --resume
ralph status
ralph clean
ralph web                        # http://127.0.0.1:8080
```

Implementation requires a git repo in the working directory.

| Flag / env | Purpose |
|------------|---------|
| `--dry-run` | PRD only |
| `--resume` | Continue from `prd.json` (checkpoint-aware) |
| `--skip-cleanup` | Skip post-implementation cleanup |
| `--yolo` / `RALPH_YOLO=1` | Skip clarify and PRD approval |
| `--verbose` | Debug logging |
| `RALPH_RUNNER` | `claude`, `opencode`, `pi`, `cursor`, or `copilot` |
| `RALPH_RUNNER_TIMEOUT` | Per-session timeout, e.g. `30m` |
| `RALPH_BRANCH_PREFIX` | Branch prefix for PRD `branch_name` (default: `feature`) |
| `RALPH_DEFAULT_BRANCHES` | Comma-separated default branch names (default: detect from git, then `main`, `master`, `develop`, `trunk`) |
| `RALPH_TEST_COMMAND` | Override auto-detected project test command |

On startup, Ralph detects an existing codebase from project manifests (e.g. `go.mod`, `package.json`) or source files, and picks a test command when none is set (`go test ./...`, `npm test`, `cargo test`, etc.). PRD generation uses `RALPH_BRANCH_PREFIX` for suggested branch names. Implementation checks out the PRD branch only when the current branch is a configured default.

`ralph clean` removes `prd.json`, its lock, and `.ralph/` (including temp files and run data).

## Workflow

1. **Clarify** — runner may write `.ralph/questions.json`; Ralph reads and removes it
2. **Generate/load PRD** — runner writes `prd.json`
3. **PRD self-review** — `--yolo` runs only; failures keep the last revision
4. **Review PRD** — approve or revise (skipped with `--yolo` / `auto_approve`)
5. **Implement** — one runner session per pending slice; Ralph marks `slice.passes` and `story.passes` when the runner succeeds
6. **Implementation review** — critical diff review after each story; findings trigger an automatic recovery loop (re-review until clean or limits hit). Web `POST .../implementation-review` and `--resume` continue stalled review checkpoints.
7. **Cleanup** — optional final pass (skip with `--skip-cleanup`)

TUI and web share `workflow.Driver` → `Executor`. Web adds registry + SSE via `RunController`; TUI uses `FileReviewLoop` under `.ralph/runs/prd-local/`.

## Runners

| `RALPH_RUNNER` | Binary | Notes |
|----------------|--------|-------|
| `claude` (default) | `claude` | [Claude Code](https://github.com/anthropics/claude-code) |
| `opencode` | `opencode` | [OpenCode](https://github.com/opencode-ai/opencode) |
| `pi` | `pi` | [pi](https://pi.dev) |
| `cursor` | `cursor-agent` | [Cursor](https://cursor.com) |
| `copilot` | `copilot` | [Copilot CLI](https://docs.github.com/en/copilot/how-tos/copilot-cli); `copilot login` or token env vars |

Ralph does not handle runner auth.

## Web API

`ralph web` serves a local REST/SSE API (default `http://127.0.0.1:8080`). Prefer this for programmatic use (no TTY).

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
| `GET /api/version`, `POST /api/update`, `POST /api/clean` | Meta | — |

`ralph web --port 3000` overrides the default port. The `review` body is strict: `action` must be exactly `approve` or `revise`, and `revise` **requires** a non-empty `critique` (not `feedback`) — wrong/missing fields return `{"error":"..."}` with the expected name. The SSE stream replays `.ralph/runs/{id}/events.ndjson` then streams live events, e.g. `EventOutput` (`{payload:{Text}}`), `EventClarifyingQuestions` (`{payload:{Questions}}`), `EventPRDReview`, `EventCompleted`.

Statuses: `running`, `waiting_clarify`, `waiting_review`, `waiting_implementation_review`, `implementing`, `completed`, `failed`, `cancelled`. TUI runs use id `prd-local`.

## State files

Written in the working directory (gitignored). New runs archive prior state to `.ralph/backups/<timestamp>/`; `--resume` does not archive.

| Path | Purpose |
|------|---------|
| `prd.json` / `prd.json.lock` | PRD and file lock |
| `.ralph/questions.json` | Clarification questions (temporary) |
| `.ralph/prd_review.json` | PRD self-review verdict in `--yolo` runs (temporary) |
| `.ralph/prd.tmp.*` | Atomic-save temp files |
| `.ralph/runs/<id>/meta.json` | Status, checkpoint, review loop state |
| `.ralph/runs/<id>/events.ndjson` | Event log for SSE replay |
| `.ralph/runs/<id>/review-*.txt` | Implementation review transcripts |
| `.ralph/backups/<timestamp>/` | Archived prior state |

Checkpoints: `prd_review`, `impl_review`, `followup`, `complete`.

## Architecture

Go CLI/TUI with optional embedded web UI (`web/` → `internal/web/static/dist/`).

| Package | Role |
|---------|------|
| `internal/app` | Coordinator: CLI routing, config, validation |
| `internal/args` | Flag parsing |
| `internal/workflow` | State machine, phases, events, `Driver` |
| `internal/shared/session` | Facade over `Driver` for TUI/web |
| `internal/tui` | Bubble Tea UI |
| `internal/web` | HTTP server, handlers, `RunController`, registry |
| `internal/shared/prd` | PRD model, locking, storage |
| `internal/shared/runner` | Runner integrations + mock |
| `internal/shared/workdir` | Git branch helpers, codebase/test-command detection |
| `internal/shared/runpaths` | `.ralph/runs/<id>/` path helpers |
| `internal/prompt` | Embedded templates, kind markers, render helpers |

**Start reading:** `internal/app/coordinator.go`, `internal/workflow/phase_generate.go`, `internal/workflow/phase_implement.go`, `internal/workflow/phase_implement_review.go`.

Rendered prompts include a machine-readable kind marker (`===ralph-prompt-kind:…===`) so runners and tests can identify prompt type without parsing prose. See [`docs/prompts/`](docs/prompts/).

## Development

```bash
scripts/build.sh -o "$(go env GOPATH)/bin/ralph"
go test ./...
cd web && npm test
cd e2e && npx playwright test
go generate ./internal/web/...   # after web UI changes
```

Agent prompts live in `internal/prompt/templates/` (embedded at build time). See `docs/prompts/` for the template index and editing workflow.

## Caveats

- `passes: true` means the runner exited 0, not that tests passed
- `ralph status` is progress, not QA sign-off
- Large PRDs can overscope; keep stories small
- Implementation review needs git and a runner that emits `===ralph-findings===` JSON
