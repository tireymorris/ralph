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

1. **PRD Generation** - Analyzes your prompt and codebase to generate structured user stories with acceptance criteria
2. **Implementation Loop** - Iteratively implements each story, writes tests, runs tests, and commits changes

## Installation

### Prerequisites

- Go 1.21+
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

# Generate PRD first, then implement
ralph "Build a blog system" --dry-run
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

Create `ralph.config.json` in your project root:

```json
{
  "model": "opencode/big-pickle",
  "max_iterations": 50,
  "retry_attempts": 3,
  "prd_file": "prd.json"
}
```

For Claude Code models:
```json
{
  "model": "claude-code/claude-3.5-sonnet",
  "max_iterations": 50,
  "retry_attempts": 3,
  "prd_file": "prd.json"
}
```

| Option | Default | Description |
|--------|---------|-------------|
| `model` | `opencode/big-pickle` | AI model for code generation (OpenCode or Claude Code) |
| `max_iterations` | `50` | Maximum total implementation iterations |
| `retry_attempts` | `3` | Max retries per story before failing |
| `prd_file` | `prd.json` | PRD filename |

### Supported Models

#### OpenCode Models
- `opencode/big-pickle` (default)
- `opencode/glm-4.7-free`
- `opencode/gpt-5-nano`
- `opencode/minimax-m2.1-free`

#### Claude Code Models
- `claude-code/claude-3.5-sonnet`
- `claude-code/claude-3.5-haiku`
- `claude-code/claude-3-opus`

## PRD Format

Ralph generates `prd.json`:

```json
{
  "project_name": "User Authentication System",
  "branch_name": "feature/user-authentication",
  "context": "Go 1.21 with standard testing. Main code in cmd/ and internal/. Tests alongside code as _test.go files. Run with 'go test ./...'. Uses chi router for HTTP.",
  "stories": [
    {
      "id": "story-1",
      "title": "Create user model",
      "description": "Implement user model with email, password hash, and timestamps",
      "acceptance_criteria": [
        "User model exists with required fields",
        "Password is securely hashed"
      ],
      "test_spec": "Integration test: 1) Create user, 2) Verify password hashing",
      "priority": 1,
      "passes": false,
      "retry_count": 0
    }
  ]
}
```

| Field | Description |
|-------|-------------|
| `context` | Cached codebase context (language, structure, patterns) passed to each story |
| `id` | Unique story identifier |
| `title` | Short descriptive title |
| `description` | Detailed implementation requirements |
| `acceptance_criteria` | Conditions that must be met |
| `test_spec` | Guidance for integration tests |
| `priority` | Implementation order (1 = highest) |
| `passes` | Story completion status |
| `retry_count` | Implementation attempts |

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

1. Check `prd.json` for failed stories
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
which claude-code
claude-code --version
```

## License

MIT