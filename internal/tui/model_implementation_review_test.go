package tui

import (
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/workflow/events"
)

func TestHandleWorkflowEventImplementationReviewLogs(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.logger.SetSize(80, 20)

	m.handleWorkflowEvent(events.EventImplementationReviewStarted{Iteration: 2})
	m.handleWorkflowEvent(events.EventImplementationReview{
		Findings: []events.ImplementationFinding{
			{ID: "finding-1", Summary: "missing tests"},
			{ID: "finding-2", Summary: "unsafe cast"},
		},
	})
	m.handleWorkflowEvent(events.EventImplementationReviewCompleted{Iteration: 2, Clean: false})

	logView := m.logger.GetView().View()
	for _, want := range []string{
		"Cleanup review started (iteration 2)",
		"Review finding: missing tests",
		"Review finding: unsafe cast",
		"Cleanup review completed (iteration 2, findings)",
	} {
		if !strings.Contains(logView, want) {
			t.Fatalf("log view missing %q, got:\n%s", want, logView)
		}
	}
}
