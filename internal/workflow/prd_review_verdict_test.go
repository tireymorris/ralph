package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/prompt"
)

func TestParsePRDReviewVerdict(t *testing.T) {
	t.Run("valid verdict", func(t *testing.T) {
		v := parsePRDReviewVerdict([]byte(`{"approved": true, "summary": "looks good"}`))
		if !v.Approved {
			t.Error("expected Approved to be true")
		}
		if v.Summary != "looks good" {
			t.Errorf("Summary = %q, want %q", v.Summary, "looks good")
		}
	})

	t.Run("malformed json returns zero-value verdict", func(t *testing.T) {
		v := parsePRDReviewVerdict([]byte(`not json`))
		if v != (PRDReviewVerdict{}) {
			t.Errorf("expected zero-value verdict, got %+v", v)
		}
	})

	t.Run("empty data returns zero-value verdict", func(t *testing.T) {
		v := parsePRDReviewVerdict(nil)
		if v != (PRDReviewVerdict{}) {
			t.Errorf("expected zero-value verdict, got %+v", v)
		}
	})
}

func TestPRDReviewVerdictReaderReadRemove(t *testing.T) {
	t.Run("reads verdict and deletes file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, prompt.PRDSelfReviewVerdictFile)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(`{"approved": true, "summary": "ok"}`), 0o644); err != nil {
			t.Fatal(err)
		}

		v, err := PRDReviewVerdictReader{WorkDir: dir}.ReadRemove()
		if err != nil {
			t.Fatalf("ReadRemove() error = %v", err)
		}
		if !v.Approved || v.Summary != "ok" {
			t.Errorf("verdict = %+v, want approved with summary %q", v, "ok")
		}
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Error("expected verdict file to be deleted")
		}
	})

	t.Run("missing file returns zero-value verdict and no error", func(t *testing.T) {
		v, err := PRDReviewVerdictReader{WorkDir: t.TempDir()}.ReadRemove()
		if err != nil {
			t.Fatalf("ReadRemove() error = %v", err)
		}
		if v != (PRDReviewVerdict{}) {
			t.Errorf("expected zero-value verdict, got %+v", v)
		}
	})

	t.Run("malformed file returns zero-value verdict and no error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, prompt.PRDSelfReviewVerdictFile)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(`not json`), 0o644); err != nil {
			t.Fatal(err)
		}

		v, err := PRDReviewVerdictReader{WorkDir: dir}.ReadRemove()
		if err != nil {
			t.Fatalf("ReadRemove() error = %v", err)
		}
		if v != (PRDReviewVerdict{}) {
			t.Errorf("expected zero-value verdict, got %+v", v)
		}
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Error("expected verdict file to be deleted")
		}
	})
}
