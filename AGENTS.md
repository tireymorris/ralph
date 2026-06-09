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
3. Optionally resume from an existing PRD (checkpoint-aware via `--resume`)
4. Either show `status`, start the web UI (`ralph web`), or launch the TUI workflow
5. Workflow phases:
   - clarify
   - generate/load PRD
   - review PRD (user approves or revises)
   - implement stories (one runner session per ready story)
   - **implementation review** — critical diff review after each story; may block on findings
   - cleanup (optional; skipped with `--skip-cleanup`)
   - complete

TUI and web share the same `workflow.Driver` → `Executor` engine. Web adds `RunController` registry persistence and SSE; TUI uses `FileReviewLoop` under `.ralph/runs/prd-local/`.

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
- CLI flags: `--help`, `--dry-run`, `--resume`, `--verbose`, `--skip-cleanup`, `status`, `web` (optional `--port`)
- Unknown flags are rejected by validation
- `RALPH_RUNNER` selects the runner binary
- Git worktree required for implementation and web run creation (`workdir.ValidateGit`)
- PRD path is rooted in the working directory and validated to prevent path traversal
- PRD persistence uses file locks and atomic rename writes
- The workflow emits typed events consumed by the TUI and web UI (SSE)
- Clarification is file-based: AI writes questions to a temp JSON file, workflow reads/removes it, and optional answers are sent back through a channel
- Implementation reloads the PRD each loop, marks stories passed, saves, runs implementation review, and continues until complete
- Implementation review: `review.ReviewDiff` after each story; findings block until user continues (TUI: Enter; web: `POST /api/runs/{id}/implementation-review`). Duplicate fingerprint on unchanged diff skips the runner.
- Review loop state persisted in run `meta.json` (`checkpoint`, `review_iteration`, `review_fingerprint`, `last_review_changed_files_hash`, etc.)
- Checkpoints: `prd_review`, `impl_review`, `followup`, `complete` (`internal/shared/runstate`)

## Repository structure
- `internal/prompt` — prompt templates for clarification, PRD generation, implementation, and critical diff review
- `internal/shared/cli` — formatting helpers for PRD/story output
- `internal/shared/constants` — tuning constants
- `internal/shared/gitdiff` — changed-files discovery and hashing for review/cleanup
- `internal/shared/logger` — shared slog wrapper
- `internal/shared/prd` — PRD model, validation, locking, and storage
- `internal/shared/runner` — runner abstraction, concrete integrations, and configurable mock runner
- `internal/shared/runstate` — shared run IDs, checkpoints, and status constants
- `internal/workflow` — workflow phases, events, persistence, test execution, and shared `Driver`
- `internal/workflow/driver.go` — `Driver` encapsulates phase sequencing, clarify state, PRD tracking, and checkpoint resume (shared by TUI and web)
- `internal/workflow/review` — critical diff review, findings parse/fingerprint, transcript storage
- `internal/workflow/file_review_loop.go` — file-backed review loop for TUI/CLI (`prd-local` run id)
- `internal/tui` — Bubble Tea model, view, update loop, and workflow integration (embeds `Driver`)
- `internal/status` — plain-text PRD status output
- `internal/web` — HTTP server, static file embedding, handler registration
- `internal/web/handlers` — REST/SSE API handlers (one file per endpoint group) and `respond.go` for shared response helpers
- `internal/web/runner` — `RunController` that bridges workflow execution to the web API (embeds `Driver`)
- `internal/web/runs` — run persistence (`Registry`), `IsTerminalStatus`, `OngoingLocalPRD`, and `ReadEventTranscript`
- `web/src/api` — API client (`client.ts`) and shared types (`types.ts`)
- `web/src/hooks` — React hooks (`useRunEventStream`, `useRunPolling`, `usePRDLoader`, `useTimelineScroll`)
- `web/src/lib` — pure utility modules (`format.ts`, `timeline.ts`)
- `web/src/pages` — routed page components (`RunDetail`, `NewRunPage`)
- `web/src/components` — reusable UI components (`ClarifyForm`, `PRDReviewPanel`, `ImplementationReviewPanel`, `FollowUpComposer`, etc.)
- `e2e/` — Playwright E2E test suite (builds Go binary + frontend, runs against mock runner)

## State files
Ralph writes these files in the working directory (all covered by `.gitignore`):
- `prd.json` / `prd.json.lock` — PRD and its file lock
- `.ralph_questions.json` — temporary clarification questions (deleted after read)
- `.ralph/runs/<id>/meta.json` — per-run metadata (status, checkpoint, review loop); TUI uses `prd-local`
- `.ralph/runs/<id>/events.ndjson` — per-run event log for SSE replay (web UI)
- `.ralph/runs/<id>/review-*.txt` — implementation review transcripts
- `.ralph/backups/<timestamp>/` — prior state moved aside before a new TUI prompt or `POST /api/runs` (not used by `--resume`; `ralph clean` deletes in place)

## Web API (selected)
- `POST /api/runs/{id}/review` — approve/revise PRD (`waiting_review`)
- `POST /api/runs/{id}/implementation-review` — run recovery from review findings, then continue implementation (`waiting_implementation_review`)
- `POST /api/runs/{id}/resume` — force resume from checkpoint

## Testing notes
- Broad test coverage exists across args/config, PRD storage/validation, runner parsing, workflow phases, TUI behavior, and status output
- Top-level integration tests exercise CLI behavior, context persistence, and the web server (`web_integration_test.go`)
- Frontend unit tests via Vitest (`web/src/**/*.test.{ts,tsx}`)
- E2E tests via Playwright (`e2e/tests/*.spec.ts`) covering run creation, clarify, PRD review, follow-up, cancellation, and sidebar navigation
- CI runs all test suites via GitHub Actions (`.github/workflows/ci.yml`)

## Start here
If you need to understand the system quickly:
1. `README.md`
2. `internal/app/app.go`
3. `internal/workflow/phase_generate.go`
4. `internal/workflow/phase_implement.go`
5. `internal/workflow/phase_implement_review.go`
6. `internal/tui/state.go`
