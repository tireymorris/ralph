package headless

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"os/exec"
	"strings"
	"testing"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
)

func TestRunCompletesUnattended(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.PRDFile = "prd.json"
	cfg.Runner = "mock"
	cfg.SkipCleanup = true
	initGitRepo(t, cfg.WorkDir)
	if err := os.WriteFile(filepath.Join(cfg.WorkDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitFile(t, cfg.WorkDir, "main.go", "init source file")

	var stderr bytes.Buffer
	r := New(cfg, runner.NewMock(cfg), &stderr)

	code := r.Run("build a feature", false)
	if code != 0 {
		t.Fatalf("Run() = %d, want 0; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), `"type":"EventCompleted"`) {
		t.Fatalf("stderr = %q, want EventCompleted JSON", stderr.String())
	}
}

func TestRunAutoContinuesImplementationReview(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.PRDFile = "prd.json"
	cfg.Runner = "mock"
	cfg.SkipCleanup = true
	initGitRepo(t, cfg.WorkDir)
	if err := os.WriteFile(filepath.Join(cfg.WorkDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitFile(t, cfg.WorkDir, "main.go", "init source file")

	var stderr bytes.Buffer
	auto := &autoContinueRunner{
		workDir: cfg.WorkDir,
	}
	r := New(cfg, auto, &stderr)

	code := r.Run("build a feature", false)
	if code != 0 && code != 1 {
		t.Fatalf("Run() = %d, want 0 or 1; stderr=%s; calls=%#v", code, stderr.String(), auto.calls)
	}
	if auto.reviewCalls < 1 {
		t.Fatalf("reviewCalls = %d, want at least 1", auto.reviewCalls)
	}
	if !strings.Contains(stderr.String(), `"type":"EventImplementationReview"`) {
		t.Fatalf("stderr = %q, want implementation review JSON", stderr.String())
	}
	if !containsKind(auto.calls, "recover") {
		t.Fatalf("calls = %#v, want recovery prompt", auto.calls)
	}
}

func TestRunResumesFromCheckpoint(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.PRDFile = "prd.json"
	cfg.Runner = "mock"
	cfg.SkipCleanup = true
	initGitRepo(t, cfg.WorkDir)
	if err := os.WriteFile(filepath.Join(cfg.WorkDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitFile(t, cfg.WorkDir, "main.go", "init source file")

	if err := os.WriteFile(filepath.Join(cfg.WorkDir, "prd.json"), []byte(`{"project_name":"Resume","stories":[{"id":"story-1","title":"S1","description":"d","acceptance_criteria":["a"],"priority":1,"passes":false}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	loopDir := filepath.Join(cfg.WorkDir, ".ralph", "runs", "prd-local")
	if err := os.MkdirAll(loopDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loopDir, "meta.json"), []byte(`{"checkpoint":"impl_review"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	r := New(cfg, runner.NewMock(cfg), &stderr)

	code := r.Run("", true)
	if code != 0 {
		t.Fatalf("Run() = %d, want 0; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), `"type":"EventCompleted"`) {
		t.Fatalf("stderr = %q, want EventCompleted JSON", stderr.String())
	}
}

type autoContinueRunner struct {
	workDir      string
	reviewCalls  int
	calls        []string
}

func (r *autoContinueRunner) Run(_ context.Context, promptText string, outputCh chan<- runner.OutputLine) error {
	r.calls = append(r.calls, promptKind(promptText))
	switch {
	case strings.Contains(promptText, "Write the PRD file") || strings.Contains(promptText, "Write the updated PRD file"):
		data := `{"project_name":"Test","stories":[{"id":"story-1","title":"S1","description":"d","acceptance_criteria":["a"],"priority":1}]}`
		return os.WriteFile(filepath.Join(r.workDir, "prd.json"), []byte(data), 0o644)
	case strings.Contains(promptText, prompt.PRDSelfReviewVerdictFile):
		return os.WriteFile(filepath.Join(r.workDir, prompt.PRDSelfReviewVerdictFile), []byte(`{"approved":true,"summary":"ok"}`), 0o644)
	case strings.Contains(promptText, "critical diff review"):
		r.reviewCalls++
		if r.reviewCalls == 1 {
			_, _ = os.Stat(filepath.Join(r.workDir, "delta.txt"))
			outputCh <- runner.OutputLine{Text: findingsTranscript}
			return nil
		}
		outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
		return nil
	case strings.Contains(promptText, "Recovery") || strings.Contains(promptText, "recover"):
		if err := os.WriteFile(filepath.Join(r.workDir, "main.go"), []byte("package main\n// fixed\n"), 0o644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(r.workDir, "delta.txt"), []byte("fixed\n"), 0o644)
	default:
		return nil
	}
}

func (r *autoContinueRunner) RunnerName() string  { return "mock" }
func (r *autoContinueRunner) CommandName() string { return "mock" }
func (r *autoContinueRunner) IsInternalLog(string) bool {
	return false
}

const (
	findingsTranscript    = "===ralph-findings===\n[{\"category\":\"bug\",\"path\":\"delta.txt\",\"summary\":\"issue\"}]\n===/ralph-findings===\n"
	cleanReviewTranscript = "===ralph-findings===\n[]\n===/ralph-findings===\n"
)

func promptKind(promptText string) string {
	switch {
	case strings.Contains(promptText, "Write the PRD file") || strings.Contains(promptText, "Write the updated PRD file"):
		return "prd"
	case strings.Contains(promptText, prompt.PRDSelfReviewVerdictFile):
		return "selfreview"
	case strings.Contains(promptText, "critical diff review"):
		return "impl-review"
	case strings.Contains(promptText, "Recovery") || strings.Contains(promptText, "recover"):
		return "recover"
	case strings.Contains(promptText, "Implement story:"):
		return "implement"
	default:
		return "other"
	}
}

func containsKind(kinds []string, want string) bool {
	for _, kind := range kinds {
		if kind == want {
			return true
		}
	}
	return false
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "t@example.com"},
		{"config", "user.name", "t"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func commitFile(t *testing.T, dir, file, message string) {
	t.Helper()
	for _, args := range [][]string{
		{"add", file},
		{"commit", "-m", message},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}
