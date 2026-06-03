# Ralph

Turn a natural-language goal into a `prd.json`, then implement it story-by-story via an AI coding CLI (clarify → PRD → review → implement).

## Quick start

```bash
git clone https://github.com/tireymorris/ralph .tmp-ralph && cd .tmp-ralph && go install . && cd .. && rm -rf .tmp-ralph
```

From a clone: `go install .` or `go build -o ralph .`

**Requires:** Go 1.24.0+, Git, and one runner on `PATH`: `claude` (default), `opencode`, `pi`, or `cursor-agent` (Cursor).

## Usage

```bash
ralph                               # TUI (needs a terminal)
ralph "build a todo app"
ralph "build a todo app" --dry-run  # PRD only
ralph --resume                      # continue from prd.json
ralph status                        # non-interactive progress
ralph web                           # local web UI (default http://127.0.0.1:8080)
ralph web --port 3000               # web UI on another port
```

| Flag | Purpose |
|------|---------|
| `--dry-run` | Generate PRD only |
| `--resume` | Resume from existing `prd.json` |
| `--port PORT` | Web server port (with `ralph web`; default 8080) |
| `-v`, `--verbose` | Debug logging |
| `-h`, `--help` | Help |

## State files

Ralph writes the following files in the working directory. All are covered by the repo `.gitignore`.

| Path | Created by | Purpose |
|------|-----------|---------|
| `prd.json` | TUI + web | The generated PRD |
| `prd.json.lock` | TUI + web | File lock for concurrent PRD access |
| `.ralph_questions.json` | Runner | Temporary clarification questions (deleted after read) |
| `.ralph/runs/<id>/meta.json` | Web UI | Per-run metadata (prompt, status, timestamps) |
| `.ralph/runs/<id>/events.ndjson` | Web UI | Per-run event log for SSE replay |
| `ralph.log` | All modes | Application log |

## Runner

Set `RALPH_RUNNER` to `claude`, `opencode`, `pi`, or `cursor` (Cursor Agent). Ralph does not pick a model itself—that stays in your runner's config.

Backends: [Claude Code](https://github.com/anthropics/claude-code), [OpenCode](https://github.com/opencode-ai/opencode), [pi](https://pi.dev), Cursor Agent.

## Development

```bash
go test ./...                 # Go unit + integration tests
cd web && npm test            # React/Vitest frontend tests
cd e2e && npx playwright test # Playwright E2E tests (builds Go + frontend first)
```

When you change the web UI (`web/`), rebuild the embedded assets:

```bash
go generate ./internal/web/...
```

CI runs all three test suites on push and PR via GitHub Actions (`.github/workflows/ci.yml`).
