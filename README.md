# Ralph

Ralph is an autonomous software development agent that transforms natural language requirements into working code through iterative user story implementation.

Named after the "Ralph Wiggum pattern" - autonomous agents that learn and improve through clean iteration boundaries.

## Background

- **Original Pattern**: [Geoffrey Huntley](https://ghuntley.com/)
- **Ralph Philosophy**: [everything is a ralph loop](https://ghuntley.com/loop/)
- **History**: [A brief history of ralph](https://www.humanlayer.dev/blog/brief-history-of-ralph)

## How It Works

Ralph follows a two-phase approach:

1. **PRD Generation** - Analyzes your prompt and existing codebase to generate a structured Product Requirements Document with prioritized user stories
2. **Autonomous Implementation** - Iteratively implements each story, runs tests, and commits changes

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
│  • Create prd.json                                          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  Phase 2: Implementation Loop                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  For each story (by priority):                         │ │
│  │    1. Read existing code                               │ │
│  │    2. Implement solution                               │ │
│  │    3. Run tests                                        │ │
│  │    4. Commit changes                                   │ │
│  │    5. Mark story complete                              │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  Output: Working code with git history                      │
└─────────────────────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Ruby >= 3.0
- Git

### Setup

```bash
git clone https://github.com/your-org/ralph.git
cd ralph
bundle install
```

## Usage

### Full Implementation

Run Ralph with a natural language description of what you want to build:

```bash
./bin/ralph "Add user authentication with login and registration"
```

Ralph will:
1. Analyze your codebase
2. Generate a PRD with user stories
3. Create a feature branch
4. Implement each story iteratively
5. Run tests and commit changes
6. Clean up `prd.json` on completion

### Dry Run (PRD Only)

Generate a PRD without implementing to review the plan first:

```bash
./bin/ralph "Add user authentication" --dry-run
```

This creates `prd.json` for review. You can then run `--resume` to implement.

### Resume Implementation

Continue from an existing `prd.json` (useful after interruption or dry run):

```bash
./bin/ralph --resume
```

### CLI Options

| Option | Description |
|--------|-------------|
| `--dry-run` | Generate PRD only, skip implementation |
| `--resume` | Resume from existing `prd.json` |
| `--help`, `-h` | Show help message |

## Configuration

Ralph can be configured via `ralph.config.json` in your project root:

```json
{
  "model": "opencode/grok-code",
  "max_iterations": 50,
  "retry_attempts": 3,
  "retry_delay": 5,
  "log_level": "info"
}
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `model` | `opencode/grok-code` | AI model to use |
| `max_iterations` | `50` | Maximum implementation iterations |
| `retry_attempts` | `3` | Retries per story before giving up |
| `retry_delay` | `5` | Seconds between retries |
| `log_level` | `info` | Logging verbosity (`debug`, `info`, `warn`, `error`) |
| `prd_file` | `prd.json` | PRD filename |

### Supported Models

- `opencode/big-pickle`
- `opencode/glm-4.7-free`
- `opencode/gpt-5-nano`
- `opencode/grok-code` (default)
- `opencode/minimax-m2.1-free`

## PRD Format

The generated `prd.json` follows this structure:

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
      "priority": 1,
      "passes": false
    }
  ]
}
```

### Story Fields

| Field | Description |
|-------|-------------|
| `id` | Unique story identifier |
| `title` | Short story title |
| `description` | Detailed implementation description |
| `acceptance_criteria` | List of criteria that must be met |
| `priority` | Implementation order (1 = highest) |
| `passes` | Completion status (set by Ralph) |
| `retry_count` | Number of implementation attempts |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success - all stories completed |
| `1` | Failure - fatal error occurred |
| `2` | Partial - some stories completed, others failed |

## Development

### Running Tests

```bash
bundle exec rspec
```

### Linting

```bash
bundle exec rubocop
```

## Troubleshooting

### Interrupted Run

If Ralph is interrupted (Ctrl+C), your progress is saved in `prd.json`. Resume with:

```bash
./bin/ralph --resume
```

### Stories Failing Repeatedly

If stories exceed the retry limit, Ralph stops and reports which stories failed. You can:

1. Review the failing stories in `prd.json`
2. Make manual fixes
3. Run `./bin/ralph --resume` to continue

### Debug Mode

Set the `DEBUG` environment variable for verbose error output:

```bash
DEBUG=1 ./bin/ralph "your prompt"
```

## License

MIT
