# Prompt sources

Ralph prompts are embedded at compile time. They are **language-agnostic**. Stack-specific detail should land in PRD `context` during planning, not in per-framework template forks.

When you refine personal skills under `~/.agents/skills`, update the corresponding partial here.

| Partial / template | Primary skill sources | What was distilled |
|--------------------|----------------------|-------------------|
| `partials/commit-rules.tmpl` | commit, test-driven-development | Red-green-refactor, watch-the-fail, delete early code, observable tests, commit message rules |
| `partials/working-conventions.tmpl` | style-guide (process + diff hygiene) | Read nearby files, additive diffs, no drive-by refactors, scan diff before commit |
| `partials/review-conventions.tmpl` | code-review-excellence, style-guide (testing antipatterns) | Review focus, scope creep, actionable findings, category hints |
| `partials/refactor-discipline.tmpl` | refactor, style-guide (cross-cutting) | Behavior preserved, changed-files scope, no compat shims without rollout reason |
| `prd-generate.tmpl` (context bullets) | style-guide, ralph | Record observed conventions in `context` for later agents |
| `recovery.tmpl` | working-conventions + recovery flow | Focused fixes only |

Skills that stay **outside** runner prompts (Cursor/human workflow): `pr-description`, `gh-cli`, `TODO`, `ralph` (CLI docs), `demo-recorder`, `cleanup-local-repo`, `honeybadger-cli`, `golang-patterns` (framework-specific — capture in `context` when relevant).
