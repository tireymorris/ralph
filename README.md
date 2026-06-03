# Ralph

Turn a natural-language goal into a `prd.json`, then implement it story-by-story via an AI coding CLI (clarify → PRD → review → implement).

## Quick start

```bash
curl -fsSL https://raw.githubusercontent.com/tireymorris/ralph/main/scripts/install.sh | bash
```

Pin a branch or tag: `curl -fsSL .../install.sh | bash -s -- main`

Upgrade an existing install (same flow as `scripts/install.sh`):

```bash
ralph update
ralph update --check    # exit 0 if up to date, 2 if a newer commit is on the remote
ralph update --ref main # install a specific branch or tag
```

Override the git remote: `RALPH_REPO=https://github.com/you/ralph.git ralph update`

From a clone: `go install .` or `scripts/build.sh -o ralph` (embeds git version metadata)

Manual one-liner:

```bash
git clone https://github.com/tireymorris/ralph .tmp-ralph && cd .tmp-ralph && go install . && cd .. && rm -rf .tmp-ralph
```

**Requires:** Go 1.24.0+, Git, and one runner on `PATH`: `claude` (default), `opencode`, `pi`, or `cursor-agent` (Cursor).

## Usage

```bash
ralph                               # TUI (needs a terminal)
ralph "build a todo app"
ralph "build a todo app" --dry-run  # PRD only
ralph --resume                      # continue from prd.json
ralph status                        # non-interactive progress
ralph clean                         # remove all Ralph agent state in cwd
ralph version                       # build version, commit, and ref
ralph update                        # reinstall from GitHub (default branch main)
ralph update --check                # compare local commit to remote
ralph web                           # local web UI (default http://127.0.0.1:8080)
ralph web --port 3000               # web UI on another port
```

`ralph clean` removes all Ralph agent state in the current working directory: `prd.json`, its lock, `.ralph_questions.json`, orphaned `.prd.tmp.*` files, and `.ralph/` run data.

| Flag / env | Purpose |
|------------|---------|
| `--dry-run` | Generate PRD only |
| `--resume` | Resume from existing `prd.json` |
| `--port PORT` | Web server port (with `ralph web`; default 8080) |
| `--ref REF` | Branch or tag for `ralph update` (default `main`) |
| `--check` | With `ralph update`: report whether a newer commit exists on the remote |
| `RALPH_REPO` | Git URL for `ralph update` (default `https://github.com/tireymorris/ralph.git`) |
| `-v`, `--verbose` | Debug logging |
| `-h`, `--help` | Help |

## State files

Ralph writes the following files in the working directory. All are covered by the repo `.gitignore`. Run `ralph clean` to delete these artifacts idempotently (safe to run when none exist).

| Path | Created by | Purpose |
|------|-----------|---------|
| `prd.json` | TUI + web | The generated PRD |
| `prd.json.lock` | TUI + web | File lock for concurrent PRD access |
| `.ralph_questions.json` | Runner | Temporary clarification questions (deleted after read) |
| `.prd.tmp.*` | TUI + web | Atomic-save temp files (orphans removed by `ralph clean`) |
| `.ralph/runs/<id>/meta.json` | Web UI | Per-run metadata (prompt, status, timestamps) |
| `.ralph/runs/<id>/events.ndjson` | Web UI | Per-run event log for SSE replay |

## Runner

Set `RALPH_RUNNER` to `claude`, `opencode`, `pi`, or `cursor` (Cursor Agent). Ralph does not pick a model itself—that stays in your runner's config.

Backends: [Claude Code](https://github.com/anthropics/claude-code), [OpenCode](https://github.com/opencode-ai/opencode), [pi](https://pi.dev), Cursor Agent.

## Development

Release-style build with git metadata embedded in the binary:

```bash
scripts/build.sh -o "$(go env GOPATH)/bin/ralph"
```

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
