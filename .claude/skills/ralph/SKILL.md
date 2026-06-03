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

Run from the target repo root:

```bash
ralph                 # TUI
ralph "..."          # TUI flow
ralph "..." --dry-run # PRD only
ralph --resume        # continue from existing prd.json
ralph status          # current PRD status
ralph web             # local UI
```

Set `RALPH_RUNNER` to `claude`, `cursor`, `opencode`, or `pi`.

## How it works

1. Clarify — runner may write `.ralph_questions.json`; Ralph reads and removes it
2. Generate/load PRD — runner writes `prd.json`
3. Review PRD — via TUI or `ralph web`
4. Implement — Ralph spawns one runner session per ready story and marks `passes: true` when the runner exits 0

Ralph is an orchestrator; it does not write code.

## Key files

Gitignore these in the target repo:

- `prd.json`
- `prd.json.lock`
- `.ralph_questions.json`
- `.ralph/` (web UI session history)
- `.prd.tmp.*`

## Caveats

- `passes: true` is not proof tests passed
- `ralph status` is progress, not QA sign-off
- large PRD runs can overscope badly
- `--dry-run` may still need a real TTY in some environments
- Ralph does not load `CLAUDE.md` unless the runner does
