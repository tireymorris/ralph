package workflow

import (
	"context"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

// blockedPRDStore bypasses prd.Load validation so tests can return a PRD
// where every incomplete story has an unsatisfied dependency.
type blockedPRDStore struct {
	p *prd.PRD
}

func (s blockedPRDStore) Load(cfg *config.Config) (*prd.PRD, error) { return s.p, nil }
func (s blockedPRDStore) Save(cfg *config.Config, p *prd.PRD) error { return nil }
func (s blockedPRDStore) Exists(cfg *config.Config) (bool, error)   { return true, nil }

func newAllBlockedPRD() *prd.PRD {
	return &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{
			{ID: "story-a", Title: "A", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, DependsOn: []string{"story-missing"}},
		},
	}
}

func TestRunImplementationAllStoriesBlockedReturnsError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	blocked := newAllBlockedPRD()
	ch := make(chan Event, 100)
	exec := NewExecutorWithRunnerAndStore(cfg, ch, newMockRunner(), blockedPRDStore{p: blocked})

	done := make(chan error, 1)
	go func() { done <- exec.RunImplementation(context.Background(), blocked) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("RunImplementation() should return error when all incomplete stories are blocked")
		}
	case <-time.After(time.Second):
		t.Fatal("RunImplementation() did not return within 1s; busy loop on blocked stories")
	}
}

func TestRunImplementationBlockedErrorNamesStoriesAndDependencies(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	blocked := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{
			{ID: "story-a", Title: "A", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, DependsOn: []string{"story-missing"}},
			{ID: "story-b", Title: "B", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 2, DependsOn: []string{"story-a"}},
		},
	}
	ch := make(chan Event, 100)
	exec := NewExecutorWithRunnerAndStore(cfg, ch, newMockRunner(), blockedPRDStore{p: blocked})

	err := exec.RunImplementation(context.Background(), blocked)
	if err == nil {
		t.Fatal("RunImplementation() should return error when all incomplete stories are blocked")
	}

	msg := err.Error()
	for _, want := range []string{"story-a", "story-missing", "story-b"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q should mention %q", msg, want)
		}
	}
}
