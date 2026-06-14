# Ralph

Turn a goal into `prd.json`, then implement it story-by-story via an AI coding CLI.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/tireymorris/ralph/main/scripts/install.sh | bash
```

**Requires:** Go 1.24.0+, Git, and one runner on `PATH`: `claude` (default), `opencode`, `pi`, `cursor-agent` (Cursor), or `copilot`.

Upgrade: `ralph update` (see `ralph update --help`). From a clone: `go install .` or `scripts/build.sh -o ralph`.

## Usage

```bash
ralph "build a feature"          # TUI (needs a terminal)
ralph "build a feature" --dry-run
ralph --resume
ralph status
ralph clean
ralph web                        # local UI at http://127.0.0.1:8080
```

Implementation needs a **git repository** in the working directory.

`ralph clean` removes Ralph state in cwd: `prd.json`, its lock, `.ralph/` (including `.ralph/prd.tmp.*` orphans and run data). Safe to run when nothing exists.

| Flag / env | Purpose |
|------------|---------|
| `--dry-run` | PRD only |
| `--resume` | Continue from `prd.json` |
| `--skip-cleanup` | Skip post-implementation cleanup |
| `--yolo` / `RALPH_YOLO=1` | Skip manual clarify and PRD approval |
| `RALPH_RUNNER` | `claude`, `opencode`, `pi`, `cursor`, or `copilot` |
| `RALPH_RUNNER_TIMEOUT` | Per-session timeout, e.g. `30m` |
| `-v`, `--verbose` | Debug logging |

## Runners

| `RALPH_RUNNER` | Binary | Link |
|----------------|--------|------|
| `claude` (default) | `claude` | [Claude Code](https://github.com/anthropics/claude-code) |
| `opencode` | `opencode` | [OpenCode](https://github.com/opencode-ai/opencode) |
| `pi` | `pi` | [pi](https://pi.dev) |
| `cursor` | `cursor-agent` | [Cursor](https://cursor.com) |
| `copilot` | `copilot` | [Copilot CLI](https://docs.github.com/en/copilot/how-tos/copilot-cli) |

Ralph does not handle runner auth. For Copilot: `copilot login`, or set `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, or `GITHUB_TOKEN`.

## Web API

`ralph web` serves a local UI and REST/SSE API (default `http://127.0.0.1:8080`). Prefer this over the TUI for automation.

| Endpoint | Purpose |
|----------|---------|
| `POST /api/runs` | Start a run (`{ "prompt": "...", "auto_approve": true }` = `--yolo`) |
| `GET /api/runs` | List runs |
| `GET /api/runs/{id}` | Run status |
| `GET /api/runs/{id}/prd` | PRD JSON |
| `GET /api/runs/{id}/events` | SSE replay + live stream |
| `POST /api/runs/{id}/clarify` | Submit clarification answers |
| `POST /api/runs/{id}/review` | Approve or revise PRD |
| `POST /api/runs/{id}/implementation-review` | Continue after review findings |
| `POST /api/runs/{id}/cancel` | Cancel run |
| `POST /api/runs/{id}/resume` | Force resume from checkpoint |
| `POST /api/runs/{id}/followup` | Follow-up on a terminal run |
| `GET /api/version` | Version / update check |
| `POST /api/update` | Install update |
| `POST /api/clean` | Remove Ralph state in cwd |

Run statuses: `running`, `waiting_clarify`, `waiting_review`, `waiting_implementation_review`, `implementing`, `completed`, `failed`, `cancelled`. TUI PRD runs appear as `prd-local`.

## State files

Written in the working directory (gitignored). New runs archive prior state to `.ralph/backups/<timestamp>/`. `--resume` does not archive existing state.

| Path | Purpose |
|------|---------|
| `prd.json` / `prd.json.lock` | PRD and file lock |
| `.ralph/questions.json` | Clarification questions (temporary) |
| `.ralph/prd_review.json` | PRD self-review verdict in `--yolo` runs (temporary) |
| `.ralph/prd.tmp.*` | Atomic-save temp files |
| `.ralph/runs/<id>/` | Run metadata, events, review transcripts |
| `.ralph/backups/<timestamp>/` | Prior state from earlier runs |

`ralph clean` deletes these artifacts idempotently.

## Development

```bash
scripts/build.sh -o "$(go env GOPATH)/bin/ralph"
go test ./...
cd web && npm test
cd e2e && npx playwright test
go generate ./internal/web/...   # after web UI changes
```
