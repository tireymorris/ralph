package status

import (
	"bytes"
	"os"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
)

func TestDisplay(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		PRDFile:       "test_prd.json",
		RetryAttempts: 3,
		WorkDir:       tmpDir,
	}

	t.Run("no PRD file", func(t *testing.T) {
		var buf bytes.Buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		err := Display(cfg)
		w.Close()
		buf.ReadFrom(r)

		if err != nil {
			t.Errorf("Display() returned error: %v", err)
		}

		output := buf.String()
		expected := "No PRD file found. Run ralph with a prompt to create one.\n"
		if output != expected {
			t.Errorf("Expected %q, got %q", expected, output)
		}
	})

	t.Run("valid PRD file", func(t *testing.T) {
		// Create a test PRD
		testPRD := &prd.PRD{
			ProjectName: "Test Project",
			BranchName:  "main",
			Stories: []*prd.Story{
				{
					ID:       "story-1",
					Title:    "Completed story",
					Priority: 1,
					Passes:   true,
				},
				{
					ID:         "story-2",
					Title:      "Pending story",
					Priority:   2,
					Passes:     false,
					RetryCount: 0,
				},
				{
					ID:         "story-3",
					Title:      "Failed story",
					Priority:   3,
					Passes:     false,
					RetryCount: 3,
				},
			},
		}

		err := prd.Save(cfg, testPRD)
		if err != nil {
			t.Fatalf("Failed to save test PRD: %v", err)
		}

		var buf bytes.Buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		err = Display(cfg)
		w.Close()
		buf.ReadFrom(r)

		if err != nil {
			t.Errorf("Display() returned error: %v", err)
		}

		output := buf.String()

		// Check project info line
		expectedProject := "Project: Test Project (Branch: main)\n"
		if output != expectedProject && len(output) > len(expectedProject) &&
			output[:len(expectedProject)] != expectedProject {
			t.Errorf("Expected project line %q, got %q", expectedProject, output[:len(expectedProject)])
		}

		// Check story counts line
		expectedCounts := "Stories: 3 total, 1 completed, 1 pending, 1 failed\n"
		if !containsString(output, expectedCounts) {
			t.Errorf("Expected counts line %q in output %q", expectedCounts, output)
		}

		// Check story lines
		expectedStories := []string{
			"✓ [story-1] Completed story (priority: 1)",
			"⏳ [story-2] Pending story (priority: 2)",
			"✗ [story-3] Failed story (priority: 3)",
		}

		for _, expectedStory := range expectedStories {
			if !containsString(output, expectedStory) {
				t.Errorf("Expected story line %q in output %q", expectedStory, output)
			}
		}
	})

	t.Run("PRD without branch name", func(t *testing.T) {
		testPRD := &prd.PRD{
			ProjectName: "Simple Project",
			Stories: []*prd.Story{
				{
					ID:       "story-1",
					Title:    "Only story",
					Priority: 1,
					Passes:   true,
				},
			},
		}

		err := prd.Save(cfg, testPRD)
		if err != nil {
			t.Fatalf("Failed to save test PRD: %v", err)
		}

		var buf bytes.Buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		err = Display(cfg)
		w.Close()
		buf.ReadFrom(r)

		if err != nil {
			t.Errorf("Display() returned error: %v", err)
		}

		output := buf.String()
		expectedProject := "Project: Simple Project\n"
		if output != expectedProject && len(output) > len(expectedProject) &&
			output[:len(expectedProject)] != expectedProject {
			t.Errorf("Expected project line %q, got %q", expectedProject, output[:len(expectedProject)])
		}
	})
}

func TestDisplay_EmptyStories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		PRDFile:       "empty_prd.json",
		RetryAttempts: 3,
		WorkDir:       tmpDir,
	}

	// Create a PRD with empty stories array
	testPRD := &prd.PRD{
		ProjectName: "Empty Project",
		Stories:     []*prd.Story{},
	}

	err := prd.Save(cfg, testPRD)
	if err != nil {
		t.Fatalf("Failed to save test PRD: %v", err)
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	err = Display(cfg)
	w.Close()
	buf.ReadFrom(r)

	if err != nil {
		t.Errorf("Display() returned error: %v", err)
	}

	output := buf.String()
	expectedCounts := "Stories: 0 total, 0 completed, 0 pending, 0 failed\n"
	if !containsString(output, expectedCounts) {
		t.Errorf("Expected counts line %q in output %q", expectedCounts, output)
	}
}

func TestDisplay_AllCompletedStories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		PRDFile:       "all_completed_prd.json",
		RetryAttempts: 3,
		WorkDir:       tmpDir,
	}

	// Create a PRD with all completed stories
	testPRD := &prd.PRD{
		ProjectName: "All Completed Project",
		Stories: []*prd.Story{
			{
				ID:       "story-1",
				Title:    "First completed",
				Priority: 1,
				Passes:   true,
			},
			{
				ID:       "story-2",
				Title:    "Second completed",
				Priority: 2,
				Passes:   true,
			},
		},
	}

	err := prd.Save(cfg, testPRD)
	if err != nil {
		t.Fatalf("Failed to save test PRD: %v", err)
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	err = Display(cfg)
	w.Close()
	buf.ReadFrom(r)

	if err != nil {
		t.Errorf("Display() returned error: %v", err)
	}

	output := buf.String()
	expectedCounts := "Stories: 2 total, 2 completed, 0 pending, 0 failed\n"
	if !containsString(output, expectedCounts) {
		t.Errorf("Expected counts line %q in output %q", expectedCounts, output)
	}
}

func TestDisplay_AllFailedStories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		PRDFile:       "all_failed_prd.json",
		RetryAttempts: 2,
		WorkDir:       tmpDir,
	}

	// Create a PRD with all failed stories
	testPRD := &prd.PRD{
		ProjectName: "All Failed Project",
		Stories: []*prd.Story{
			{
				ID:         "story-1",
				Title:      "First failed",
				Priority:   1,
				Passes:     false,
				RetryCount: 2,
			},
			{
				ID:         "story-2",
				Title:      "Second failed",
				Priority:   2,
				Passes:     false,
				RetryCount: 3,
			},
		},
	}

	err := prd.Save(cfg, testPRD)
	if err != nil {
		t.Fatalf("Failed to save test PRD: %v", err)
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	err = Display(cfg)
	w.Close()
	buf.ReadFrom(r)

	if err != nil {
		t.Errorf("Display() returned error: %v", err)
	}

	output := buf.String()
	expectedCounts := "Stories: 2 total, 0 completed, 0 pending, 2 failed\n"
	if !containsString(output, expectedCounts) {
		t.Errorf("Expected counts line %q in output %q", expectedCounts, output)
	}
}

func TestDisplayWithCorruptedPRD(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		PRDFile:       "corrupted_prd.json",
		RetryAttempts: 3,
		WorkDir:       tmpDir,
	}

	// Write corrupted JSON
	err := os.WriteFile(cfg.PRDPath(), []byte("{ invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted PRD: %v", err)
	}

	err = Display(cfg)
	if err == nil {
		t.Error("Display() should have returned error for corrupted PRD")
	}

	expectedError := "failed to load PRD"
	if err.Error()[:len(expectedError)] != expectedError {
		t.Errorf("Expected error starting with %q, got %q", expectedError, err.Error())
	}
}

func containsLine(s, substr string) bool {
	lines := []string{s}
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[i+1:])
		}
	}
	for _, line := range lines {
		if line == substr {
			return true
		}
	}
	return false
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || s[len(s)-len(substr):] == substr ||
			containsInMiddle(s, substr))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
