# Ralph

Ralph is an autonomous software development agent that transforms natural language requirements into working code through iterative user story implementation.

Named after the "Ralph Wiggum pattern" - autonomous agents that learn and improve through clean iteration boundaries.

## Quick Start

```bash
# Install
go install .

# Run with a prompt (interactive TUI)
ralph "Add user authentication with login and registration"

# Or run headless for CI/scripts
ralph run "Add user authentication" --dry-run
```

## Background

- **Original Pattern**: [Geoffrey Huntley](https://ghuntley.com/)
- **Ralph Philosophy**: [Everything is a Ralph loop](https://ghuntley.com/loop/)
- **History**: [A brief history of Ralph](https://www.humanlayer.dev/blog/brief-history-of-ralph)

## How It Works

Ralph follows a two-phase approach:

1. **PRD Generation** - Analyzes your prompt and existing codebase to generate a structured Product Requirements Document with prioritized user stories
2. **Autonomous Implementation** - Iteratively implements each story, writes and runs tests, and commits changes

```
┌─────────────────────────────────────────────────────────────┐
│  Prompt: "Add user authentication"                          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  Phase 1: PRD Generation                                    │
│  • Scan codebase for patterns and conventions               │
│  • Generate user stories with acceptance criteria           │
│  • Include test specifications for each story               │
│  • Save to prd.json                                         │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  Phase 2: Implementation Loop                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  For each story (by priority):                         │ │
│  │    1. Read existing code                               │ │
│  │    2. Implement the feature                            │ │
│  │    3. Write integration tests                          │ │
│  │    4. Run tests until passing                          │ │
│  │    5. Commit changes                                   │ │
│  │    6. Mark story complete                              │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  Output: Working code with tests and git history            │
└─────────────────────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.21+
- Git
- [opencode](https://github.com/opencode-ai/opencode) CLI

### Build from Source

```bash
git clone https://github.com/your-org/ralph.git
cd ralph
go build -o ralph .
```

### Install Globally

```bash
go install .
```

## Usage

Ralph has two modes: **interactive TUI** (default) and **headless** (for CI/scripts).

### Interactive TUI Mode

```bash
# Full implementation with live progress display
ralph "Add user authentication with login and registration"

# Generate PRD only (review before implementing)
ralph "Add user authentication" --dry-run

# Resume from existing prd.json
ralph --resume

# Enable debug logging
ralph "Add feature" --verbose
```

### Headless Mode

Use the `run` command for non-interactive execution (ideal for CI pipelines):

```bash
# Full implementation with streaming stdout
ralph run "Add user authentication"

# Generate PRD only
ralph run "Add user authentication" --dry-run

# Resume with debug output
ralph run --resume --verbose
```

### CLI Reference

```
Usage:
   ralph "your feature description"               # Full implementation (TUI)
   ralph "your feature description" --dry-run     # Generate PRD only (TUI)
   ralph --resume                                 # Resume from existing prd.json (TUI)
   ralph run "your feature description"           # Full implementation (stdout)
   ralph run "your feature description" --dry-run # Generate PRD only (stdout)
   ralph run --resume                             # Headless, resume

Options:
   --dry-run      Generate PRD only, don't implement
   --resume       Resume implementation from existing prd.json
   --verbose, -v  Enable debug logging
   --help, -h     Show this help message

Modes:
   (default)    Interactive TUI with progress display
   run          Non-interactive stdout output (for CI/scripts)

Controls (TUI mode):
   q, Ctrl+C    Quit the application

Examples:
   ralph "Add user authentication with login and registration"
   ralph "Create a REST API for managing todos" --dry-run
   ralph --resume
   ralph run "Add unit tests for the API" --dry-run
   ralph run "Add feature" --verbose
```

### Usage Examples

#### Basic Feature Implementation
```bash
# Implement a complete feature with TUI progress display
ralph "Add a contact form with validation and email sending"

# Use headless mode for CI/CD pipelines
ralph run "Add API rate limiting" --verbose
```

#### Iterative Development
```bash
# Generate PRD first to review stories before implementation
ralph "Build a blog system with posts, comments, and admin panel" --dry-run

# Implement after reviewing the generated prd.json
ralph --resume
```

#### Configuration Examples
```bash
# Use a custom model for code generation
echo '{"model": "opencode/grok-code"}' > ralph.config.json
ralph "Add dark mode toggle to the UI"

# Increase retry attempts for complex features
echo '{"max_iterations": 100, "retry_attempts": 5}' > ralph.config.json
ralph "Implement complex data visualization charts"
```

#### Common Patterns
```bash
# API Development
ralph "Create REST API endpoints for user management"

# UI Components
ralph "Add a modal dialog for user confirmation"

# Database Integration
ralph "Add PostgreSQL database support with migrations"

# Testing
ralph "Add comprehensive test suite with mocking"

# Configuration
ralph "Add environment-based configuration loading"
```
Usage:
  ralph "your feature description"               # Interactive TUI
  ralph "your feature description" --dry-run     # Generate PRD only (TUI)
  ralph --resume                                 # Resume from prd.json (TUI)
  ralph run "your feature description"           # Headless mode
  ralph run "your feature description" --dry-run # Headless, PRD only
  ralph run --resume                             # Headless, resume

Options:
  --dry-run      Generate PRD only, skip implementation
  --resume       Resume implementation from existing prd.json
  --verbose, -v  Enable debug logging (stderr)
  --help, -h     Show help message

TUI Controls:
  q, Ctrl+C    Quit the application

Exit Codes:
  0    Success - all stories completed
  1    Failure - fatal error or all stories failed
  2    Partial - some stories completed, others failed
```

## Configuration

Create a `ralph.config.json` in your project root:

```json
{
  "model": "opencode/grok-code",
  "max_iterations": 50,
  "retry_attempts": 3,
  "retry_delay": 5,
  "prd_file": "prd.json"
}
```

| Option | Default | Description |
|--------|---------|-------------|
| `model` | `opencode/grok-code` | AI model for code generation |
| `max_iterations` | `50` | Maximum total implementation iterations |
| `retry_attempts` | `3` | Max retries per story before failing |
| `retry_delay` | `5` | Seconds between retries |
| `prd_file` | `prd.json` | PRD filename |

### Supported Models

- `opencode/big-pickle`
- `opencode/glm-4.7-free`
- `opencode/gpt-5-nano`
- `opencode/grok-code` (default)
- `opencode/minimax-m2.1-free`

## PRD Format

Ralph generates a `prd.json` file:

```json
{
  "project_name": "User Authentication System",
  "branch_name": "feature/user-authentication",
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
| `id` | Unique story identifier |
| `title` | Short descriptive title |
| `description` | Detailed implementation requirements |
| `acceptance_criteria` | Conditions that must be met |
| `test_spec` | Guidance for integration tests |
| `priority` | Implementation order (1 = highest) |
| `passes` | Story completion status |
| `retry_count` | Implementation attempts |

## Architecture

```
ralph/
├── main.go                 # Entry point, CLI initialization
├── internal/
│   ├── args/               # CLI argument parsing
│   ├── cli/                # Headless runner with event handling
│   ├── config/             # Configuration loading and defaults
│   ├── git/                # Git operations (branch, commit, status)
│   ├── logger/             # Structured logging (slog-based)
│   ├── prd/                # PRD types, generation, storage
│   │   ├── generator.go    # PRD generation with JSON parsing
│   │   ├── storage.go      # Load/save/delete operations
│   │   └── types.go        # PRD and Story structs
│   ├── prompt/             # AI prompt templates
│   ├── runner/             # OpenCode process execution
│   ├── story/              # Story implementation logic
│   ├── tui/                # Terminal UI (bubbletea)
│   │   ├── model.go        # TUI state management
│   │   ├── view.go         # Rendering logic
│   │   ├── commands.go     # Async operations
│   │   └── styles.go       # lipgloss styling
│   └── workflow/           # Orchestration and event system
├── ralph.config.json       # Default configuration
└── prd.json.example        # Example PRD
```

### Key Design Decisions

- **Dependency Injection**: All major components accept interfaces for testability (`CodeRunner`, `PRDGenerator`, `StoryImplementer`, `GitManager`)
- **Event-Driven**: Workflow emits events consumed by both TUI and CLI runners
- **Streaming Output**: Real-time output from opencode via channels
- **Structured Logging**: Debug logging via `log/slog` with `--verbose` flag
- **Graceful Interruption**: State saved to `prd.json` on interrupt for resume

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# With coverage
go test ./... -cover

# Verbose output for specific package
go test ./internal/prd -v

# Run with race detector
go test ./... -race
```

### Building

```bash
# Current platform
go build -o ralph .

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o ralph-linux .
GOOS=darwin GOARCH=arm64 go build -o ralph-macos .
GOOS=windows GOARCH=amd64 go build -o ralph.exe .
```

### Debug Mode

```bash
# See all internal logging
ralph run "test prompt" --verbose 2>&1 | less

# Structured log output example:
# time=2024-01-15T10:30:00 level=DEBUG msg="invoking opencode" model=opencode/grok-code prompt_length=1842
# time=2024-01-15T10:30:01 level=DEBUG msg="PRD generated" project="Test Project" stories=3
```

### Adding New Features

1. **New CLI flags**: Update `internal/args/args.go` and add to `HelpText()`
2. **New workflow events**: Add event type to `internal/workflow/workflow.go`, handle in both `cli/cli.go` and `tui/model.go`
3. **New configuration**: Add field to `internal/config/config.go` with default value

## Troubleshooting

### Interrupted Run

Progress is saved to `prd.json`. Resume with:

```bash
ralph --resume
```

### Stories Failing Repeatedly

1. Check `prd.json` for failed stories
2. Review output for error details
3. Fix issues manually
4. Resume: `ralph --resume`

### Debug Output

Enable verbose logging to see internal state:

```bash
ralph run "your prompt" --verbose
```

### OpenCode Not Found

Ensure opencode is installed and in PATH:

```bash
which opencode
opencode --version
```

### PRD Parse Errors

The JSON parser handles:
- Braces inside quoted strings
- Escaped characters
- Surrounding text (extracts JSON from response)

If parsing fails, check opencode's raw output with `--verbose`.

## License

MIT
