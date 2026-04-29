# Ralph

Turns a natural-language goal into a PRD and implements user stories iteratively ([pi](https://pi.dev), OpenCode, or Claude Code behind the scenes).

**Flow:** optional clarifying questions → PRD → review → implement stories (priority + dependencies, tests + commits). Failed stories roll back with `git reset --hard`. When everything passes, the PRD file (default `prd.json`) stays in place with all stories marked done; nothing renames or archives it.

## Requirements

- Go 1.24+
- Git
- [pi](https://www.npmjs.com/package/@mariozechner/pi-coding-agent) (`npm install -g @mariozechner/pi-coding-agent`), [OpenCode](https://github.com/opencode-ai/opencode), or [Claude Code](https://github.com/anthropics/claude-code) on `PATH` (default uses `pi`)

## Install

From the repository root:

```bash
go build -o ralph .   # or: go install .
```

## Usage


| Command                                          |                                      |
| ------------------------------------------------ | ------------------------------------ |
| `ralph "…"`                                      | TUI: full run                        |
| `ralph "…" --dry-run`                            | TUI: PRD only                        |
| `ralph --resume`                                 | TUI: continue from existing PRD file |
| `ralph status`                                   | Print PRD progress                   |
| `ralph run "…"`                                  | Headless full run                    |
| `ralph run "…" --dry-run` / `ralph run --resume` | Headless variants                    |


**Flags:** `--dry-run`, `--resume`, `-v` / `--verbose`, `-h` / `--help`

Typical review path: `ralph "…" --dry-run` → edit PRD → `ralph --resume`.

## Configuration

Environment variables (optional overrides; defaults shown):


| Variable               | Default                   | Role                                       |
| ---------------------- | ------------------------- | ------------------------------------------ |
| `RALPH_MODEL`          | `claude-code/sonnet`      | Model id (must use one of the prefixes below) |
| `RALPH_MAX_ITERATIONS` | `50`                      | Cap on implementation iterations           |
| `RALPH_RETRY_ATTEMPTS` | `3`                       | Retries per story before giving up         |
| `RALPH_PRD_FILE`       | `prd.json`                | PRD filename in the working directory      |
| `RALPH_TEST_COMMAND`   | `go test ./...`           | Command run to verify each story           |


On success the PRD is only updated in place (no rename). This repository gitignores `prd.json` by default so it stays out of `git` unless you change `.gitignore` or `RALPH_PRD_FILE`.

**Supported `RALPH_MODEL` prefixes** (see `internal/config/config.go`):

pi: `pi/<model>` runs `pi` with `--provider cursor` and `--model <model>` (for example `pi/auto`). Use `pi/<provider>/<model>` to set both flags (for example `pi/openai/gpt-4o`).

OpenCode: `opencode/...`, `opencode-go/...`, `anthropic/...`, `ollama/...` (examples: `opencode/kimi-k2.5-free`, `opencode/big-pickle`).

Claude Code: `claude-code/sonnet`, `claude-code/haiku`, `claude-code/opus`.

Each CLI’s docs describe how patterns map to real models.

## Development

```bash
go test ./...
go build -o ralph .
```

