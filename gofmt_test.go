package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestGoSourcesGofmtClean(t *testing.T) {
	t.Helper()
	out, err := exec.Command("gofmt", "-l", ".").Output()
	if err != nil {
		t.Fatalf("gofmt -l .: %v", err)
	}
	if unformatted := strings.TrimSpace(string(out)); unformatted != "" {
		t.Fatalf("gofmt -l . listed unformatted files:\n%s", unformatted)
	}
}
