package workflow

import (
	"testing"

	"ralph/internal/shared/runstate"
)

func TestReviewLoopUpdateIsRunstateType(t *testing.T) {
	var _ runstate.ReviewLoopUpdate = ReviewLoopUpdate{}
}
