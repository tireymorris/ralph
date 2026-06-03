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

Writes `prd.json` in the working directory. The web UI also stores per-run metadata under `.ralph/runs/` in the working directory.

| Flag | Purpose |
|------|---------|
| `--dry-run` | Generate PRD only |
| `--resume` | Resume from existing `prd.json` |
| `--port PORT` | Web server port (with `ralph web`; default 8080) |
| `-v`, `--verbose` | Debug logging |
| `-h`, `--help` | Help |

## Runner

Set `RALPH_RUNNER` to `claude`, `opencode`, `pi`, or `cursor` (Cursor Agent). Ralph does not pick a model itself—that stays in your runner’s config.

Backends: [Claude Code](https://github.com/anthropics/claude-code), [OpenCode](https://github.com/opencode-ai/opencode), [pi](https://pi.dev), Cursor Agent.

## Development

```bash
go test ./...
```

When you change the web UI (`web/`), build the embedded assets before Go tests:

```bash
cd web && npm ci && npm run build
```
