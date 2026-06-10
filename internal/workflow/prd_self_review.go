package workflow

import (
	"context"
	"fmt"
	"os"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
)

// runPRDSelfReview has the agent critique and revise the PRD against the
// rubric in prompt.PRDSelfReview, looping until it approves or rounds run out.
func (e *Executor) runPRDSelfReview(ctx context.Context, userPrompt string) (*prd.PRD, error) {
	maxRounds := constants.MaxPRDSelfReviewRounds

	var p *prd.PRD
	for round := 1; round <= maxRounds; round++ {
		e.emit(EventOutput{Output: Output{Text: fmt.Sprintf("PRD self-review round %d of %d", round, maxRounds)}})

		reviewPrompt := prompt.PRDSelfReview(userPrompt, e.cfg.PRDFile, round, maxRounds)
		if err := e.runWithForwardedOutput(ctx, reviewPrompt); err != nil {
			return nil, fmt.Errorf("PRD self-review round %d failed with runner %s: %w", round, e.cfg.Runner, err)
		}

		reloaded, err := e.store.Load(e.cfg)
		if err != nil {
			return nil, fmt.Errorf("loading PRD after self-review round %d: %w", round, err)
		}
		p = reloaded

		verdictReader := PRDReviewVerdictReader{WorkDir: e.cfg.WorkDir}
		if _, statErr := os.Stat(verdictReader.Path()); os.IsNotExist(statErr) {
			logger.Warn("PRD self-review verdict file missing, treating as approved", "round", round, "file", verdictReader.Path())
			break
		}

		verdict, err := verdictReader.ReadRemove()
		if err != nil {
			return nil, err
		}
		if verdict.Summary != "" {
			e.emit(EventOutput{Output: Output{Text: "Self-review verdict: " + verdict.Summary}})
		}
		if verdict.Approved {
			break
		}
	}
	return p, nil
}
