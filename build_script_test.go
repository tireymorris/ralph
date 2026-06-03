package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildScriptInjectsCommit(t *testing.T) {
	t.Parallel()

	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	buildScript := filepath.Join(repoRoot, "scripts", "build.sh")
	if _, err := os.Stat(buildScript); err != nil {
		t.Fatalf("scripts/build.sh: %v", err)
	}

	headCmd := exec.Command("git", "rev-parse", "HEAD")
	headCmd.Dir = repoRoot
	headOut, err := headCmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	wantCommit := strings.TrimSpace(string(headOut))
	if len(wantCommit) != 40 {
		t.Fatalf("HEAD = %q, want 40-char hex SHA", wantCommit)
	}

	outPath := filepath.Join(t.TempDir(), "ralph-built")
	buildCmd := exec.Command(buildScript, "-o", outPath)
	buildCmd.Dir = repoRoot
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build.sh: %v\n%s", err, buildOut)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read binary: %v", err)
	}
	if !bytes.Contains(data, []byte(wantCommit)) {
		t.Fatalf("binary does not embed commit %s", wantCommit)
	}
}
