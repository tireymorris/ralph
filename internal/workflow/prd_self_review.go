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
// Round failures degrade to the current on-disk PRD rather than failing the run.
func (e *Executor) runPRDSelfReview(ctx context.Context, userPrompt string) (*prd.PRD, error) {
	maxRounds := constants.MaxPRDSelfReviewRounds
	verdictReader := PRDReviewVerdictReader{WorkDir: e.cfg.WorkDir}
	if err := os.Remove(verdictReader.Path()); err != nil && !os.IsNotExist(err) {
		logger.Warn("failed to remove stale PRD self-review verdict", "error", err, "file", verdictReader.Path())
	}

	approved := false
	for round := 1; round <= maxRounds; round++ {
		e.emit(EventOutput{Output: Output{Text: fmt.Sprintf("PRD self-review round %d of %d", round, maxRounds)}})

		reviewPrompt := prompt.PRDSelfReview(userPrompt, e.cfg.PRDFile, round, maxRounds)
		if err := e.runWithForwardedOutput(ctx, reviewPrompt); err != nil {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("PRD self-review round %d: %w", round, err)
			}
			logger.Warn("PRD self-review round failed, proceeding with current PRD", "round", round, "error", err)
			e.emit(EventOutput{Output: Output{Text: fmt.Sprintf("Self-review round %d failed, proceeding with current PRD", round)}})
			break
		}

		verdict, status, err := verdictReader.readAndRemove()
		if err != nil {
			logger.Warn("failed to read PRD self-review verdict, proceeding with current PRD", "round", round, "error", err)
			e.emit(EventOutput{Output: Output{Text: fmt.Sprintf("Self-review round %d: could not read verdict, proceeding with current PRD", round)}})
			break
		}
		switch status {
		case verdictMissing:
			logger.Warn("PRD self-review verdict file missing, counting round as not approved", "round", round, "file", verdictReader.Path())
			e.emit(EventOutput{Output: Output{Text: fmt.Sprintf("Self-review round %d: verdict file missing, retrying", round)}})
			continue
		case verdictMalformed:
			logger.Warn("PRD self-review verdict malformed, counting round as not approved", "round", round, "file", verdictReader.Path())
			e.emit(EventOutput{Output: Output{Text: fmt.Sprintf("Self-review round %d: malformed verdict, retrying", round)}})
			continue
		}
		if verdict.Summary != "" {
			e.emit(EventOutput{Output: Output{Text: "Self-review verdict: " + verdict.Summary}})
		}
		if verdict.Approved {
			approved = true
			break
		}
	}
	if !approved {
		e.emit(EventOutput{Output: Output{Text: fmt.Sprintf("PRD self-review did not approve within %d rounds; proceeding with last PRD revision", maxRounds)}})
	}

	p, err := e.store.Load(e.cfg)
	if err != nil {
		return nil, fmt.Errorf("loading PRD after self-review: %w", err)
	}
	return p, nil
}
