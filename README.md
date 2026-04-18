# Ralph

Autonomous software development agent that transforms natural language requirements into working code through iterative user story implementation.

## How It Works

Ralph follows a two-phase approach:

1. **PRD Generation** - Optionally asks clarifying questions, then analyzes your prompt and codebase to generate a high-quality PRD with specific, measurable user stories and acceptance criteria.
2. **Implementation Loop** - Iteratively implements stories by priority (up to 2 in parallel, respecting declared dependencies), writes tests, runs tests, and commits changes. Failed stories are rolled back via `git reset --hard`; on success the PRD is archived to `prd-completed-<timestamp>.json`.

## Installation

### Prerequisites

- Go 1.24+
- Git
- [opencode](https://github.com/opencode-ai/opencode) CLI **OR** [Claude Code](https://github.com/anthropics/claude-code) CLI

### Install

```bash
# From source
git clone https://github.com/your-org/ralph.git
cd ralph
go build -o ralph .

# Global install
go install .
```

## Usage

### Commands

```bash
ralph "your feature description"               # Interactive TUI
ralph "your feature description" --dry-run     # Generate PRD only (TUI)
ralph --resume                                 # Resume from prd.json (TUI)
ralph status                                   # Show PRD progress
ralph run "your feature description"           # Headless mode
ralph run "your feature description" --dry-run # Headless, PRD only
ralph run --resume                             # Headless, resume
```

### Options

- `--dry-run` - Generate PRD only, skip implementation
- `--resume` - Resume implementation from existing prd.json
- `--verbose, -v` - Enable debug logging (stderr)
- `--help, -h` - Show help message

### Examples

```bash
# Basic feature implementation
ralph "Add user authentication with login and registration"

# Generate PRD first, review, then implement
ralph "Build a blog system" --dry-run
ralph status
ralph --resume

# CI/CD usage
ralph run "Add API rate limiting" --verbose

ralph "Create REST API endpoints for user management"
ralph "Add PostgreSQL database support with migrations"
```

## Configuration

Ralph is configured via environment variables:

```bash
export RALPH_MODEL="opencode/kimi-k2.5-free"
export RALPH_MAX_ITERATIONS=50
export RALPH_RETRY_ATTEMPTS=3
export RALPH_PRD_FILE="prd.json"
```

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `RALPH_MODEL` | `opencode/kimi-k2.5-free` | AI model for code generation (OpenCode or Claude Code) |
| `RALPH_MAX_ITERATIONS` | `50` | Maximum total implementation iterations |
| `RALPH_RETRY_ATTEMPTS` | `3` | Max retries per story before failing |
| `RALPH_PRD_FILE` | `prd.json` | PRD filename |

See [opencode](https://github.com/opencode-ai/opencode) and [Claude Code](https://github.com/anthropics/claude-code) docs for available models.

## Development

### Testing

```bash
go test ./...                 # All tests
go test ./... -cover          # With coverage
go test ./internal/prd -v     # Verbose package tests
go test ./... -race           # Race detector
```

### Building

```bash
go build -o ralph .
```

