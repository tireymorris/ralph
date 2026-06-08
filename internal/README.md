# Internal code organization

Ralph keeps UI-specific code separate from shared workflow and domain behavior:

- `internal/workflow` owns shared workflow orchestration, including driver behavior and workflow events used by both interfaces.
- `internal/shared` owns shared domain, configuration, runner, run-state, workdir, and session facade code that is independent of any UI.
- `internal/tui` owns Bubble Tea/Lip Gloss rendering, input handling, and TUI-only operation state.
- `internal/web` owns the Go `net/http` API, run controllers, handlers, and web-only server state.
- `web/src` owns the Vite React TypeScript frontend and its Vitest-covered client behavior.

New behavior needed by both the TUI and web UI should live under `internal/workflow` or `internal/shared` (for example, `internal/shared/session`) rather than being duplicated in `internal/tui` and `internal/web`.

Run the full feature test suite with:

```sh
go test ./... && cd web && npm test && cd ../e2e && npx playwright test
```

When web UI files change, run `go generate ./internal/web/...` before committing embedded assets.
