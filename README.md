# Ralph

Autonomous software development agent that transforms natural language requirements into working code through iterative user story implementation.

## Quick Start

```bash
# Install
go install .

# Interactive TUI mode
ralph "Add user authentication with login and registration"

# Headless mode for CI/scripts
ralph run "Add user authentication" --dry-run
```

## How It Works

Ralph follows a two-phase approach:

1. **PRD Generation** - Analyzes your prompt and codebase to generate a high-quality PRD with specific, measurable user stories and acceptance criteria. The prompt includes detailed instructions to avoid vague requirements and ensure implementation-ready specifications.
2. **Implementation Loop** - Iteratively implements each story by priority, writes tests, runs tests, and commits changes

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

# Common patterns
ralph "Create REST API endpoints for user management"
ralph "Add a modal dialog for user confirmation"
ralph "Add PostgreSQL database support with migrations"
ralph "Add comprehensive test suite with mocking"
```

## Configuration

Ralph is configured via environment variables:

```bash
export RALPH_MODEL="opencode/kimi-k2.5-free"
export RALPH_MAX_ITERATIONS=50
export RALPH_RETRY_ATTEMPTS=3
export RALPH_PRD_FILE="prd.json"
```

For Claude Code models:
```bash
export RALPH_MODEL="claude-code/sonnet"
```

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `RALPH_MODEL` | `opencode/kimi-k2.5-free` | AI model for code generation (OpenCode or Claude Code) |
| `RALPH_MAX_ITERATIONS` | `50` | Maximum total implementation iterations |
| `RALPH_RETRY_ATTEMPTS` | `3` | Max retries per story before failing |
| `RALPH_PRD_FILE` | `prd.json` | PRD filename |

### Supported Models

#### OpenCode Models
- `opencode/kimi-k2.5-free` (default)
- `opencode/big-pickle`
- `opencode/glm-4.7-free`
- `opencode/gpt-5-nano`
- `opencode/minimax-m2.1-free`
- `opencode/trinity-large-preview-free`

#### Claude Code Models
- `claude-code/sonnet`
- `claude-code/haiku`
- `claude-code/opus`

## PRD Format

Ralph generates `prd.json`:

```json
{
  "version": 1,
  "project_name": "User Authentication System",
  "branch_name": "feature/user-authentication",
  "context": "Go 1.24 with standard testing. Main code in cmd/ and internal/. Tests alongside code as _test.go files. Run with 'go test ./...'. Uses chi router for HTTP.",
  "test_spec": "1) Create user and verify password hashing, 2) Login with valid/invalid credentials, 3) Session persistence across requests",
  "stories": [
    {
      "id": "story-1",
      "title": "Create user model",
      "description": "Implement user model with email, password hash, and timestamps",
      "acceptance_criteria": [
        "User model exists with required fields",
        "Password is securely hashed"
      ],
      "priority": 1,
      "passes": false,
      "retry_count": 0
    }
  ]
}
```

| Field | Level | Description |
|-------|-------|-------------|
| `version` | PRD | Auto-incremented on each save for optimistic locking |
| `project_name` | PRD | Descriptive project name |
| `branch_name` | PRD | Git branch for the feature |
| `context` | PRD | Cached codebase context (language, structure, patterns) passed to each story |
| `test_spec` | PRD | Holistic test scenarios covering the entire feature (string, not array) |
| `id` | Story | Unique story identifier |
| `title` | Story | Short descriptive title |
| `description` | Story | Detailed implementation requirements |
| `acceptance_criteria` | Story | Conditions that must be met |
| `priority` | Story | Implementation order (1 = first) |
| `passes` | Story | Story completion status |
| `retry_count` | Story | Implementation attempts |

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
go build -o ralph .                                           # Current platform
GOOS=linux GOARCH=amd64 go build -o ralph-linux .           # Linux
GOOS=darwin GOARCH=arm64 go build -o ralph-macos .           # macOS
GOOS=windows GOARCH=amd64 go build -o ralph.exe .            # Windows
```

## Troubleshooting

### Interrupted Run

Progress saved to `prd.json`. Resume with:

```bash
ralph --resume
```

### Failed Stories

1. Run `ralph status` to see which stories failed
2. Review output for error details
3. Fix issues manually
4. Resume with `ralph --resume`

### Debug Mode

```bash
ralph run "your prompt" --verbose
```

### Missing AI CLI Tools

```bash
# For OpenCode models
which opencode
opencode --version

# For Claude Code models
which claude
claude --version
```

## License

MIT