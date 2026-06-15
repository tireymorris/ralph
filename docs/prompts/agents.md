# Ralph agent philosophy

Ralph sends rendered templates from `internal/prompt/templates/` to the coding runner on stdin. Prompt prose is **language-agnostic** — stack-specific conventions belong in the PRD `context` field (filled during planning from the repo), not in framework-specific template branches.

Cross-cutting discipline (TDD, diff hygiene, review focus, refactor rules) is distilled into shared partials. When you update personal agent skills (`~/.agents/skills`), refresh the matching partials here — see [`skill-sources.md`](skill-sources.md).

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
- Size stories for ~1–10 slices each; verifiable behaviors only
- Capture observed local conventions in PRD `context` for later phases
- Stop after writing the PRD file — never implement during planning

## Implementation agent

**Templates:** `story-implement`, `partials/working-conventions`, `partials/commit-rules`

**Philosophy:**
- TDD slices: red → green → mandatory refactor → commit
- Read nearby files; focused additive diffs; no drive-by refactors
- Many small commits; no conventional-commit prefixes or trailers
- Tests assert observable behavior, not implementation details
- Update `prd.json` `context` only when a new pattern must be recorded for later stories

## Review agent

**Template:** `diff-review`, `partials/review-conventions`

**Philosophy:**
- Review logic, edge cases, tests, and scope — not formatting
- Emit `===ralph-findings===` JSON with specific, actionable findings
- Uncommitted work in listed files still counts as delivered

## Recovery agent

**Template:** `recovery`, `partials/working-conventions`

**Philosophy:**
- Fix findings or story failures with focused diffs
- Commit every fix before stopping; do not mark PRD stories complete

## Cleanup agent

**Templates:** `cleanup`, `partials/refactor-discipline`, `partials/working-conventions`

**Philosophy:**
- Skip if nothing worth doing; otherwise refactor changed files only
- Preserve behavior; do not mix cleanup with new features
- Run targeted tests, not the full suite unless the repo is small

## Where to edit

Change template files under [`internal/prompt/templates/`](../../internal/prompt/templates/). See [`README.md`](README.md) for the full index.
