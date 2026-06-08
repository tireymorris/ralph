package main

import (
	"os"
	"strings"
	"testing"
)

func TestInternalReadmeDocumentsCodeOrganizationBoundaries(t *testing.T) {
	data, err := os.ReadFile("internal/README.md")
	if err != nil {
		t.Fatalf("read internal/README.md: %v", err)
	}
	content := string(data)

	for _, want := range []string{
		"internal/workflow",
		"internal/shared",
		"internal/tui",
		"internal/web",
		"web/src",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("internal/README.md must document %s", want)
		}
	}

	for _, want := range []string{
		"go test ./... && cd web && npm test && cd ../e2e && npx playwright test",
		"go generate ./internal/web/...",
		"internal/shared/session",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("internal/README.md must include %q", want)
		}
	}
}
