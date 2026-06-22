package workflow

import (
	"os/exec"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/prd/prdtest"
	"ralph/internal/shared/testgit"
)

type inMemoryPRDStore struct {
	p *prd.PRD
}

func (s inMemoryPRDStore) Load(cfg *config.Config) (*prd.PRD, error) { return s.p, nil }
func (s inMemoryPRDStore) Save(cfg *config.Config, p *prd.PRD) error { return nil }
func (s inMemoryPRDStore) Exists(cfg *config.Config) (bool, error)   { return true, nil }

func saveSingleStoryPRD(t *testing.T, skipCleanup bool) (*config.Config, *prd.PRD) {
	t.Helper()
	tmpDir := t.TempDir()
	testgit.InitRepo(t, tmpDir)
	return saveSingleStoryPRDInDir(t, tmpDir, skipCleanup)
}

func saveSingleStoryPRDInDir(t *testing.T, workDir string, skipCleanup bool) (*config.Config, *prd.PRD) {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = skipCleanup
	testPRD := prdtest.SingleStoryPRD("AC")
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}
	commitPRDFile(t, workDir, cfg.PRDFile)
	return cfg, testPRD
}

func commitPRDFile(t *testing.T, dir, prdFile string) {
	t.Helper()
	testgit.CommitFile(t, dir, prdFile, "add prd")
}

func setupCleanupBranchWithUpstreamDiff(t *testing.T) string {
	t.Helper()

	workDir := t.TempDir()
	testgit.InitRepo(t, workDir)

	runGit := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	remoteDir := t.TempDir()
	runGit(remoteDir, "init", "--bare")
	runGit(workDir, "remote", "add", "origin", remoteDir)
	runGit(workDir, "push", "-u", "origin", "main")

	testgit.WriteFile(t, workDir, "existing-change.txt", "existing change\n")
	testgit.CommitFile(t, workDir, "existing-change.txt", "existing change")

	return workDir
}

func TestSaveSingleStoryPRDSeedsSlices(t *testing.T) {
	cfg, saved := saveSingleStoryPRD(t, false)

	if len(saved.Stories) != 1 {
		t.Fatalf("Stories len = %d, want 1", len(saved.Stories))
	}
	slices := saved.Stories[0].Slices
	if len(slices) != 1 {
		t.Fatalf("Slices len = %d, want 1", len(slices))
	}
	if slices[0].ID != "slice-1" || slices[0].Behavior != "AC" || slices[0].RedHint == "" {
		t.Fatalf("slice = %+v, want id slice-1 behavior AC with non-empty red_hint", slices[0])
	}

	loaded, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Stories[0].Slices) != 1 {
		t.Fatalf("loaded Slices len = %d, want 1", len(loaded.Stories[0].Slices))
	}
}
