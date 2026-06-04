package workflow

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/prd"
)

func drainEvents(ch chan Event) []Event {
	var evts []Event
	for len(ch) > 0 {
		evts = append(evts, <-ch)
	}
	return evts
}

func saveSingleStoryPRD(t *testing.T, skipCleanup bool) (*config.Config, *prd.PRD) {
	t.Helper()
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = skipCleanup
	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{{
			ID:                 "1",
			Title:              "Story",
			Description:        "Desc",
			AcceptanceCriteria: []string{"AC"},
			Priority:           1,
			Passes:             false,
		}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}
	return cfg, testPRD
}

type cleanupEventCounts struct {
	started   int
	completed int
}

func countCleanupEvents(evts []Event) cleanupEventCounts {
	var counts cleanupEventCounts
	for _, e := range evts {
		switch e.(type) {
		case EventCleanupStarted:
			counts.started++
		case EventCleanupCompleted:
			counts.completed++
		}
	}
	return counts
}

func assertCleanupPassSequence(t *testing.T, evts []Event, wantPasses int) {
	t.Helper()
	var startedPasses, completedPasses []int
	firstStartedIdx, firstCompletedIdx := -1, -1
	for i, e := range evts {
		switch ev := e.(type) {
		case EventCleanupStarted:
			startedPasses = append(startedPasses, ev.Pass)
			if ev.Total != constants.CleanupPassCount {
				t.Errorf("EventCleanupStarted Total=%d, want %d", ev.Total, constants.CleanupPassCount)
			}
			if firstStartedIdx < 0 {
				firstStartedIdx = i
			}
		case EventCleanupCompleted:
			completedPasses = append(completedPasses, ev.Pass)
			if ev.Total != constants.CleanupPassCount {
				t.Errorf("EventCleanupCompleted Total=%d, want %d", ev.Total, constants.CleanupPassCount)
			}
			if firstCompletedIdx < 0 {
				firstCompletedIdx = i
			}
		}
	}
	if len(startedPasses) != wantPasses {
		t.Errorf("expected %d EventCleanupStarted events, got %d", wantPasses, len(startedPasses))
	}
	if len(completedPasses) != wantPasses {
		t.Errorf("expected %d EventCleanupCompleted events, got %d", wantPasses, len(completedPasses))
	}
	for i, pass := range startedPasses {
		if pass != i+1 {
			t.Errorf("EventCleanupStarted pass %d: Pass=%d, want %d", i, pass, i+1)
		}
	}
	for i, pass := range completedPasses {
		if pass != i+1 {
			t.Errorf("EventCleanupCompleted pass %d: Pass=%d, want %d", i, pass, i+1)
		}
	}
	if wantPasses > 0 && firstStartedIdx >= firstCompletedIdx {
		t.Error("first EventCleanupStarted should come before first EventCleanupCompleted")
	}
}

func cleanupEventIndices(evts []Event) (startedIdxs, completedIdxs []int, completedIdx int) {
	completedIdx = -1
	for i, e := range evts {
		switch e.(type) {
		case EventCleanupStarted:
			startedIdxs = append(startedIdxs, i)
		case EventCleanupCompleted:
			completedIdxs = append(completedIdxs, i)
		case EventCompleted:
			completedIdx = i
		}
	}
	return startedIdxs, completedIdxs, completedIdx
}

func assertCleanupBeforeCompleted(t *testing.T, evts []Event) {
	t.Helper()
	startedIdxs, completedIdxs, completedIdx := cleanupEventIndices(evts)
	if len(startedIdxs) != constants.CleanupPassCount {
		t.Fatalf("expected %d EventCleanupStarted events, got %d", constants.CleanupPassCount, len(startedIdxs))
	}
	if len(completedIdxs) != constants.CleanupPassCount {
		t.Fatalf("expected %d EventCleanupCompleted events, got %d", constants.CleanupPassCount, len(completedIdxs))
	}
	if completedIdx == -1 {
		t.Fatal("expected EventCompleted to be emitted")
	}
	for _, idx := range startedIdxs {
		if idx >= completedIdx {
			t.Error("each EventCleanupStarted must come before EventCompleted")
		}
	}
	for _, idx := range completedIdxs {
		if idx >= completedIdx {
			t.Error("each EventCleanupCompleted must come before EventCompleted")
		}
	}
	for i := 0; i < constants.CleanupPassCount; i++ {
		if startedIdxs[i] >= completedIdxs[i] {
			t.Errorf("cleanup pass %d: EventCleanupStarted must come before EventCleanupCompleted", i+1)
		}
	}
}

func setupCleanupBranchWithUpstreamDiff(t *testing.T) string {
	t.Helper()

	workDir := t.TempDir()
	runGit := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	writeFile := func(name, contents string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(workDir, name), []byte(contents), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	runGit(workDir, "init")
	runGit(workDir, "checkout", "-b", "main")
	runGit(workDir, "config", "user.name", "Test User")
	runGit(workDir, "config", "user.email", "test@example.com")
	writeFile("base.txt", "base\n")
	runGit(workDir, "add", "base.txt")
	runGit(workDir, "commit", "-m", "base commit")

	remoteDir := t.TempDir()
	runGit(remoteDir, "init", "--bare")
	runGit(workDir, "remote", "add", "origin", remoteDir)
	runGit(workDir, "push", "-u", "origin", "main")

	writeFile("existing-change.txt", "existing change\n")
	runGit(workDir, "add", "existing-change.txt")
	runGit(workDir, "commit", "-m", "existing change")

	return workDir
}
