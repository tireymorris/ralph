package workflow

import (
	"testing"

	"github.com/tireymorris/ralph/internal/shared/runstate"
)

func TestReviewLoopUpdateIsRunstateType(t *testing.T) {
	var _ runstate.ReviewLoopUpdate = ReviewLoopUpdate{}
}
