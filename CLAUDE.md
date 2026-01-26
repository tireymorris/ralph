# Ralph

Autonomous software development agent that transforms natural language requirements into working code through iterative user story implementation.

## Quick Reference

```bash
# Build
go build -o ralph .

# Test
go test ./...
go test -race ./...    # with race detection

# Run
./ralph "your prompt here"           # TUI mode
./ralph run "your prompt here"       # headless mode
./ralph --resume                     # resume from existing PRD
./ralph --verbose "prompt"           # debug logging
```

## Configuration

Create `ralph.config.json`:

```json
{
  "model": "opencode/big-pickle",
  "max_iterations": 50,
  "retry_attempts": 3,
  "prd_file": "prd.json"
}
```

**Supported models**: `opencode/big-pickle` (default), `opencode/glm-4.7-free`, `opencode/gpt-5-nano`, `opencode/minimax-m2.1-free`, `claude-code/sonnet`, `claude-code/haiku`, `claude-code/opus`

## Architecture

```
main.go → args → tui/cli → workflow → runner → prd/storage
```

| Package | Purpose |
|---------|---------|
| `internal/args` | CLI argument parsing |
| `internal/cli` | Headless execution mode |
| `internal/tui` | Interactive terminal UI (Bubbletea) |
| `internal/workflow` | Orchestration - PRD generation and story implementation loop |
| `internal/runner` | AI CLI execution (OpenCode/Claude Code subprocess management) |
| `internal/prd` | PRD data models and file I/O with atomic writes and locking |
| `internal/prompt` | Prompt templates for AI |
| `internal/config` | Configuration loading and validation |
| `internal/logger` | Structured logging (slog) |

## How It Works

1. **PRD Generation**: User prompt → AI generates `prd.json` with stories
2. **Implementation Loop**: For each story (by priority):
   - Generate story-specific prompt
   - Run AI CLI
   - AI updates PRD with `passes: true` when complete
   - Repeat until all stories pass or max iterations reached

## Key Patterns

- **Atomic file writes**: Temp file + rename prevents corruption
- **File locking**: `gofrs/flock` for concurrent access safety
- **Event-driven**: Workflow emits typed events consumed by TUI/CLI
- **Factory pattern**: `runner.New()` selects OpenCode vs Claude runner
- **Context cancellation**: Graceful shutdown throughout

## PRD Schema

```go
type PRD struct {
    ProjectName string   `json:"project_name"`
    BranchName  string   `json:"branch_name,omitempty"`
    Context     string   `json:"context,omitempty"`
    TestSpec    string   `json:"test_spec,omitempty"` // Holistic test spec for entire feature
    Stories     []*Story `json:"stories"`
}

type Story struct {
    ID                 string   `json:"id"`
    Title              string   `json:"title"`
    Description        string   `json:"description"`
    AcceptanceCriteria []string `json:"acceptance_criteria"`
    Priority           int      `json:"priority"`      // lower = first
    Passes             bool     `json:"passes"`
    RetryCount         int      `json:"retry_count"`
}
```

## Exit Codes

- `0`: All stories completed
- `1`: Failure
- `2`: Partial completion

## Known Limitations

- No checkpoint/resume within a story (crash loses current story progress)
- No automatic git rollback on failed stories
- No story dependencies (only priority ordering)
- Stories processed sequentially (no parallelization)
