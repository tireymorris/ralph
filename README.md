# Ralph

Ralph turns a natural-language goal into a PRD, then implements the work story by story using an AI coding backend.

Supported backends:
- [pi](https://pi.dev)
- [OpenCode](https://github.com/opencode-ai/opencode)
- [Claude Code](https://github.com/anthropics/claude-code)
- Cursor Agent (`agent`)

## Flow

1. optionally ask clarifying questions
2. generate a PRD
3. review the PRD
4. implement stories in priority/dependency order
5. run tests and retry failed stories

## Requirements

- Go 1.24+
- Git
- One of the supported CLIs on `PATH`:
  - `pi`
  - `opencode`
  - `claude`
  - `agent` (Cursor Agent)

## Install

```bash
go build -o ralph .
go install .
```

## Usage

Run with no arguments to open the TUI prompt screen (requires a terminal). Provide a prompt on the command line to start immediately.

```bash
ralph                                            # TUI prompt screen
ralph "build a todo app"                         # TUI mode
ralph "build a todo app" --dry-run               # generate PRD only
ralph --dry-run                                  # prompt in TUI, then generate PRD only
ralph --resume                                   # resume from existing prd.json
ralph status                                     # show current PRD status
```

PRD output is written to `prd.json` in the working directory.

### Options

```text
--dry-run      Generate PRD only, don't implement
--resume       Resume implementation from existing prd.json
--verbose, -v  Enable debug logging
--help, -h     Show help
```

## Environment

Use `RALPH_RUNNER` to select the AI runner binary. Defaults to `claude`. Ralph does not pass a model to the runner; configure model selection in the runner itself.

Supported values:

```text
pi
cursor
claude
opencode
```

## Development

```bash
go test ./...
go build -o ralph .
```