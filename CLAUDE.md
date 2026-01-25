# Claude Code Context - Ralph Project

**Last Updated**: 2026-01-24
**Analysis Date**: 2026-01-24

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Component Details](#component-details)
4. [Critical Issues](#critical-issues)
5. [Code Quality Issues](#code-quality-issues)
6. [Security Concerns](#security-concerns)
7. [Missing Features](#missing-features)
8. [Architectural Critiques](#architectural-critiques)
9. [Recommendations](#recommendations)
10. [Testing Strategy](#testing-strategy)
11. [File Reference](#file-reference)

---

## Project Overview

**Ralph** is an autonomous software development agent written in Go 1.24 that transforms natural language requirements into working code through iterative user story implementation.

### Core Functionality

Ralph operates in two phases:

1. **Phase 1: PRD Generation**
   - Analyzes natural language prompts and codebase
   - Generates structured Product Requirements Documents (PRDs)
   - Creates user stories with acceptance criteria

2. **Phase 2: Implementation Loop**
   - Iteratively implements each story
   - Writes tests
   - Runs tests
   - Commits changes
   - Updates PRD status

### Operating Modes

- **Interactive TUI Mode**: Rich terminal UI with real-time progress visualization using Charmbracelet Bubbletea
- **Headless CLI Mode**: Non-interactive stdout output for CI/CD pipelines and scripting

### Supported AI Backends

**OpenCode Models**:
- `opencode/big-pickle` (default)
- `opencode/glm-4.7-free`
- `opencode/gpt-5-nano`
- `opencode/minimax-m2.1-free`

**Claude Code Models**:
- `claude-code/claude-3.5-sonnet`
- `claude-code/claude-3.5-haiku`
- `claude-code/claude-3-opus`

---

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────┐
│          Main Entry Point (main.go)         │
├─────────────────────────────────────────────┤
│  ┌──────────────────────────────────────┐   │
│  │   Command Line Interface (args)      │   │
│  └──────────────────────────────────────┘   │
│  ┌────────────────┐         ┌──────────┐   │
│  │  TUI Layer     │   OR    │ CLI Layer│   │
│  │  (Interactive) │         │(Headless)│   │
│  └────────────────┘         └──────────┘   │
│  ┌──────────────────────────────────────┐   │
│  │    Workflow Layer (Executor)         │   │
│  │  - PRD Generation/Loading            │   │
│  │  - Story Implementation              │   │
│  │  - Event Management                  │   │
│  └──────────────────────────────────────┘   │
│  ┌────────────────┐    ┌────────────────┐  │
│  │   Runner       │    │    Prompt      │  │
│  │  (OpenCode or  │    │  Generation    │  │
│  │  Claude Code)  │    │                │  │
│  └────────────────┘    └────────────────┘  │
│  ┌──────────────────────────────────────┐   │
│  │    Configuration & Storage           │   │
│  │  - Config (Config.go)                │   │
│  │  - PRD Management (types.go, storage)│   │
│  │  - Logging                           │   │
│  └──────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

### Layered Architecture

```
CLI/TUI Layer → Workflow Orchestration → Runner (AI Integration) → Storage/Config
```

### Data Flow

1. **User Input** → Command-line arguments
2. **Configuration** → Loaded from `ralph.config.json`
3. **PRD Generation**:
   - User prompt → AI model (via Runner)
   - Codebase analysis by AI
   - Generated `prd.json` with stories
4. **Story Implementation Loop**:
   - Load current PRD state
   - Select next pending story (by priority)
   - Generate story-specific prompt with context
   - Execute AI model
   - AI updates PRD file with `passes: true` when complete
   - Repeat until all stories pass or max iterations reached
5. **Output** → Events emitted to UI/CLI, logs to stderr

### Key Design Patterns

1. **Factory Pattern**: `runner.New()` selects between OpenCode and Claude runners
2. **Event-Driven Architecture**: Workflow emits typed events for UI consumption
3. **Interface-Based Design**: `RunnerInterface` for pluggable AI runners
4. **Context-Based Cancellation**: Uses Go context for graceful shutdown
5. **Buffered Channels**: Large channel buffers (10000) to prevent blocking
6. **Decorator Pattern**: Output filtering for verbose vs. non-verbose modes

---

## Component Details

### 1. Command Line Interface (`internal/args/args.go`)

**Location**: `internal/args/args.go` (98 lines)

**Purpose**: Parses and validates command-line arguments

**Key Features**:
- Flags: `--dry-run`, `--resume`, `--verbose`, `--help`, `run` (headless mode)
- Validates argument combinations
- Provides comprehensive help text

**Arguments**:
- Positional: User prompt (required unless `--resume`)
- `--resume`: Resume from existing PRD
- `--dry-run`: Dry run mode (flag exists but not fully implemented)
- `--verbose`: Enable debug logging
- `--help`: Show help text
- `run`: Enable headless mode

### 2. Configuration System (`internal/config/config.go`)

**Location**: `internal/config/config.go` (111 lines)

**Purpose**: Loads and manages configuration from `ralph.config.json`

**Configuration Schema**:
```json
{
  "model": "opencode/big-pickle",
  "max_iterations": 50,
  "retry_attempts": 3,
  "prd_file": "prd.json"
}
```

**Defaults**:
- `model`: `opencode/big-pickle`
- `max_iterations`: 50
- `retry_attempts`: 3
- `prd_file`: `prd.json`

**Validation**:
- Model must be in supported list
- `max_iterations` must be > 0
- `retry_attempts` must be >= 0
- `prd_file` cannot be empty

**ISSUE**: No validation for path traversal in `prd_file` (see Security Concerns)

### 3. Prompt Generation (`internal/prompt/prompt.go`)

**Location**: `internal/prompt/prompt.go` (128 lines)

**Purpose**: Creates structured prompts for AI models

**Key Functions**:
- `PRDGeneration(userPrompt, prdFile, branchName)`: Generates prompt for PRD creation
- `StoryImplementation(storyID, title, desc, criteria, testSpec, context, prdFile, iteration, completed, total)`: Generates prompt for implementing stories
- `JSONRepair(prdFile, errorMsg)`: Generates prompt for fixing corrupted PRD JSON

**Prompt Strategy**:
- Includes explicit instructions for AI behavior
- References PRD file location for AI to read/write
- Provides context from codebase analysis
- Instructs AI to update PRD when story is complete

### 4. Runner Layer (`internal/runner/`)

**Location**:
- `internal/runner/runner.go` (209 lines) - OpenCode runner
- `internal/runner/claude.go` (160 lines) - Claude Code runner

**Purpose**: Executes external AI CLI tools (OpenCode or Claude Code)

**Interface**:
```go
type RunnerInterface interface {
    Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error
}
```

**Key Features**:
- Abstracts differences between OpenCode and Claude Code CLIs
- Manages subprocess execution with context
- Captures stdout and stderr with configurable buffering
- Filters verbose output based on patterns
- Factory pattern for selecting appropriate runner

**Implementation Details**:
- Uses `os/exec` to spawn CLI processes
- Streams output via channels
- Buffer size: 1MB for scanner
- Detects "verbose" lines to filter noise
- Handles process exit codes

**ISSUE**: Potential goroutine leaks on context cancellation (see Critical Issues)

### 5. PRD Management (`internal/prd/`)

**Location**:
- `internal/prd/types.go` (74 lines) - Data models
- `internal/prd/storage.go` (50 lines) - File I/O

**Data Models**:
```go
type Story struct {
    ID                 string   `json:"id"`                 // Unique identifier
    Title              string   `json:"title"`              // Short title
    Description        string   `json:"description"`        // Detailed requirements
    AcceptanceCriteria []string `json:"acceptance_criteria"` // Conditions for completion
    TestSpec           string   `json:"test_spec,omitempty"` // Testing guidance
    Priority           int      `json:"priority"`           // Implementation order (lower = first)
    Passes             bool     `json:"passes"`             // Completion status
    RetryCount         int      `json:"retry_count"`        // Failed attempts
}

type PRD struct {
    ProjectName string   `json:"project_name"`        // Project description
    BranchName  string   `json:"branch_name,omitempty"` // Git branch name
    Context     string   `json:"context,omitempty"`   // Cached codebase context
    Stories     []*Story `json:"stories"`             // Implementation stories
}
```

**PRD Methods**:
- `NextPendingStory(maxRetries)`: Returns next story to implement (lowest priority, not passing, under retry limit)
- `CompletedCount()`: Returns number of completed stories
- `FailedStories(maxRetries)`: Returns stories that exceeded retry limit
- `AllCompleted()`: Returns true if all stories pass
- `GetStory(id)`: Finds story by ID

**Storage Operations**:
- `Load(cfg)`: Reads PRD from JSON file
- `Save(cfg, prd)`: Writes PRD to JSON file (indented, 0644 permissions)
- `Delete(cfg)`: Removes PRD file
- `Exists(cfg)`: Checks if PRD exists

**CRITICAL ISSUE**: No file locking, no atomic writes, no transaction management (see Critical Issues)

### 6. Workflow Executor (`internal/workflow/workflow.go`)

**Location**: `internal/workflow/workflow.go` (322 lines)

**Purpose**: Orchestrates the entire PRD generation and implementation process

**Key Methods**:
- `RunGenerate(ctx, userPrompt)`: Generates PRD from user prompt
- `RunLoad(ctx)`: Loads existing PRD
- `RunImplementation(ctx, prd)`: Iterates through stories, implementing each

**Event System**:
Emits typed events for UI/CLI updates:
- `EventPRDGenerating`: PRD generation started
- `EventPRDGenerated`: PRD created successfully
- `EventPRDLoaded`: Existing PRD loaded
- `EventStoryStarted`: Story implementation started
- `EventStoryCompleted`: Story finished (success/failure)
- `EventOutput`: Output from AI runner
- `EventError`: Error occurred
- `EventCompleted`: All stories completed
- `EventFailed`: Some stories failed

**Implementation Loop Logic** (workflow.go:161-266):
1. Check context cancellation
2. Reload PRD from disk
3. Check if all stories completed → success
4. Get next pending story (by priority, under retry limit)
5. Check if all remaining stories failed → failure
6. Check iteration limit → failure
7. Start story implementation
8. Run AI with story prompt
9. Reload PRD to check if story marked as complete
10. Update retry count if needed
11. Emit completion event
12. Repeat

**JSON Repair Mechanism** (workflow.go:294-322):
- Detects JSON parse errors in PRD
- Asks AI to fix the JSON (max 2 attempts)
- Falls back to failure if repair doesn't work

**CRITICAL ISSUES**:
- Silent error handling on PRD reload (line 180-181)
- Event dropping when channel full (line 273-274)
- Race conditions on retry count increment (line 255-259)

### 7. CLI Runner (`internal/cli/cli.go`)

**Location**: `internal/cli/cli.go` (159 lines)

**Purpose**: Headless execution mode for CI/CD and scripts

**Features**:
- Processes workflow events and outputs to stdout
- Handles graceful shutdown on interrupt signals (SIGINT, SIGTERM)
- Formats output for terminal and log parsing
- Exit codes: 0 (success), 1 (failure), 2 (partial completion)

**Event Handling**:
- `EventPRDGenerating`: Print status
- `EventPRDGenerated`: Print PRD summary
- `EventPRDLoaded`: Print resume message
- `EventStoryStarted`: Print story info
- `EventStoryCompleted`: Print success/failure with emoji
- `EventOutput`: Print text (filter verbose if not enabled)
- `EventError`: Print error to stderr
- `EventCompleted`: Success message, exit 0
- `EventFailed`: Failure message, exit 1

### 8. TUI (Terminal User Interface) (`internal/tui/`)

**Location**:
- `internal/tui/model.go` (238 lines) - Bubbletea Model
- `internal/tui/view.go` (179 lines) - Rendering logic
- `internal/tui/operations.go` (88 lines) - Async operations
- `internal/tui/styles.go` (~150 lines) - Visual styling
- `internal/tui/logger.go` (~3KB) - Output logging

**Purpose**: Rich interactive terminal UI using Charmbracelet Bubbletea

**Phases**:
- `PhaseInit`: Initialization
- `PhasePRDGeneration`: Generating PRD
- `PhaseImplementation`: Implementing stories
- `PhaseCompleted`: Success
- `PhaseFailed`: Failed

**Display Features**:
- Progress bar showing story completion
- Real-time log viewer with scrolling
- Story list with status icons (✓ completed, ⧗ in progress, ○ pending, ✗ failed)
- Project and branch information
- Styled output with colors and formatting

**Bubbletea Architecture**:
- Model: Holds application state
- Update: Processes messages (events, key presses)
- View: Renders UI from state

**POTENTIAL ISSUE**: Unbounded log buffer could cause memory leaks

### 9. Logging (`internal/logger/logger.go`)

**Location**: `internal/logger/logger.go` (54 lines)

**Purpose**: Structured logging using Go's `slog` package

**Log Levels**:
- Debug: Only when `--verbose` flag is used
- Info: General information
- Warn: Non-fatal issues
- Error: Failures and exceptions

**Output**: Logs go to stderr (not stdout)

---

## Critical Issues

### 1. Race Conditions and Data Integrity ⚠️ CRITICAL

**Severity**: HIGH
**Impact**: Data corruption, lost updates, incorrect retry counts

**Problem**: The PRD file acts as shared state between Ralph and the AI CLI tools, with NO file locking or transaction management.

**Affected Code**:
- `workflow.go:178-181`: PRD reloaded every iteration without checking if AI is still writing
- `workflow.go:255-259`: Retry count increment has TOCTOU (time-of-check-time-of-use) bug
- `storage.go:25-35`: No atomic file operations - crash during Save() could corrupt PRD
- No protection against multiple Ralph instances racing on same PRD

**Race Condition Example**:
```
Thread 1: Load PRD (story-1 retry_count=0)
Thread 2 (AI): Load PRD (story-1 retry_count=0)
Thread 1: Increment retry_count to 1, Save
Thread 2 (AI): Mark story as passes=true, Save
Result: Lost update - either retry_count or passes could be overwritten
```

**Observed Symptoms**:
- JSON corruption (hence the repair mechanism exists)
- Stories being processed twice
- Incorrect retry counts

**Recommendations**:
1. **Immediate**: Implement file locking using `syscall.Flock` or `github.com/gofrs/flock`
2. **Better**: Use atomic file writes (write to temp file, then rename)
3. **Best**: Replace file storage with SQLite for ACID guarantees
4. **Add**: Version/checksum to PRD for conflict detection

**Example Fix**:
```go
// Add to PRD struct
type PRD struct {
    Version int64 `json:"version"` // Incremented on each save
    // ... existing fields
}

// In Save()
func Save(cfg *config.Config, p *PRD) error {
    p.Version++ // Optimistic locking
    // ... write to temp file
    // ... atomic rename
}
```

### 2. Silent Error Swallowing ⚠️ CRITICAL

**Severity**: HIGH
**Impact**: Operates on stale data, potentially infinite loops, incorrect behavior

**Affected Code**:

**workflow.go:178-181**:
```go
p, err := prd.Load(e.cfg)
if err != nil {
    logger.Warn("failed to reload PRD", "error", err)
    // BUG: Continues with old/stale 'p' variable!
}
```

**workflow.go:244-246**:
```go
if loadErr != nil {
    logger.Warn("failed to reload PRD after story", "error", loadErr)
    e.emit(EventStoryCompleted{Story: next, Success: false})
    continue  // BUG: Continues with stale data
}
```

**Consequences**:
- Re-processing already completed stories
- Missing new stories added by AI
- Operating on stale retry counts
- Infinite loop if PRD becomes unreadable

**Recommendation**: Treat reload failures as FATAL errors, not warnings.

**Example Fix**:
```go
p, err := prd.Load(e.cfg)
if err != nil {
    logger.Error("critical: failed to reload PRD", "error", err)
    e.emit(EventError{Err: fmt.Errorf("cannot continue: %w", err)})
    return err
}
```

### 3. Channel Blocking and Event Loss ⚠️ MEDIUM

**Severity**: MEDIUM
**Impact**: Lost visibility, missing critical errors in UI

**Affected Code**:

**workflow.go:268-275**:
```go
func (e *Executor) emit(event Event) {
    if e.eventsCh != nil {
        select {
        case e.eventsCh <- event:
        default:
            logger.Warn("event channel full, dropping event")
            // BUG: Drops events silently
        }
    }
}
```

**Problem**: Uses non-blocking send with 10,000 buffer. If TUI is slow:
- Events get **silently dropped**
- User loses visibility into what's happening
- Output logs might be missing critical errors
- No way to know WHICH events were dropped

**Better Approaches**:
1. Use blocking send with context cancellation
2. Apply backpressure to slow down workflow if UI can't keep up
3. Log which specific events are dropped (include event type)
4. Increase buffer size or make it configurable

**Example Fix**:
```go
func (e *Executor) emit(event Event) {
    if e.eventsCh != nil {
        select {
        case e.eventsCh <- event:
        default:
            logger.Warn("event channel full, dropping event",
                "event_type", fmt.Sprintf("%T", event),
                "event_details", fmt.Sprintf("%+v", event))
            // Could also block here with timeout:
            // ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            // defer cancel()
            // select {
            // case e.eventsCh <- event:
            // case <-ctx.Done():
            //     logger.Error("failed to emit event after timeout")
            // }
        }
    }
}
```

### 4. Subprocess Cleanup and Goroutine Leaks ⚠️ MEDIUM

**Severity**: MEDIUM
**Impact**: Goroutine leaks, zombie processes, resource exhaustion

**Affected Code**:

**runner.go:115-167**:
```go
if err := cmd.Start(); err != nil {
    return fmt.Errorf("failed to start opencode: %w", err)
}

var wg sync.WaitGroup
wg.Add(2)

go func() {
    defer wg.Done()
    scanner := bufio.NewScanner(stdout)
    // ... reads stdout
}()

go func() {
    defer wg.Done()
    scanner := bufio.NewScanner(stderr)
    // ... reads stderr
}()

wg.Wait()    // Wait for goroutines
err = cmd.Wait()  // Wait for process
```

**Problems**:
1. If context is cancelled, process is killed, but goroutines might block on `scanner.Scan()`
2. No explicit timeout for goroutine shutdown
3. Pipes might not close immediately after process termination
4. No cleanup of stdout/stderr pipes on error paths

**Potential Leak Scenario**:
```
1. Context cancelled
2. Process killed
3. Stdout pipe has buffered data
4. Goroutine blocked reading from stdout (pipe not closed yet)
5. Goroutine never exits → leak
```

**Recommendations**:
1. Use `io.Copy` with context-aware readers
2. Explicitly close pipes after process termination
3. Add timeout for goroutine cleanup
4. Use errgroup for better error handling

**Example Fix**:
```go
// Use errgroup for better coordination
import "golang.org/x/sync/errgroup"

func (r *Runner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
    // ... setup cmd ...

    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start opencode: %w", err)
    }

    // Ensure process is killed on context cancellation
    go func() {
        <-ctx.Done()
        if cmd.Process != nil {
            cmd.Process.Kill()
        }
    }()

    var eg errgroup.Group

    eg.Go(func() error {
        defer stdout.Close()
        scanner := bufio.NewScanner(stdout)
        // ... scan stdout
        return scanner.Err()
    })

    eg.Go(func() error {
        defer stderr.Close()
        scanner := bufio.NewScanner(stderr)
        // ... scan stderr
        return scanner.Err()
    })

    // Wait for both readers with timeout
    readersDone := make(chan error)
    go func() {
        readersDone <- eg.Wait()
    }()

    select {
    case <-readersDone:
    case <-time.After(5 * time.Second):
        logger.Warn("goroutines did not exit in time")
    }

    return cmd.Wait()
}
```

### 5. Undefined Behavior on Story State ⚠️ MEDIUM

**Severity**: MEDIUM
**Impact**: Stories processed multiple times, wasted AI calls

**Problem**: Stories only have two states: `passes: false` and `passes: true`. No way to mark a story as "in progress".

**Affected Code**:
- `types.go:3-12`: Story struct only has `Passes bool`
- `workflow.go:190-196`: `NextPendingStory()` returns any story with `passes: false`

**Race Scenario**:
```
Iteration 1: Start story-1 (passes=false)
AI is working on story-1 (takes 5 minutes)
Iteration 2: Reload PRD, story-1 still shows passes=false
           → NextPendingStory() returns story-1 again!
           → Start story-1 again in parallel
```

**Consequences**:
- Wasted AI API calls
- Potential conflicts if both instances modify same files
- Confusing logs and UI

**Recommendation**: Add proper state machine for stories.

**Example Fix**:
```go
type StoryStatus string

const (
    StoryStatusPending    StoryStatus = "pending"
    StoryStatusInProgress StoryStatus = "in_progress"
    StoryStatusCompleted  StoryStatus = "completed"
    StoryStatusFailed     StoryStatus = "failed"
)

type Story struct {
    ID                 string      `json:"id"`
    Status             StoryStatus `json:"status"`
    StartedAt          *time.Time  `json:"started_at,omitempty"`
    CompletedAt        *time.Time  `json:"completed_at,omitempty"`
    // ... other fields
}

func (p *PRD) NextPendingStory(maxRetries int) *Story {
    for _, story := range p.Stories {
        if story.Status != StoryStatusPending {
            continue
        }
        if story.RetryCount >= maxRetries {
            continue
        }
        // ... select by priority
    }
}
```

### 6. JSON Repair is Fragile ⚠️ LOW

**Severity**: LOW
**Impact**: Failed repair loses all work, hard to debug

**Affected Code**: `workflow.go:294-322`

**Problems**:
1. AI has no context on WHY the JSON is corrupted
2. Could apply the wrong fix
3. Only 2 attempts before giving up
4. No backup of corrupted file for debugging
5. Repair prompt may not be specific enough

**Better Approach**:
1. Keep transaction log of PRD changes
2. Backup PRD before each write
3. Use JSON schema validation before writing
4. Implement automatic rollback on corruption
5. Parse error more carefully to give AI better context

**Example Fix**:
```go
func (e *Executor) repairPRD(ctx context.Context, parseErr error) (*prd.PRD, error) {
    // 1. Backup corrupted file
    corruptedPath := e.cfg.PRDPath() + ".corrupted." + time.Now().Format("20060102-150405")
    if data, err := os.ReadFile(e.cfg.PRDPath()); err == nil {
        os.WriteFile(corruptedPath, data, 0644)
        logger.Info("backed up corrupted PRD", "path", corruptedPath)
    }

    // 2. Extract specific JSON error location
    errorContext := extractJSONErrorContext(parseErr, e.cfg.PRDPath())

    // 3. Try repair with better prompt
    repairPrompt := prompt.JSONRepair(e.cfg.PRDFile, errorContext)
    // ... rest of repair logic
}

func extractJSONErrorContext(err error, filePath string) string {
    // Parse "invalid character '}' at line 45, column 12" type errors
    // Return surrounding lines from file for context
}
```

### 7. File Permissions Too Permissive ⚠️ LOW

**Severity**: LOW (Security)
**Impact**: PRD file readable by other users on system

**Affected Code**: `storage.go:31`

```go
os.WriteFile(cfg.PRDPath(), data, 0644)
```

**Problem**: Uses `0644` permissions (owner read/write, group/others read)
- PRD might contain sensitive project information
- Other users can read the PRD
- Should be `0600` for user-only access

**Fix**:
```go
os.WriteFile(cfg.PRDPath(), data, 0600)  // User read/write only
```

### 8. No Git Safety Checks ⚠️ MEDIUM

**Severity**: MEDIUM
**Impact**: Lost work, corrupted git history, wrong branch

**Problem**: AI is instructed to create commits, but Ralph doesn't verify:
- Git repo exists before starting
- No uncommitted changes that might be lost
- Branch name in PRD exists
- Not in detached HEAD state
- Not about to commit to main/master directly

**Affected Code**: No pre-flight checks in `workflow.RunGenerate()` or `RunImplementation()`

**Recommendation**: Add git safety checks before starting work.

**Example Implementation**:
```go
func validateGitState(cfg *config.Config, branchName string) error {
    // Check if git repo exists
    cmd := exec.Command("git", "rev-parse", "--git-dir")
    cmd.Dir = cfg.WorkDir
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("not a git repository")
    }

    // Check for uncommitted changes
    cmd = exec.Command("git", "status", "--porcelain")
    cmd.Dir = cfg.WorkDir
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("failed to check git status: %w", err)
    }
    if len(output) > 0 {
        return fmt.Errorf("working directory has uncommitted changes")
    }

    // Check if on correct branch
    cmd = exec.Command("git", "branch", "--show-current")
    cmd.Dir = cfg.WorkDir
    output, err = cmd.Output()
    if err != nil {
        return fmt.Errorf("failed to get current branch: %w", err)
    }
    currentBranch := strings.TrimSpace(string(output))
    if currentBranch != branchName {
        return fmt.Errorf("not on branch %q (currently on %q)", branchName, currentBranch)
    }

    return nil
}
```

---

## Code Quality Issues

### 1. Inconsistent Logging

**Location**: `workflow.go:233`

```go
logger.Debug("opencode returned error", "story_id", next.ID, "error", err)
```

**Problem**: Logs "opencode" even when using Claude Code runner.

**Fix**: Use generic term like "ai runner" or "runner".

### 2. Magic Numbers

**Locations**:
- `workflow.go:106,213`: `10000` buffer size - why this number?
- `runner.go:122`: `1024*1024` buffer size - no constant defined
- `workflow.go:15`: `maxJSONRepairAttempts = 2` - should be configurable

**Recommendation**: Define constants with explanatory comments.

```go
const (
    // EventChannelBuffer controls how many events can be queued before blocking.
    // Set to 10000 to handle burst output from AI runners.
    EventChannelBuffer = 10000

    // ScannerBufferSize is the maximum line size for reading AI output.
    // Set to 1MB to handle very long output lines.
    ScannerBufferSize = 1024 * 1024

    // MaxJSONRepairAttempts is how many times we'll try to fix corrupted PRD JSON
    // before giving up.
    MaxJSONRepairAttempts = 2
)
```

### 3. Misleading Function Names

**Location**: `runner.go:169-208`

**Function**: `isVerboseLine()`

**Problem**: Name suggests it detects "verbose" lines, but it actually detects **OpenCode-specific internal log patterns**.

**Issues**:
- Claude Code might have different log patterns
- Function is hardcoded to OpenCode log format
- Patterns like "service=bus", "type=message." are OpenCode internals

**Recommendation**: Rename and make configurable per runner.

```go
// OpenCodeRunner
func (r *OpenCodeRunner) isInternalLog(line string) bool {
    // OpenCode-specific patterns
}

// ClaudeRunner
func (r *ClaudeRunner) isInternalLog(line string) bool {
    // Claude-specific patterns
}
```

### 4. No Unit Tests

**Problem**: No `*_test.go` files found in codebase.

**Missing Test Coverage**:
1. **PRD Logic** (`internal/prd/types.go`):
   - What if duplicate story IDs?
   - What if negative priority values?
   - What if `Stories` slice is nil?
   - Edge cases in `NextPendingStory()`

2. **Configuration** (`internal/config/config.go`):
   - Invalid JSON handling
   - Path traversal in `prd_file`
   - Validation edge cases

3. **Workflow** (`internal/workflow/workflow.go`):
   - Event ordering
   - Error handling paths
   - PRD reload logic

4. **Runner** (`internal/runner/runner.go`):
   - Mock command execution
   - Output parsing
   - Error handling

**Recommendation**: Add comprehensive unit tests, aim for >80% coverage.

**Example Test**:
```go
func TestPRD_NextPendingStory(t *testing.T) {
    tests := []struct {
        name       string
        prd        *PRD
        maxRetries int
        want       *Story
    }{
        {
            name: "returns lowest priority pending story",
            prd: &PRD{
                Stories: []*Story{
                    {ID: "1", Priority: 2, Passes: false, RetryCount: 0},
                    {ID: "2", Priority: 1, Passes: false, RetryCount: 0},
                },
            },
            maxRetries: 3,
            want:       &Story{ID: "2", Priority: 1},
        },
        {
            name: "skips stories that exceeded retry limit",
            prd: &PRD{
                Stories: []*Story{
                    {ID: "1", Priority: 1, Passes: false, RetryCount: 3},
                    {ID: "2", Priority: 2, Passes: false, RetryCount: 0},
                },
            },
            maxRetries: 3,
            want:       &Story{ID: "2", Priority: 2},
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.prd.NextPendingStory(tt.maxRetries)
            if got == nil && tt.want != nil {
                t.Errorf("got nil, want %v", tt.want)
            }
            if got != nil && tt.want == nil {
                t.Errorf("got %v, want nil", got)
            }
            if got != nil && tt.want != nil && got.ID != tt.want.ID {
                t.Errorf("got ID %q, want %q", got.ID, tt.want.ID)
            }
        })
    }
}
```

### 5. Error Context Missing

**Problem**: Many errors lack context about what operation failed.

**Examples**:

**workflow.go:116**:
```go
return nil, err  // Which prompt? Which model? What was the output?
```

**storage.go:14**:
```go
return nil, fmt.Errorf("failed to read PRD file: %w", err)
// Which file path? Working directory?
```

**Recommendation**: Wrap errors with context at each layer.

**Example Fix**:
```go
// In workflow.go
if err := e.runner.Run(ctx, prdPrompt, outputCh); err != nil {
    return nil, fmt.Errorf("PRD generation failed: model=%s, prompt_len=%d: %w",
        e.cfg.Model, len(prdPrompt), err)
}

// In storage.go
func Load(cfg *config.Config) (*PRD, error) {
    path := cfg.PRDPath()
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read PRD file %q: %w", path, err)
    }
    // ...
}
```

### 6. Potential Memory Issues in TUI

**Location**: `internal/tui/logger.go`

**Concern**: Without seeing full implementation, unbounded log buffers could grow indefinitely during long-running sessions.

**Recommendation**: Implement circular buffer with max size:

```go
type LogBuffer struct {
    lines    []string
    maxLines int
    mu       sync.Mutex
}

func (lb *LogBuffer) Append(line string) {
    lb.mu.Lock()
    defer lb.mu.Unlock()

    lb.lines = append(lb.lines, line)
    if len(lb.lines) > lb.maxLines {
        lb.lines = lb.lines[1:] // Drop oldest
    }
}
```

---

## Security Concerns

### 1. Command Injection via Prompt ⚠️ LOW

**Severity**: LOW (Go's exec.Command is safe from shell injection)
**Impact**: Potential for extremely long prompts, special character issues

**Affected Code**: `runner.go:88`

```go
args = append(args, prompt)
```

**Problem**: User prompt passed directly as command argument.

**Concerns**:
- Very long prompts could exceed OS argument limits (typically 128KB-2MB)
- Special characters might be misinterpreted by AI CLI
- No length validation
- No sanitization

**Mitigation**: Go's `exec.Command` doesn't use shell, so no shell injection risk.

**Recommendation**: Add validation:

```go
const MaxPromptLength = 100000 // 100KB

func (r *Runner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
    if len(prompt) > MaxPromptLength {
        return fmt.Errorf("prompt too long: %d bytes (max %d)", len(prompt), MaxPromptLength)
    }

    // ... rest of implementation
}
```

### 2. Path Traversal in Config ⚠️ MEDIUM

**Severity**: MEDIUM
**Impact**: PRD file could be written to arbitrary locations

**Affected Code**: `config.go:65-67`

```go
if fileCfg.PRDFile != "" {
    cfg.PRDFile = fileCfg.PRDFile
}
```

**Problem**: No validation that `PRDFile` doesn't contain:
- Absolute paths: `/etc/passwd`, `/tmp/prd.json`
- Path traversal: `../../../sensitive.json`, `../../.ssh/authorized_keys`
- Special characters: `prd.json; rm -rf /`

**Attack Scenario**:
```json
{
  "prd_file": "../../../.ssh/authorized_keys"
}
```

Result: Ralph writes PRD to `~/.ssh/authorized_keys`, potentially corrupting SSH config.

**Recommendation**: Validate that `prd_file` is a simple filename with no path components.

**Fix**:
```go
import "path/filepath"

func (c *Config) Validate() error {
    // ... existing validation

    // Validate prd_file is just a filename, no path components
    if filepath.Base(c.PRDFile) != c.PRDFile {
        return fmt.Errorf("prd_file must be a simple filename, got %q", c.PRDFile)
    }

    if filepath.IsAbs(c.PRDFile) {
        return fmt.Errorf("prd_file cannot be an absolute path, got %q", c.PRDFile)
    }

    if strings.Contains(c.PRDFile, "..") {
        return fmt.Errorf("prd_file cannot contain .., got %q", c.PRDFile)
    }

    return nil
}
```

### 3. No Input Validation on PRD Fields ⚠️ LOW

**Severity**: LOW
**Impact**: Malicious AI could craft PRD to cause issues

**Problem**: When loading PRD from disk (which AI wrote), no validation on:
- Length of `Context` field (could be gigabytes)
- Duplicate story IDs
- Empty story IDs
- Extremely large number of stories
- Invalid priority values (negative, max int)
- Extremely long story descriptions

**Affected Code**: `storage.go:11-23`

**Potential Issues**:
1. **Memory exhaustion**: 1GB `Context` field loads entire file into memory
2. **Infinite loops**: Duplicate story IDs could confuse `GetStory()`
3. **Integer overflow**: Priority values near max int

**Recommendation**: Add validation after loading PRD.

**Example Fix**:
```go
const (
    MaxContextSize      = 1 * 1024 * 1024 // 1MB
    MaxStories          = 1000
    MaxStoryDescSize    = 100 * 1024      // 100KB
    MaxAcceptanceCriteria = 50
)

func (p *PRD) Validate() error {
    if len(p.Context) > MaxContextSize {
        return fmt.Errorf("context too large: %d bytes (max %d)", len(p.Context), MaxContextSize)
    }

    if len(p.Stories) > MaxStories {
        return fmt.Errorf("too many stories: %d (max %d)", len(p.Stories), MaxStories)
    }

    seen := make(map[string]bool)
    for i, story := range p.Stories {
        if story.ID == "" {
            return fmt.Errorf("story %d has empty ID", i)
        }

        if seen[story.ID] {
            return fmt.Errorf("duplicate story ID: %q", story.ID)
        }
        seen[story.ID] = true

        if len(story.Description) > MaxStoryDescSize {
            return fmt.Errorf("story %q description too large: %d bytes", story.ID, len(story.Description))
        }

        if story.Priority < 0 {
            return fmt.Errorf("story %q has negative priority: %d", story.ID, story.Priority)
        }

        if len(story.AcceptanceCriteria) > MaxAcceptanceCriteria {
            return fmt.Errorf("story %q has too many acceptance criteria: %d", story.ID, len(story.AcceptanceCriteria))
        }
    }

    return nil
}

// In storage.go Load()
func Load(cfg *config.Config) (*PRD, error) {
    // ... existing load code

    if err := p.Validate(); err != nil {
        return nil, fmt.Errorf("PRD validation failed: %w", err)
    }

    return &p, nil
}
```

---

## Missing Features

### 1. No Progress Persistence

**Problem**: If Ralph crashes mid-implementation:
- Current story progress is lost
- No checkpoint to resume from exact position
- Iteration counter resets to 0
- No record of what AI was doing when crash occurred

**Impact**: User must restart from beginning or manually inspect PRD.

**Recommendation**: Add checkpoint mechanism.

**Example Implementation**:
```go
type Checkpoint struct {
    Timestamp    time.Time   `json:"timestamp"`
    CurrentStory string      `json:"current_story"`
    Iteration    int         `json:"iteration"`
    Phase        string      `json:"phase"` // "prd_generation", "story_implementation"
}

// Save checkpoint before each story
func (e *Executor) saveCheckpoint(storyID string, iteration int) error {
    cp := Checkpoint{
        Timestamp:    time.Now(),
        CurrentStory: storyID,
        Iteration:    iteration,
        Phase:        "story_implementation",
    }

    data, _ := json.Marshal(cp)
    path := e.cfg.ConfigPath("ralph.checkpoint.json")
    return os.WriteFile(path, data, 0600)
}

// Load checkpoint on resume
func (e *Executor) loadCheckpoint() (*Checkpoint, error) {
    path := e.cfg.ConfigPath("ralph.checkpoint.json")
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cp Checkpoint
    err = json.Unmarshal(data, &cp)
    return &cp, err
}
```

### 2. No Rollback Mechanism

**Problem**: If story implementation breaks the codebase:
- No automatic rollback
- No git stash of changes
- User must manually revert using git

**Impact**: Failed story implementation could leave codebase in broken state.

**Recommendation**: Implement git-based rollback.

**Example Implementation**:
```go
// Before story implementation
func (e *Executor) createSavepoint(storyID string) (string, error) {
    // Create git tag for rollback
    tagName := fmt.Sprintf("ralph-savepoint-%s-%d", storyID, time.Now().Unix())

    cmd := exec.Command("git", "tag", tagName)
    cmd.Dir = e.cfg.WorkDir
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("failed to create savepoint: %w", err)
    }

    return tagName, nil
}

// After story fails
func (e *Executor) rollback(tagName string) error {
    // Reset to savepoint
    cmd := exec.Command("git", "reset", "--hard", tagName)
    cmd.Dir = e.cfg.WorkDir
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to rollback: %w", err)
    }

    // Delete tag
    exec.Command("git", "tag", "-d", tagName).Run()

    return nil
}
```

### 3. No Dependency Management Between Stories

**Problem**: PRD format doesn't support:
- "Story B depends on Story A"
- Parallel story execution
- Story grouping/milestones

**Current Limitation**: Stories only have `Priority int` for ordering.

**Impact**: Can't express dependencies like "implement API endpoint before frontend".

**Recommendation**: Add dependency graph to PRD.

**Example Schema**:
```go
type Story struct {
    ID                 string   `json:"id"`
    DependsOn          []string `json:"depends_on,omitempty"` // Story IDs this depends on
    BlockedBy          []string `json:"-"`                     // Computed: stories blocking this one
    // ... existing fields
}

func (p *PRD) NextPendingStory(maxRetries int) *Story {
    for _, story := range p.Stories {
        if story.Passes || story.RetryCount >= maxRetries {
            continue
        }

        // Check if all dependencies are completed
        allDepsMet := true
        for _, depID := range story.DependsOn {
            dep := p.GetStory(depID)
            if dep == nil || !dep.Passes {
                allDepsMet = false
                break
            }
        }

        if allDepsMet {
            return story
        }
    }
    return nil
}
```

### 4. Limited Observability

**Problem**: No metrics or telemetry:
- How long each story takes
- Token usage per story
- Success/failure rates
- Performance trends over time

**Impact**: Can't optimize or debug performance issues.

**Recommendation**: Add metrics collection.

**Example Implementation**:
```go
type StoryMetrics struct {
    StoryID      string        `json:"story_id"`
    StartTime    time.Time     `json:"start_time"`
    EndTime      time.Time     `json:"end_time"`
    Duration     time.Duration `json:"duration"`
    Success      bool          `json:"success"`
    RetryCount   int           `json:"retry_count"`
    OutputLines  int           `json:"output_lines"`
}

type SessionMetrics struct {
    StartTime     time.Time       `json:"start_time"`
    EndTime       time.Time       `json:"end_time"`
    TotalStories  int             `json:"total_stories"`
    Completed     int             `json:"completed"`
    Failed        int             `json:"failed"`
    StoryMetrics  []StoryMetrics  `json:"story_metrics"`
}

// Save metrics after session
func (e *Executor) saveMetrics(metrics *SessionMetrics) error {
    data, _ := json.MarshalIndent(metrics, "", "  ")
    path := e.cfg.ConfigPath("ralph.metrics.json")
    return os.WriteFile(path, data, 0644)
}
```

### 5. Dry-Run Not Implemented

**Problem**: `--dry-run` flag exists but doesn't actually skip AI execution.

**Affected Code**:
- `args.go`: Defines `DryRun` flag
- `workflow.go`: Doesn't check dry-run mode

**Expected Behavior**: In dry-run mode:
- Generate PRD but don't implement stories
- Show what would be done without doing it
- Validate prompts and configuration

**Recommendation**: Implement dry-run logic.

**Example Fix**:
```go
// In workflow.go
type Executor struct {
    cfg      *config.Config
    eventsCh chan Event
    runner   runner.RunnerInterface
    dryRun   bool  // Add this
}

func (e *Executor) RunImplementation(ctx context.Context, p *PRD) error {
    // ... existing setup

    if e.dryRun {
        logger.Info("dry-run mode: skipping AI execution")
        e.emit(EventOutput{Output{Text: fmt.Sprintf(
            "[DRY RUN] Would implement story: %s", next.Title)}})
        continue
    }

    // ... actual implementation
}
```

---

## Architectural Critiques

### 1. File-Based State is Brittle

**Current Approach**: Uses `prd.json` as source of truth.

**Fundamental Problems**:
- **No ACID guarantees**: Partial writes can corrupt state
- **Prone to corruption**: As evidenced by JSON repair mechanism
- **Hard to debug**: No audit trail of changes
- **Can't handle concurrent access**: Multiple instances will conflict
- **No versioning**: Can't track who changed what when
- **No transactions**: Can't group related changes atomically

**Better Approaches**:

1. **SQLite** (Recommended):
```go
// Schema
CREATE TABLE stories (
    id TEXT PRIMARY KEY,
    title TEXT,
    description TEXT,
    status TEXT,
    priority INTEGER,
    retry_count INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE story_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    story_id TEXT,
    field_name TEXT,
    old_value TEXT,
    new_value TEXT,
    changed_at TIMESTAMP,
    changed_by TEXT
);
```

Advantages:
- ACID transactions
- Concurrent access with locking
- Audit trail via `story_changes` table
- Can query/filter stories efficiently
- Schema versioning and migrations

2. **Event Sourcing**:
```go
type Event interface {
    ApplyTo(*PRD) error
}

type StoryCreatedEvent struct {
    StoryID string
    Title   string
    // ...
}

type StoryCompletedEvent struct {
    StoryID string
    CompletedAt time.Time
}

// Rebuild state by replaying events
func ReplayEvents(events []Event) (*PRD, error) {
    prd := &PRD{}
    for _, event := range events {
        if err := event.ApplyTo(prd); err != nil {
            return nil, err
        }
    }
    return prd, nil
}
```

Advantages:
- Complete audit trail
- Can replay to any point in time
- Can't lose data from corruption
- Easy to debug what happened

### 2. Tight Coupling to CLI Tools

**Current Approach**: Shells out to `opencode` and `claude-code` CLIs.

**Problems**:
- **No version control**: CLI might change behavior between versions
- **Dependent on CLI stability**: CLI bugs become Ralph bugs
- **Hard to mock/test**: Must mock subprocess execution
- **Can't retry individual API calls**: Must re-run entire CLI
- **Limited error handling**: Only get exit code, not structured errors
- **No streaming control**: Can't pause/resume API calls

**Better Approach**: Use official SDKs/APIs directly.

**Example with Anthropic SDK**:
```go
import "github.com/anthropics/anthropic-sdk-go"

type APIRunner struct {
    client *anthropic.Client
}

func (r *APIRunner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
    stream, err := r.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
        Model: anthropic.ModelClaude3_5Sonnet,
        Messages: []anthropic.MessageParam{
            anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
        },
        MaxTokens: 4096,
    })

    if err != nil {
        return err
    }

    for stream.Next() {
        event := stream.Current()
        // Handle different event types
        switch e := event.(type) {
        case anthropic.ContentBlockDeltaEvent:
            outputCh <- OutputLine{Text: e.Delta.Text}
        // ... other event types
        }
    }

    return stream.Err()
}
```

Advantages:
- Direct API access
- Structured errors
- Retry individual calls
- Better testing (mock client)
- Version control via go.mod

### 3. Synchronous Story Processing

**Current Approach**: Stories processed one at a time sequentially.

**Limitation**: Even independent stories must wait for previous ones.

**Potential Optimization**: Parallel story execution (if dependency graph added).

**Example**:
```go
// Identify stories that can run in parallel
func (p *PRD) GetParallelizableStories(maxRetries int) [][]*Story {
    // Returns groups of stories that can run in parallel
    // Stories in same group have no dependencies on each other

    var groups [][]*Story

    // Simple algorithm: group by dependency depth
    depthMap := make(map[*Story]int)
    for _, story := range p.Stories {
        depthMap[story] = p.getDependencyDepth(story)
    }

    // Group by depth
    grouped := make(map[int][]*Story)
    for story, depth := range depthMap {
        grouped[depth] = append(grouped[depth], story)
    }

    // Convert to slice of groups
    for depth := 0; depth < len(grouped); depth++ {
        if stories, ok := grouped[depth]; ok {
            groups = append(groups, stories)
        }
    }

    return groups
}

// Execute stories in parallel within each group
func (e *Executor) RunImplementationParallel(ctx context.Context, p *PRD) error {
    groups := p.GetParallelizableStories(e.cfg.RetryAttempts)

    for _, group := range groups {
        // Run all stories in this group in parallel
        var wg sync.WaitGroup
        for _, story := range group {
            wg.Add(1)
            go func(s *Story) {
                defer wg.Done()
                e.implementStory(ctx, s)
            }(story)
        }
        wg.Wait()

        // Check if all succeeded before moving to next group
    }

    return nil
}
```

### 4. Event System Lacks Replay

**Current Approach**: Events are fire-and-forget via channel.

**Problems**:
- Can't replay to rebuild UI state
- No event sourcing
- Hard to debug what happened
- Events can be dropped (non-blocking send)

**Better Approach**: Store events for replay.

**Example**:
```go
type EventStore struct {
    events []Event
    mu     sync.RWMutex
}

func (es *EventStore) Append(event Event) {
    es.mu.Lock()
    defer es.mu.Unlock()
    es.events = append(es.events, event)
}

func (es *EventStore) Replay(handler func(Event)) {
    es.mu.RLock()
    defer es.mu.RUnlock()

    for _, event := range es.events {
        handler(event)
    }
}

func (es *EventStore) Save(path string) error {
    es.mu.RLock()
    defer es.mu.RUnlock()

    data, err := json.Marshal(es.events)
    if err != nil {
        return err
    }

    return os.WriteFile(path, data, 0644)
}

// In TUI, can replay events to rebuild state after crash
func (m *Model) RecoverFromCrash() error {
    store := &EventStore{}
    if err := store.Load("ralph.events.json"); err != nil {
        return err
    }

    store.Replay(func(event Event) {
        m.handleEvent(event)
    })

    return nil
}
```

---

## Recommendations

### Priority Matrix

| Priority | Category | Item | Estimated Effort | Impact |
|----------|----------|------|------------------|--------|
| P0 | Critical | Fix silent error handling (workflow.go:180) | 1 hour | High |
| P0 | Critical | Implement file locking for PRD | 2-4 hours | High |
| P0 | Security | Validate config path traversal | 30 min | Medium |
| P1 | Critical | Fix subprocess cleanup/leaks | 2 hours | Medium |
| P1 | Quality | Add comprehensive unit tests | 1-2 days | High |
| P1 | Feature | Add story status state machine | 3-4 hours | Medium |
| P1 | Safety | Add git safety checks | 2 hours | Medium |
| P2 | Quality | Add error context to all errors | 2 hours | Low |
| P2 | Quality | Fix inconsistent logging | 1 hour | Low |
| P2 | Quality | Define constants for magic numbers | 30 min | Low |
| P2 | Feature | Implement dry-run mode | 2 hours | Low |
| P2 | Feature | Add progress persistence | 3-4 hours | Medium |
| P3 | Architecture | Migrate to SQLite storage | 2-3 days | High |
| P3 | Architecture | Use API SDKs instead of CLI | 3-5 days | High |
| P3 | Feature | Add story dependencies | 1-2 days | Medium |
| P3 | Feature | Add metrics/observability | 1-2 days | Medium |
| P3 | Feature | Implement rollback mechanism | 1 day | Medium |

### Implementation Phases

#### Phase 1: Critical Fixes (1-2 days)
1. Fix silent error handling
2. Implement file locking
3. Validate config paths
4. Add git safety checks
5. Fix subprocess cleanup

#### Phase 2: Testing & Quality (2-3 days)
1. Add unit tests for all packages
2. Add integration tests
3. Add error context
4. Fix logging inconsistencies
5. Define constants

#### Phase 3: Features (1 week)
1. Story status state machine
2. Dry-run implementation
3. Progress persistence
4. Metrics collection
5. Story dependencies

#### Phase 4: Architecture (2-3 weeks)
1. Migrate to SQLite
2. Implement API SDKs
3. Add event sourcing
4. Parallel story execution
5. Rollback mechanism

---

## Testing Strategy

### Unit Tests

**Package**: `internal/prd`
- Test all PRD methods with edge cases
- Nil checks, empty slices, duplicate IDs
- Priority ordering
- Retry limit logic

**Package**: `internal/config`
- Config loading with valid/invalid JSON
- Default values
- Validation edge cases
- Path traversal attacks

**Package**: `internal/workflow`
- Event emission and ordering
- Error handling paths
- PRD reload logic
- JSON repair mechanism

**Package**: `internal/runner`
- Mock command execution
- Output parsing
- Verbose line detection
- Error handling

### Integration Tests

1. **End-to-End PRD Generation**:
   - Mock AI CLI to return known PRD
   - Verify PRD loaded correctly
   - Test resume flow

2. **Story Implementation Loop**:
   - Mock AI CLI to update PRD stories
   - Verify retry logic
   - Test max iterations

3. **Error Recovery**:
   - Test corrupted JSON repair
   - Test failed story handling
   - Test context cancellation

### Test Helpers

```go
// test/helpers.go

func MockRunner(responses map[string]string) runner.RunnerInterface {
    return &mockRunner{responses: responses}
}

func CreateTestPRD(stories ...*prd.Story) *prd.PRD {
    return &prd.PRD{
        ProjectName: "test-project",
        Stories:     stories,
    }
}

func CreateTestStory(id string, priority int, passes bool) *prd.Story {
    return &prd.Story{
        ID:       id,
        Priority: priority,
        Passes:   passes,
    }
}
```

---

## File Reference

### Directory Structure

```
/Users/tmorris/workspace/ralph/
├── main.go (85 lines)                    # Entry point
├── go.mod                                 # Dependencies
├── go.sum                                 # Lock file
├── ralph.config.json                     # Configuration
├── README.md                             # Documentation
├── PRD.md                                # Project PRD
├── CLAUDE.md                             # This file
├── .github/
│   └── workflows/ci.yml                  # GitHub Actions
│
└── internal/
    ├── args/
    │   └── args.go (98 lines)            # CLI argument parsing
    │
    ├── cli/
    │   └── cli.go (159 lines)            # Headless runner
    │
    ├── config/
    │   └── config.go (111 lines)         # Configuration management
    │
    ├── logger/
    │   └── logger.go (54 lines)          # Logging
    │
    ├── prd/
    │   ├── types.go (74 lines)           # PRD data models
    │   └── storage.go (50 lines)         # File I/O
    │
    ├── prompt/
    │   └── prompt.go (128 lines)         # Prompt templates
    │
    ├── runner/
    │   ├── runner.go (209 lines)         # OpenCode runner
    │   └── claude.go (160 lines)         # Claude Code runner
    │
    ├── tui/
    │   ├── model.go (238 lines)          # Bubbletea model
    │   ├── view.go (179 lines)           # TUI rendering
    │   ├── operations.go (88 lines)      # Async operations
    │   ├── styles.go (~150 lines)        # Styling
    │   ├── logger.go (~3KB)              # Output logging
    │   └── utils.go                      # Utilities
    │
    └── workflow/
        └── workflow.go (322 lines)       # Orchestration engine
```

### Key Files to Review

When making changes, pay special attention to these files:

1. **workflow.go** - Core orchestration logic, most critical issues here
2. **storage.go** - File I/O, race condition risks
3. **runner.go** - Subprocess management, potential leaks
4. **types.go** - Data models, need better state management
5. **config.go** - Security validation needed

### Code Statistics

- **Total Lines**: ~7,072 lines of Go code (excluding tests)
- **Test Coverage**: 0% (no tests currently)
- **External Dependencies**: 5 main packages (Charmbracelet libs)
- **Supported Models**: 7 (4 OpenCode + 3 Claude Code)

---

## Overall Assessment

**Grade: B-**

**Production Readiness: NOT READY**

### Strengths ✅
- Clean architecture with good separation of concerns
- Dual-mode UI (TUI + headless) is well-executed
- Event-driven design is solid
- Code is generally readable and well-organized
- Good use of Go idioms and standard library

### Weaknesses ❌
- Critical race conditions around file I/O
- Silent error handling that could cause data loss
- No tests
- Brittle file-based state management
- Missing production-readiness features (monitoring, rollback)
- Security vulnerabilities (path traversal)
- No input validation

### Best Suited For
- Personal projects
- Prototyping
- Proof-of-concept work
- Environments where occasional failures are acceptable

### Not Recommended For
- Production use
- Team environments
- CI/CD pipelines (until hardened)
- Mission-critical code generation

### Next Steps

1. **Immediate**: Fix P0 critical issues (error handling, file locking, security)
2. **Short-term**: Add comprehensive test coverage
3. **Medium-term**: Implement missing features (dry-run, rollback, metrics)
4. **Long-term**: Consider architectural improvements (SQLite, API SDKs)

---

**End of Document**
