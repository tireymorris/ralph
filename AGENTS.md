# AGENTS.md

## Repository overview
Ralph is a Go CLI/TUI app (with an optional local web UI) that turns a natural-language goal into a PRD, then implements the work story-by-story using an AI coding backend.

Supported backends:
- `pi`
- `opencode`
- `claude`
- Cursor Agent CLI (`cursor-agent`; set `RALPH_RUNNER=cursor`)

## High-level flow
1. Parse CLI args
2. Load config/env
3. Optionally resume from an existing PRD
4. Either show `status`, start the web UI (`ralph web`), or launch the TUI workflow
5. Workflow phases:
   - clarify
   - generate/load PRD
   - review PRD
   - implement stories
   - complete

## Important entrypoints
- `main.go` is only `os.Exit(app.Run(os.Args[1:]))`
- `internal/app` coordinates startup, validation, resume handling, status mode, and TUI startup
- `internal/args` handles CLI parsing/validation
- `internal/shared/config` handles defaults, env overrides, and path validation
- `internal/workflow` owns the long-running state machine and event protocol
- `internal/tui` renders the interactive Bubble Tea UI
- `internal/status` renders non-interactive PRD status output
- `internal/web` serves the embedded React UI and REST/SSE API for runs
- `web/` is the Vite/React source; `npm run build` writes to `internal/web/static/dist/`

## Key behavior
- CLI flags: `--help`, `--dry-run`, `--resume`, `--verbose`, `status`, `web` (optional `--port`)
- Unknown flags are rejected by validation
- `RALPH_RUNNER` selects the runner binary
- PRD path is rooted in the working directory and validated to prevent path traversal
- PRD persistence uses file locks and atomic rename writes
- The workflow emits typed events consumed by the TUI and web UI (SSE)
- Clarification is file-based: AI writes questions to a temp JSON file, workflow reads/removes it, and optional answers are sent back through a channel
- Implementation reloads the PRD each loop, marks stories passed, saves, and continues until complete

## Repository structure
- `internal/prompt` — prompt templates for clarification, PRD generation, and implementation
- `internal/shared/cli` — formatting helpers for PRD/story output
- `internal/shared/constants` — tuning constants
- `internal/shared/logger` — shared slog wrapper
- `internal/shared/prd` — PRD model, validation, locking, and storage
- `internal/shared/runner` — runner abstraction and concrete integrations
- `internal/workflow` — workflow phases, events, persistence, and test execution
- `internal/tui` — Bubble Tea model, view, update loop, and workflow integration
- `internal/status` — plain-text PRD status output
- `internal/web` — HTTP server, static file embedding, handler registration
- `internal/web/handlers` — REST/SSE API handlers (one file per endpoint group) and `respond.go` for shared response helpers
- `internal/web/runner` — `RunController` that bridges workflow execution to the web API
- `internal/web/runs` — run persistence (`Registry`), `IsTerminalStatus`, and `ReadEventTranscript`
- `web/src/api` — API client (`client.ts`) and shared types (`types.ts`)
- `web/src/hooks` — React hooks (`useRunEventStream`, `useRunPolling`, `usePRDLoader`, `useTimelineScroll`)
- `web/src/lib` — pure utility modules (`format.ts`, `timeline.ts`)
- `web/src/pages` — routed page components (`RunDetail`, `NewRunPage`, `HomePage`)
- `web/src/components` — reusable UI components (`ClarifyForm`, `PRDReviewPanel`, `FollowUpComposer`, etc.)

## Testing notes
- Broad test coverage exists across args/config, PRD storage/validation, runner parsing, workflow phases, TUI behavior, and status output
- Top-level integration tests exercise CLI behavior, context persistence, and the web server (`web_integration_test.go`)

## Start here
If you need to understand the system quickly:
1. `README.md`
2. `internal/app/app.go`
3. `internal/workflow/phase_generate.go`
4. `internal/workflow/phase_implement.go`
5. `internal/tui/state.go`
