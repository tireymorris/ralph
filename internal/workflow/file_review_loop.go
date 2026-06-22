package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ralph/internal/shared/runpaths"
	"ralph/internal/shared/runstate"
)

type fileRunMeta struct {
	Checkpoint                 string    `json:"checkpoint,omitempty"`
	ReviewIteration            int       `json:"review_iteration,omitempty"`
	ReviewFingerprint          string    `json:"review_fingerprint,omitempty"`
	ReviewElapsedMs            int64     `json:"review_elapsed_ms,omitempty"`
	StopReason                 string    `json:"stop_reason,omitempty"`
	LastReviewTranscriptPath   string    `json:"last_review_transcript_path,omitempty"`
	LastReviewChangedFilesHash string    `json:"last_review_changed_files_hash,omitempty"`
	RecoveryAttempts           int       `json:"recovery_attempts,omitempty"`
	UpdatedAt                  time.Time `json:"updated_at,omitempty"`
}

type FileReviewLoop struct {
	workDir string
	runID   string
}

func NewFileReviewLoop(workDir, runID string) *FileReviewLoop {
	return &FileReviewLoop{workDir: workDir, runID: runID}
}

func (f *FileReviewLoop) metaPath() string {
	return runpaths.MetaPath(f.workDir, f.runID)
}

func (f *FileReviewLoop) load() (fileRunMeta, error) {
	data, err := os.ReadFile(f.metaPath())
	if err != nil {
		if os.IsNotExist(err) {
			return fileRunMeta{}, nil
		}
		return fileRunMeta{}, err
	}
	var m fileRunMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return fileRunMeta{}, err
	}
	return m, nil
}

func (f *FileReviewLoop) Checkpoint() string {
	m, err := f.load()
	if err != nil {
		return ""
	}
	return m.Checkpoint
}

func (f *FileReviewLoop) Snapshot() (iteration int, fingerprint string, elapsedMs int64, changedFilesHash string) {
	m, err := f.load()
	if err != nil {
		return 0, "", 0, ""
	}
	return m.ReviewIteration, m.ReviewFingerprint, m.ReviewElapsedMs, m.LastReviewChangedFilesHash
}

func (f *FileReviewLoop) Apply(u ReviewLoopUpdate) error {
	m, err := f.load()
	if err != nil {
		return err
	}
	state := reviewLoopStateFromMeta(m)
	runstate.ApplyReviewLoopUpdate(&state, u)
	applyReviewLoopStateToMeta(&m, state)
	m.UpdatedAt = time.Now()

	dir := filepath.Dir(f.metaPath())
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create run meta dir: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(f.metaPath(), data, 0o600); err != nil {
		return fmt.Errorf("write run meta: %w", err)
	}
	return nil
}

func (f *FileReviewLoop) EnsureCheckpoint(cp string) error {
	m, err := f.load()
	if err != nil {
		return err
	}
	if m.Checkpoint != "" {
		return nil
	}
	m.Checkpoint = cp
	m.UpdatedAt = time.Now()
	dir := filepath.Dir(f.metaPath())
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f.metaPath(), data, 0o600)
}

func (f *FileReviewLoop) RecoveryAttempts() int {
	m, err := f.load()
	if err != nil {
		return 0
	}
	return m.RecoveryAttempts
}

func (f *FileReviewLoop) LastReviewTranscriptPath() string {
	m, err := f.load()
	if err != nil {
		return ""
	}
	return m.LastReviewTranscriptPath
}

func reviewLoopStateFromMeta(m fileRunMeta) runstate.ReviewLoopState {
	return runstate.ReviewLoopState{
		Checkpoint:                 m.Checkpoint,
		ReviewIteration:            m.ReviewIteration,
		ReviewFingerprint:          m.ReviewFingerprint,
		ReviewElapsedMs:            m.ReviewElapsedMs,
		StopReason:                 m.StopReason,
		LastReviewTranscriptPath:   m.LastReviewTranscriptPath,
		LastReviewChangedFilesHash: m.LastReviewChangedFilesHash,
		RecoveryAttempts:           m.RecoveryAttempts,
	}
}

func applyReviewLoopStateToMeta(m *fileRunMeta, state runstate.ReviewLoopState) {
	m.Checkpoint = state.Checkpoint
	m.ReviewIteration = state.ReviewIteration
	m.ReviewFingerprint = state.ReviewFingerprint
	m.ReviewElapsedMs = state.ReviewElapsedMs
	m.StopReason = state.StopReason
	m.LastReviewTranscriptPath = state.LastReviewTranscriptPath
	m.LastReviewChangedFilesHash = state.LastReviewChangedFilesHash
	m.RecoveryAttempts = state.RecoveryAttempts
}

var _ ReviewLoopUpdater = (*FileReviewLoop)(nil)
