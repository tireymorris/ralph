package headless

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/session"
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

func TestRunRecoversFromImplementationReviewFindings(t *testing.T) {
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

	if err := os.WriteFile(filepath.Join(cfg.WorkDir, "prd.json"), []byte(`{"project_name":"Resume","stories":[{"id":"story-1","title":"S1","description":"d","slices":[{"id":"slice-1","behavior":"a","red_hint":"add failing test","passes":false}],"priority":1,"passes":false}]}`), 0o644); err != nil {
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

func TestRunSharesSnapshotFromSession(t *testing.T) {
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

	snap := sharedSnapshot(t, r)
	if snap.Prompt != "build a feature" {
		t.Fatalf("Prompt = %q, want %q", snap.Prompt, "build a feature")
	}
	if snap.CurrentStory == nil {
		t.Fatal("CurrentStory = nil, want shared story data")
	}
	if snap.CurrentStory.Title != "Mock story" {
		t.Fatalf("CurrentStory.Title = %q, want %q", snap.CurrentStory.Title, "Mock story")
	}
	if snap.NextPendingSlice == nil {
		t.Fatal("NextPendingSlice = nil, want shared slice data")
	}
	if snap.NextPendingSlice.ID != "slice-1" {
		t.Fatalf("NextPendingSlice.ID = %q, want %q", snap.NextPendingSlice.ID, "slice-1")
	}
	if snap.CompletedStories != 0 {
		t.Fatalf("CompletedStories = %d, want 0", snap.CompletedStories)
	}
	if snap.TotalStories != 1 {
		t.Fatalf("TotalStories = %d, want 1", snap.TotalStories)
	}
}

type autoContinueRunner struct {
	workDir     string
	reviewCalls int
	calls       []string
}

func (r *autoContinueRunner) Run(_ context.Context, promptText string, outputCh chan<- runner.OutputLine) error {
	r.calls = append(r.calls, promptKind(promptText))
	switch prompt.Kind(promptText) {
	case prompt.KindPRDGenerate, prompt.KindPRDCritiqueRevision, prompt.KindPRDClarificationRevision, prompt.KindFollowUp:
		data := `{"project_name":"Test","stories":[{"id":"story-1","title":"S1","description":"d","slices":[{"id":"slice-1","behavior":"a","red_hint":"add failing test"}],"priority":1}]}`
		return os.WriteFile(filepath.Join(r.workDir, "prd.json"), []byte(data), 0o644)
	case prompt.KindPRDSelfReview:
		return os.WriteFile(filepath.Join(r.workDir, prompt.PRDSelfReviewVerdictFile), []byte(`{"approved":true,"summary":"ok"}`), 0o644)
	case prompt.KindDiffReview:
		r.reviewCalls++
		if r.reviewCalls == 1 {
			_, _ = os.Stat(filepath.Join(r.workDir, "delta.txt"))
			outputCh <- runner.OutputLine{Text: findingsTranscript}
			return nil
		}
		outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
		return nil
	case prompt.KindRecovery:
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
	switch prompt.Kind(promptText) {
	case prompt.KindPRDGenerate, prompt.KindPRDCritiqueRevision, prompt.KindPRDClarificationRevision, prompt.KindFollowUp:
		return "prd"
	case prompt.KindPRDSelfReview:
		return "selfreview"
	case prompt.KindDiffReview:
		return "impl-review"
	case prompt.KindRecovery:
		return "recover"
	case prompt.KindStoryImplement:
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

func sharedSnapshot(t *testing.T, r *Runner) session.RunSnapshot {
	t.Helper()

	method := reflect.ValueOf(r).MethodByName("Snapshot")
	if !method.IsValid() {
		t.Fatal("Runner should expose a Snapshot method backed by the shared session snapshot")
	}
	results := method.Call(nil)
	if len(results) != 1 {
		t.Fatalf("Snapshot() returned %d values, want 1", len(results))
	}
	snap, ok := results[0].Interface().(session.RunSnapshot)
	if !ok {
		t.Fatalf("Snapshot() returned %T, want session.RunSnapshot", results[0].Interface())
	}
	return snap
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
