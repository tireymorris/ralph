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
# or
go install .
```

## Usage

### Modes

```bash
ralph "build a todo app"               # TUI flow
ralph "build a todo app" --dry-run     # Generate PRD only
ralph --resume                          # Resume from existing prd.json
ralph prd "build a todo app"           # Start in PRD mode
ralph review                           # Start in review mode
ralph implement                        # Start in implementation mode
ralph status                            # Show current PRD status
```

```bash
ralph status
```

## Environment

Use `RALPH_MODEL` to select the backend and model in one value.

Use the backend's own help or model-list command as the source of truth for supported models.

Set `RALPH_MODEL` to a backend-specific string with the right format:

- `pi/<model>` or `pi/<provider>/<model>`
  - `pi --list-models`
  - `pi --help`
- `opencode/<model>`
  - `opencode models`
  - `opencode --help`
- `claude-code/<model>`
  - `claude --help`
- `cursor-agent/<model>`
  - `agent --help`

Examples:

```text
pi/github-copilot/claude-sonnet-4.6
opencode/gpt-5.5
claude-code/sonnet
cursor-agent/sonnet-4
```

## Development

```bash
go test ./...
go build -o ralph .
```