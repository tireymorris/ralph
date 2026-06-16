package workflow

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

type inMemoryPRDStore struct {
	p *prd.PRD
}

func (s inMemoryPRDStore) Load(cfg *config.Config) (*prd.PRD, error) { return s.p, nil }
func (s inMemoryPRDStore) Save(cfg *config.Config, p *prd.PRD) error { return nil }
func (s inMemoryPRDStore) Exists(cfg *config.Config) (bool, error)   { return true, nil }

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
	initGitRepoInDir(t, tmpDir)
	return saveSingleStoryPRDInDir(t, tmpDir, skipCleanup)
}

func saveSingleStoryPRDInDir(t *testing.T, workDir string, skipCleanup bool) (*config.Config, *prd.PRD) {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
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
	commitPRDFile(t, workDir, cfg.PRDFile)
	return cfg, testPRD
}

func commitPRDFile(t *testing.T, dir, prdFile string) {
	t.Helper()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	runGit("add", prdFile)
	runGit("commit", "-m", "add prd")
}

func initGitRepoInDir(t *testing.T, dir string) {
	t.Helper()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	runGit("init")
	runGit("checkout", "-b", "main")
	runGit("config", "user.name", "Test User")
	runGit("config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ok\n"), 0644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	runGit("add", "README.md")
	runGit("commit", "-m", "initial")
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

func cleanupEventIndices(evts []Event) (startedIdx, completedIdx, workflowCompletedIdx int) {
	startedIdx, completedIdx, workflowCompletedIdx = -1, -1, -1
	for i, e := range evts {
		switch e.(type) {
		case EventCleanupStarted:
			startedIdx = i
		case EventCleanupCompleted:
			completedIdx = i
		case EventCompleted:
			workflowCompletedIdx = i
		}
	}
	return startedIdx, completedIdx, workflowCompletedIdx
}

func assertCleanupBeforeCompleted(t *testing.T, evts []Event) {
	t.Helper()
	startedIdx, completedIdx, workflowCompletedIdx := cleanupEventIndices(evts)
	if startedIdx < 0 {
		t.Fatal("expected EventCleanupStarted to be emitted")
	}
	if completedIdx < 0 {
		t.Fatal("expected EventCleanupCompleted to be emitted")
	}
	if workflowCompletedIdx < 0 {
		t.Fatal("expected EventCompleted to be emitted")
	}
	if startedIdx >= workflowCompletedIdx {
		t.Error("EventCleanupStarted must come before EventCompleted")
	}
	if completedIdx >= workflowCompletedIdx {
		t.Error("EventCleanupCompleted must come before EventCompleted")
	}
	if startedIdx >= completedIdx {
		t.Error("EventCleanupStarted must come before EventCleanupCompleted")
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

func setupGitRepoWithWorkingTreeDiff(t *testing.T) (workDir, changedFile string) {
	t.Helper()
	workDir = t.TempDir()
	initGitRepoInDir(t, workDir)
	changedFile = "delta.txt"
	if err := os.WriteFile(filepath.Join(workDir, changedFile), []byte("changed\n"), 0644); err != nil {
		t.Fatalf("write %s: %v", changedFile, err)
	}
	return workDir, changedFile
}
