package main

import (
	"os"
	"strings"
	"testing"
)

func TestCIWorkflowInjectsVersionMetadata(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/ci.yml")
	if err != nil {
		t.Fatalf("read ci.yml: %v", err)
	}
	content := string(data)
	hasLdflags := strings.Contains(content, "ldflags")
	usesBuildScript := strings.Contains(content, "scripts/build.sh")
	if !hasLdflags && !usesBuildScript {
		t.Fatal("ci.yml must pass version ldflags or invoke scripts/build.sh when building Go binaries")
	}
}
