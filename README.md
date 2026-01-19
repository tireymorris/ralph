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

- Go 1.21 or later
- Git
- [opencode](https://github.com/opencode-ai/opencode) - AI coding assistant CLI

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
```

### Headless Mode

Use the `run` command for non-interactive execution (ideal for CI pipelines):

```bash
# Full implementation with stdout output
ralph run "Add user authentication"

# Generate PRD only
ralph run "Add user authentication" --dry-run

# Resume implementation
ralph run --resume
```

### CLI Reference

```
Usage:
  ralph "your feature description"               # Interactive TUI
  ralph "your feature description" --dry-run     # Generate PRD only (TUI)
  ralph --resume                                 # Resume from prd.json (TUI)
  ralph run "your feature description"           # Headless mode
  ralph run "your feature description" --dry-run # Headless, PRD only
  ralph run --resume                             # Headless, resume

Options:
  --dry-run    Generate PRD only, skip implementation
  --resume     Resume implementation from existing prd.json
  --help, -h   Show help message

TUI Controls:
  q, Ctrl+C    Quit the application
```

## Configuration

Create a `ralph.config.json` in your project root to customize behavior:

```json
{
  "model": "opencode/grok-code",
  "max_iterations": 50,
  "retry_attempts": 3,
  "retry_delay": 5,
  "log_level": "info",
  "prd_file": "prd.json"
}
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `model` | `opencode/grok-code` | AI model to use for code generation |
| `max_iterations` | `50` | Maximum total implementation iterations |
| `retry_attempts` | `3` | Max retries per story before marking as failed |
| `retry_delay` | `5` | Seconds to wait between retries |
| `log_level` | `info` | Logging verbosity (`debug`, `info`, `warn`, `error`) |
| `prd_file` | `prd.json` | Filename for the PRD |

### Supported Models

- `opencode/big-pickle`
- `opencode/glm-4.7-free`
- `opencode/gpt-5-nano`
- `opencode/grok-code` (default)
- `opencode/minimax-m2.1-free`

## PRD Format

Ralph generates a `prd.json` file with this structure:

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
        "Password is securely hashed",
        "Model validations are in place"
      ],
      "test_spec": "Create integration test that: 1) Creates a user, 2) Verifies password hashing, 3) Validates required fields",
      "priority": 1,
      "passes": false,
      "retry_count": 0
    }
  ]
}
```

### Story Fields

| Field | Description |
|-------|-------------|
| `id` | Unique story identifier |
| `title` | Short descriptive title |
| `description` | Detailed implementation requirements |
| `acceptance_criteria` | List of conditions that must be met |
| `test_spec` | Guidance for writing integration tests |
| `priority` | Implementation order (1 = highest priority) |
| `passes` | Whether the story is complete (set by Ralph) |
| `retry_count` | Number of implementation attempts |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success - all stories completed |
| `1` | Failure - fatal error or all stories failed |
| `2` | Partial - some stories completed, others failed |

## Project Structure

```
ralph/
├── main.go                 # Entry point
├── internal/
│   ├── args/               # CLI argument parsing
│   ├── cli/                # Headless runner
│   ├── config/             # Configuration loading
│   ├── git/                # Git operations
│   ├── prd/                # PRD generation, storage, types
│   ├── prompt/             # AI prompt templates
│   ├── runner/             # OpenCode process runner
│   ├── story/              # Story implementation
│   ├── tui/                # Terminal UI (bubbletea)
│   └── workflow/           # Orchestration logic
├── ralph.config.json       # Default configuration
└── prd.json.example        # Example PRD
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests for a specific package
go test ./internal/prd -v
```

### Building

```bash
# Build for current platform
go build -o ralph .

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o ralph-linux .
GOOS=darwin GOARCH=arm64 go build -o ralph-macos .
```

## Troubleshooting

### Interrupted Run

If Ralph is interrupted (Ctrl+C), progress is saved in `prd.json`. Resume with:

```bash
ralph --resume
```

### Stories Failing Repeatedly

When stories exceed the retry limit, Ralph stops and reports failures. To resolve:

1. Check `prd.json` to see which stories failed
2. Review the error messages in the output
3. Make manual fixes to address the issue
4. Run `ralph --resume` to continue

### No prd.json Found

If you see "No prd.json found to resume from":

```bash
# Generate a new PRD first
ralph "your feature description" --dry-run

# Then resume
ralph --resume
```

### OpenCode Not Found

Ensure [opencode](https://github.com/opencode-ai/opencode) is installed and in your PATH:

```bash
which opencode
```

## License

MIT
