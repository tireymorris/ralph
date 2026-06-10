package workflow

import (
	"context"
	"fmt"

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
	approved := false
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
		verdict, status, err := verdictReader.readAndRemove()
		if err != nil {
			return nil, err
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
	return p, nil
}
