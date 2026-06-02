# Ralph

Turn a natural-language goal into a `prd.json`, then implement it story-by-story via an AI coding CLI (clarify → PRD → review → implement).

## Quick start

```bash
git clone https://github.com/tireymorris/ralph .tmp-ralph && cd .tmp-ralph && go install . && cd .. && rm -rf .tmp-ralph
```

From a clone: `go install .` or `go build -o ralph .`

**Requires:** Go 1.24+, Git, and one runner on `PATH`: `claude` (default), `opencode`, `pi`, or `agent` (Cursor).

## Usage

```bash
ralph                          # TUI (needs a terminal)
ralph "build a todo app"
ralph "build a todo app" --dry-run   # PRD only
ralph --resume                 # continue from prd.json
ralph status                   # non-interactive progress
```

Writes `prd.json` in the working directory.

| Flag | Purpose |
|------|---------|
| `--dry-run` | Generate PRD only |
| `--resume` | Resume from existing `prd.json` |
| `-v`, `--verbose` | Debug logging |
| `-h`, `--help` | Help |

## Runner

Set `RALPH_RUNNER` to `claude`, `opencode`, `pi`, or `cursor` (uses `agent`). Ralph does not pick a model—that stays in your runner’s config.

Backends: [Claude Code](https://github.com/anthropics/claude-code), [OpenCode](https://github.com/opencode-ai/opencode), [pi](https://pi.dev), Cursor Agent.

## Development

```bash
go test ./...
```
