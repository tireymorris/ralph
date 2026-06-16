# Ralph agent prompts

Editable prompt templates live in [`internal/prompt/templates/`](../../internal/prompt/templates/). They are embedded into the Ralph binary at compile time via `go:embed` and rendered with `text/template`.

Each rendered prompt is prefixed with a kind marker:

```text
===ralph-prompt-kind:story-implement===
You are Ralph's implementation agent...
```

Use `prompt.Kind()`, `prompt.HasKind()`, and the `Kind*` constants in Go — do not match prompt prose in production code or mocks. Kind names match template define names (`clarify`, `prd-generate`, `story-implement`, `diff-review`, `recovery`, etc.).

## Edit workflow

1. Change the relevant `.tmpl` file under `internal/prompt/templates/`.
2. Run `go test ./internal/prompt/...` to verify behavior.
3. Run `go test ./...` before committing.

Do not edit prompt prose in Go files — `internal/prompt/prompt.go` and siblings only assemble data structs and call `render()`.

## Agent personas

| Template | Persona | Workflow phase |
|----------|---------|----------------|
| `clarify.tmpl` | planning agent | Clarification |
| `prd-generate.tmpl` | planning agent | PRD generation |
| `prd-self-review.tmpl` | planning agent | PRD self-review (`--yolo`) |
| `prd-critique-revision.tmpl` | planning agent | User PRD critique |
| `prd-clarification-revision.tmpl` | planning agent | Post-revision clarify |
| `story-implement.tmpl` | implementation agent | Story implementation |
| `diff-review.tmpl` | critical diff review agent | Implementation review |
| `recovery.tmpl` | recovery agent | Story/review recovery |
| `cleanup.tmpl` | cleanup agent | Post-implementation cleanup |
| `followup.tmpl` | planning agent | Follow-up PRD revision |

## Partials

| Partial | Used by |
|---------|---------|
| `partials/codebase-context.tmpl` | story-implement, diff-review, recovery, cleanup |
| `partials/changed-files.tmpl` | diff-review, cleanup |
| `partials/clarifications.tmpl` | prd-generate, prd-clarification-revision |
| `partials/working-conventions.tmpl` | story-implement, recovery, cleanup |
| `partials/commit-rules.tmpl` | story-implement (TDD slices + commit rules) |
| `partials/review-conventions.tmpl` | diff-review |
| `partials/refactor-discipline.tmpl` | cleanup |

See [`skill-sources.md`](skill-sources.md) for how these map to `~/.agents/skills`.

## Template variables

### `clarify.tmpl`

- `CodebaseNote` — "an existing codebase" or "a new project (no existing source code)"
- `UserPrompt`
- `QuestionsFile`

### `prd-generate.tmpl`

- `UserPrompt`, `PRDFile`, `BranchPrefix`, `ContextGuidance`
- `Clarifications` — `[]QuestionAnswer` (optional)

### `story-implement.tmpl`

- `StoryID`, `Title`, `Description`, `AcceptanceCriteria` (joined string)
- `FeatureTestSpec`, `Context`, `PRDFile`
- `Completed`, `Total`, `DependsOn`

### `recovery.tmpl`

- `AgentMarker` — must match `RecoveryAgentMarker` in Go
- `Context`, `ErrorMessage`, `FindingsJSON`, `Escalate`
- `Reason`, `Attempt`, `MaxAttempts`, `PRDFile`

See [`types.go`](../../internal/prompt/types.go) for all data structs.

## Tests

- Substring tests in `internal/prompt/*_test.go` assert template **content** (behavior users/agents see).
- Kind-marker tests in `kind_test.go` and workflow `prompt_kinds_test.go` assert routing uses `HasKind`, not prose.
- `embed_test.go` verifies templates parse and render without panicking.
