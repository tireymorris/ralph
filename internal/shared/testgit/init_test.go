package testgit

import (
	"os/exec"
	"strings"
	"testing"
)

func TestInitRepoCreatesMainBranchWithInitialCommit(t *testing.T) {
	dir := t.TempDir()
	InitRepo(t, dir)

	out, err := exec.Command("git", "-C", dir, "branch", "--show-current").CombinedOutput()
	if err != nil {
		t.Fatalf("git branch --show-current: %v\n%s", err, out)
	}
	if branch := strings.TrimSpace(string(out)); branch != "main" {
		t.Fatalf("branch = %q, want main", branch)
	}

	out, err = exec.Command("git", "-C", dir, "log", "-1", "--oneline").CombinedOutput()
	if err != nil {
		t.Fatalf("git log: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "initial") {
		t.Fatalf("log = %q, want commit message containing initial", out)
	}
}
