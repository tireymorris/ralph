# Ralph agent philosophy

Ralph does not teach agents via README or skills at runtime. Each workflow phase sends a rendered template from `internal/prompt/templates/` to the coding runner on stdin.

## Flow

```
User goal → clarify → PRD → review → implement (per story) → impl review → cleanup
```

TUI and web share the same `workflow.Driver`. Templates define what the runner is told to do at each step.

## Planning agent

**Templates:** `clarify`, `prd-generate`, `prd-self-review`, `prd-critique-revision`, `prd-clarification-revision`, `followup`

**Philosophy:**
- Ask few, high-impact clarify questions (0–5); prefer `[]` when the prompt is clear enough
- Budget ~5–15 targeted file reads for PRD work; do not grep the whole repo
- Size stories for ~3–10 TDD cycles each; verifiable acceptance criteria only
- Stop after writing the PRD file — never implement during planning

## Implementation agent

**Template:** `story-implement` (+ `partials/commit-rules`)

**Philosophy:**
- TDD in small "slices": red → green → (refactor) → commit
- Many small commits per story; no conventional-commit prefixes; no commit trailers
- Update `prd.json` `context` only when a new pattern must be recorded for later stories

## Review agent

**Template:** `diff-review`

**Philosophy:**
- Review changed files against the PRD; emit `===ralph-findings===` JSON
- Uncommitted work in listed files still counts as delivered

## Recovery agent

**Template:** `recovery`

**Philosophy:**
- Fix findings or story failures directly in the repo
- Commit every fix before stopping; do not mark PRD stories complete

## Cleanup agent

**Template:** `cleanup`

**Philosophy:**
- Skip if nothing worth refactoring; otherwise SOLID/DRY pass on changed files only
- Run targeted tests, not the full suite unless the repo is small

## Where to edit

Change template files under [`internal/prompt/templates/`](../../internal/prompt/templates/). See [`docs/prompts/README.md`](README.md) for the full index.
