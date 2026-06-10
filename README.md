# Ralph

Turn a natural-language goal into a `prd.json`, then implement it story-by-story via an AI coding CLI (clarify → PRD → review → implement → critical diff review).

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/tireymorris/ralph/main/scripts/install.sh | bash
ralph "describe the change you want"
```

Pin a branch or tag: `curl -fsSL .../install.sh | bash -s -- main`

Upgrade an existing install (same flow as `scripts/install.sh`):

```bash
ralph update
ralph update --check    # exit 0 if up to date, 2 if a newer commit is on the remote
ralph update --ref main # install a specific branch or tag
```

Override the git remote: `RALPH_REPO=https://github.com/you/ralph.git ralph update`

From a clone: `go install .` or `scripts/build.sh -o ralph` (embeds git version metadata)

Manual one-liner:

```bash
git clone https://github.com/tireymorris/ralph .tmp-ralph && cd .tmp-ralph && go install . && cd .. && rm -rf .tmp-ralph
```

**Requires:** Go 1.24.0+, Git, and one runner on `PATH`: `claude` (default), `opencode`, `pi`, or `cursor-agent` (Cursor).

## Usage

```bash
ralph                               # TUI (needs a terminal)
ralph "build a todo app"
ralph "build a todo app" --dry-run  # PRD only
ralph --resume                      # continue from prd.json
ralph status                        # non-interactive progress
ralph clean                         # remove all Ralph agent state in cwd
ralph version                       # build version, commit, and ref
ralph update                        # reinstall from GitHub (default branch main)
ralph update --check                # compare local commit to remote
ralph web                           # local web UI (default http://127.0.0.1:8080)
ralph web --port 3000               # web UI on another port
```

`ralph clean` removes all Ralph agent state in the current working directory: `prd.json`, its lock, `.ralph_questions.json`, orphaned `.prd.tmp.*` files, and `.ralph/` run data.

Implementation requires a **git repository** in the working directory (used for diff-based review between stories).

## Workflow

1. **Clarify** — optional questions via `.ralph_questions.json`
2. **Generate PRD** — runner writes `prd.json`
3. **Review PRD** — approve to implement, or revise with critique
4. **Implement** — one runner session per story; Ralph marks `passes: true` after each
5. **Implementation review** — after each story, a critical diff review runs on changed files. If findings are reported, Ralph pauses until you continue (TUI: **Enter**; web: **Continue implementation**). The same unchanged diff is not re-reviewed (duplicate fingerprint).
6. **Cleanup** — optional final pass over the branch diff (skip with `--skip-cleanup`)

`ralph --resume` continues from `prd.json` and any saved checkpoint under `.ralph/runs/prd-local/meta.json` when using the TUI. The web UI can **Force resume** stuck runs.

| Flag / env | Purpose |
|------------|---------|
| `--dry-run` | Generate PRD only |
| `--resume` | Resume from existing `prd.json` (and checkpoint if present) |
| `--skip-cleanup` | Skip post-implementation cleanup |
| `--yolo` / `RALPH_YOLO=1` | Skip manual clarify and PRD approval gates; the agent self-reviews the PRD instead |
| `--port PORT` | Web server port (with `ralph web`; default 8080) |
| `--ref REF` | Branch or tag for `ralph update` (default `main`) |
| `--check` | With `ralph update`: report whether a newer commit exists on the remote |
| `RALPH_REPO` | Git URL for `ralph update` (default `https://github.com/tireymorris/ralph.git`) |
| `RALPH_RUNNER` | Runner binary: `claude`, `opencode`, `pi`, or `cursor` |
| `RALPH_RUNNER_TIMEOUT` | Per-runner-session timeout as a Go duration, e.g. `30m` (default disabled) |
| `-v`, `--verbose` | Debug logging |
| `-h`, `--help` | Help |

## State files

Ralph writes the following files in the working directory. All are covered by the repo `.gitignore`. Run `ralph clean` to delete these artifacts idempotently (safe to run when none exist).

Starting a new run (TUI with a prompt and without `--resume`, or `POST /api/runs`) automatically moves any prior `prd.json`, its lock, `.ralph_questions.json`, orphan `.prd.tmp.*` files, and `.ralph/runs/` data into `.ralph/backups/<timestamp>/` under the workdir. Older backup folders are kept. `--resume` and checkpoint resume do not archive existing state.

| Path | Created by | Purpose |
|------|-----------|---------|
| `prd.json` | TUI + web | The generated PRD |
| `prd.json.lock` | TUI + web | File lock for concurrent PRD access |
| `.ralph_questions.json` | Runner | Temporary clarification questions (deleted after read) |
| `.ralph_prd_review.json` | Runner | PRD self-review verdict in `--yolo` runs (deleted after read) |
| `.prd.tmp.*` | TUI + web | Atomic-save temp files (orphans removed by `ralph clean`) |
| `.ralph/runs/<id>/meta.json` | TUI + web | Per-run metadata (status, checkpoint, review loop) |
| `.ralph/runs/<id>/events.ndjson` | Web UI | Per-run event log for SSE replay |
| `.ralph/runs/<id>/review-*.txt` | TUI + web | Implementation review transcripts |
| `.ralph/backups/<timestamp>/` | TUI + web | Prior state moved aside before a new run (not touched by `--resume`) |

Backends: [Claude Code](https://github.com/anthropics/claude-code), [OpenCode](https://github.com/opencode-ai/opencode), [pi](https://pi.dev), Cursor Agent.

## Development

Release-style build with git metadata embedded in the binary:

```bash
scripts/build.sh -o "$(go env GOPATH)/bin/ralph"
```

```bash
go test ./...                 # Go unit + integration tests
cd web && npm test            # React/Vitest frontend tests
cd e2e && npx playwright test # Playwright E2E tests (builds Go + frontend first)
```

When you change the web UI (`web/`), rebuild the embedded assets:

```bash
go generate ./internal/web/...
```

CI runs all three test suites on push and PR via GitHub Actions (`.github/workflows/ci.yml`).
