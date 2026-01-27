package status

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
)

// captureStdout runs fn and returns whatever it printed to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() failed: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	fn()
	w.Close()

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestDisplay(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		PRDFile:       "test_prd.json",
		RetryAttempts: 3,
		WorkDir:       tmpDir,
	}

	t.Run("no PRD file", func(t *testing.T) {
		output := captureStdout(t, func() {
			if err := Display(cfg); err != nil {
				t.Errorf("Display() returned error: %v", err)
			}
		})
		expected := "No PRD file found. Run ralph with a prompt to create one.\n"
		if output != expected {
			t.Errorf("Expected %q, got %q", expected, output)
		}
	})

	t.Run("valid PRD file", func(t *testing.T) {
		testPRD := &prd.PRD{
			ProjectName: "Test Project",
			BranchName:  "main",
			Stories: []*prd.Story{
				{ID: "story-1", Title: "Completed story", Priority: 1, Passes: true},
				{ID: "story-2", Title: "Pending story", Priority: 2, Passes: false, RetryCount: 0},
				{ID: "story-3", Title: "Failed story", Priority: 3, Passes: false, RetryCount: 3},
			},
		}
		if err := prd.Save(cfg, testPRD); err != nil {
			t.Fatalf("Failed to save test PRD: %v", err)
		}

		output := captureStdout(t, func() {
			if err := Display(cfg); err != nil {
				t.Errorf("Display() returned error: %v", err)
			}
		})

		for _, want := range []string{
			"Project: Test Project (Branch: main)",
			"Stories: 3 total, 1 completed, 1 pending, 1 failed",
			"✓ [story-1] Completed story (priority: 1)",
			"⏳ [story-2] Pending story (priority: 2)",
			"✗ [story-3] Failed story (priority: 3)",
		} {
			if !strings.Contains(output, want) {
				t.Errorf("output missing %q\ngot: %s", want, output)
			}
		}
	})

	t.Run("PRD without branch name", func(t *testing.T) {
		testPRD := &prd.PRD{
			ProjectName: "Simple Project",
			Stories:     []*prd.Story{{ID: "story-1", Title: "Only story", Priority: 1, Passes: true}},
		}
		if err := prd.Save(cfg, testPRD); err != nil {
			t.Fatalf("Failed to save test PRD: %v", err)
		}

		output := captureStdout(t, func() {
			if err := Display(cfg); err != nil {
				t.Errorf("Display() returned error: %v", err)
			}
		})

		if !strings.Contains(output, "Project: Simple Project\n") {
			t.Errorf("expected project line without branch, got: %s", output)
		}
		if strings.Contains(output, "Branch:") {
			t.Error("output should not contain branch info when branch is empty")
		}
	})
}

func TestDisplay_EmptyStories(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "empty_prd.json", RetryAttempts: 3, WorkDir: tmpDir}

	if err := prd.Save(cfg, &prd.PRD{ProjectName: "Empty Project", Stories: []*prd.Story{}}); err != nil {
		t.Fatalf("Failed to save test PRD: %v", err)
	}

	output := captureStdout(t, func() {
		if err := Display(cfg); err != nil {
			t.Errorf("Display() returned error: %v", err)
		}
	})

	if !strings.Contains(output, "Stories: 0 total, 0 completed, 0 pending, 0 failed") {
		t.Errorf("expected zero counts, got: %s", output)
	}
}

func TestDisplay_AllCompletedStories(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "prd.json", RetryAttempts: 3, WorkDir: tmpDir}

	testPRD := &prd.PRD{
		ProjectName: "All Completed",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "First", Priority: 1, Passes: true},
			{ID: "story-2", Title: "Second", Priority: 2, Passes: true},
		},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("Failed to save test PRD: %v", err)
	}

	output := captureStdout(t, func() {
		if err := Display(cfg); err != nil {
			t.Errorf("Display() returned error: %v", err)
		}
	})

	if !strings.Contains(output, "Stories: 2 total, 2 completed, 0 pending, 0 failed") {
		t.Errorf("expected all completed counts, got: %s", output)
	}
}

func TestDisplay_AllFailedStories(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "prd.json", RetryAttempts: 2, WorkDir: tmpDir}

	testPRD := &prd.PRD{
		ProjectName: "All Failed",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "First", Priority: 1, Passes: false, RetryCount: 2},
			{ID: "story-2", Title: "Second", Priority: 2, Passes: false, RetryCount: 3},
		},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("Failed to save test PRD: %v", err)
	}

	output := captureStdout(t, func() {
		if err := Display(cfg); err != nil {
			t.Errorf("Display() returned error: %v", err)
		}
	})

	if !strings.Contains(output, "Stories: 2 total, 0 completed, 0 pending, 2 failed") {
		t.Errorf("expected all failed counts, got: %s", output)
	}
}

func TestDisplay_CorruptedPRD(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "corrupted.json", RetryAttempts: 3, WorkDir: tmpDir}

	if err := os.WriteFile(cfg.PRDPath(), []byte("{ invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write corrupted PRD: %v", err)
	}

	err := Display(cfg)
	if err == nil {
		t.Fatal("Display() should return error for corrupted PRD")
	}
	if !strings.Contains(err.Error(), "failed to load PRD") {
		t.Errorf("expected error containing 'failed to load PRD', got: %v", err)
	}
}
