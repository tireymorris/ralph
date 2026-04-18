# Ralph

Turns a natural-language goal into a PRD and implements user stories iteratively (OpenCode or Claude Code behind the scenes).

**Flow:** optional clarifying questions → PRD → review → implement stories (priority + dependencies, tests + commits). Failed stories roll back with `git reset --hard`. When everything passes, the PRD file (default `prd.json`) stays in place with all stories marked done; nothing renames or archives it.

## Requirements

- Go 1.24+
- Git
- [OpenCode](https://github.com/opencode-ai/opencode) **or** [Claude Code](https://github.com/anthropics/claude-code) CLI on `PATH`

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
| `RALPH_MODEL`          | `opencode/kimi-k2.5-free` | Model id (must be one of the values below) |
| `RALPH_MAX_ITERATIONS` | `50`                      | Cap on implementation iterations           |
| `RALPH_RETRY_ATTEMPTS` | `3`                       | Retries per story before giving up         |
| `RALPH_PRD_FILE`       | `prd.json`                | PRD filename in the working directory      |
| `RALPH_TEST_COMMAND`   | `go test ./...`           | Command run to verify each story           |


On success the PRD is only updated in place (no rename). This repository gitignores `prd.json` by default so it stays out of `git` unless you change `.gitignore` or `RALPH_PRD_FILE`.

**Supported `RALPH_MODEL` values** (from `internal/config/config.go`):

OpenCode: `opencode/kimi-k2.5-free`, `opencode/big-pickle`, `opencode/glm-4.7-free`, `opencode/gpt-5-nano`, `opencode/minimax-m2.1-free`, `opencode/trinity-large-preview-free`.

Claude Code: `claude-code/sonnet`, `claude-code/haiku`, `claude-code/opus`.

Upstream CLI docs list how each tool maps these to real models.

## Development

```bash
go test ./...
go build -o ralph .
```

