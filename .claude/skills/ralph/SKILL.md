---
name: ralph
description: >-
  Ralph CLI (tireymorris/ralph): quick-start guidance for using Ralph in a repo.
  Use when the user mentions ralph, prd.json, RALPH_RUNNER, or wants to run or
  debug Ralph.
---

# Ralph

Ralph turns a natural-language goal into `prd.json`, then implements it story by story via an AI runner.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/tireymorris/ralph/main/scripts/install.sh | bash
```

## Use

Run from the target repo root (must be a **git** repo for implementation):

```bash
ralph                 # TUI
ralph "..."          # TUI flow
ralph "..." --dry-run # PRD only
ralph "..." --yolo   # non-interactive run; skip clarify, agent self-reviews the PRD instead of manual approval
ralph --resume        # continue from prd.json (+ checkpoint if saved)
ralph status          # current PRD status
ralph web             # local UI
```

Set `RALPH_RUNNER` to `claude`, `cursor`, `opencode`, or `pi`.
Use `RALPH_YOLO=1` for non-interactive auto-approve runs.
Use `RALPH_RUNNER_TIMEOUT=30m` (Go duration) to cap each runner session.

## How it works

1. Clarify — runner may write `.ralph/questions.json`; Ralph reads and removes it
2. Generate/load PRD — runner writes `prd.json`
3. Review PRD — via TUI or `ralph web` (approve or revise)
4. Implement — one runner session per ready story; Ralph marks `passes: true` when the runner exits 0
5. Implementation review — after each story, critical diff review on changed files; may pause on findings (TUI: Enter to continue; web: Continue implementation)
6. Cleanup — optional final diff pass (skip with `--skip-cleanup`)

Ralph is an orchestrator; it does not write code.

## Key files

Gitignore these in the target repo:

- `prd.json`
- `prd.json.lock`
- `.ralph/` (questions, self-review verdict, `prd.tmp.*`, run metadata, events, review transcripts; TUI uses `runs/prd-local/`)

## Caveats

- `passes: true` is not proof tests passed
- `ralph status` is progress, not QA sign-off
- large PRD runs can overscope badly
- `--dry-run` may still need a real TTY in some environments
- Ralph does not load `CLAUDE.md` unless the runner does
- implementation review requires git and a runner that emits `===ralph-findings===` JSON in its transcript
