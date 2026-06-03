package main

import (
	"os"
	"strings"
	"testing"
)

func TestReadmeDocumentsReleaseBuild(t *testing.T) {
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "scripts/build.sh") {
		t.Fatal("README.md must document scripts/build.sh for release-style builds")
	}
}
