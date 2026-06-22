package runs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"ralph/internal/shared/logger"
	"ralph/internal/shared/runstate"
)

const runDirPerm = 0o750

type Run struct {
	ID                         string    `json:"id"`
	WorkDir                    string    `json:"work_dir"`
	Prompt                     string    `json:"prompt"`
	Status                     string    `json:"status"`
	Phase                      string    `json:"phase"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
	PRDPath                    string    `json:"prd_path"`
	AutoApprove                bool      `json:"auto_approve,omitempty"`
	Checkpoint                 string    `json:"checkpoint,omitempty"`
	ReviewIteration            int       `json:"review_iteration,omitempty"`
	ReviewFingerprint          string    `json:"review_fingerprint,omitempty"`
	ReviewElapsedMs            int64     `json:"review_elapsed_ms,omitempty"`
	StopReason                 string    `json:"stop_reason,omitempty"`
	LastReviewTranscriptPath   string    `json:"last_review_transcript_path,omitempty"`
	LastReviewChangedFilesHash string    `json:"last_review_changed_files_hash,omitempty"`
	RecoveryAttempts           int       `json:"recovery_attempts,omitempty"`
}

type ReviewLoopUpdate struct {
	Checkpoint                 string
	ReviewIteration            int
	ReviewFingerprint          string
	ReviewElapsedMs            int64
	StopReason                 string
	LastReviewTranscriptPath   string
	LastReviewChangedFilesHash string
	RecoveryAttempts           int
}

type LifecycleUpdate struct {
	Status     string
	Phase      string
	Checkpoint string
}

type Registry struct {
	mu   sync.RWMutex
	runs map[string]*Run
}

func NewRegistry() *Registry {
	return &Registry{
		runs: make(map[string]*Run),
	}
}

func (r *Registry) LoadFromWorkDir(workDir string) error {
	root := filepath.Join(workDir, ".ralph", "runs")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read runs dir %q: %w", root, err)
	}

	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		metaPath := filepath.Join(root, ent.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			logger.Warn("skipping unreadable run meta", "path", metaPath, "error", err)
			continue
		}
		var run Run
		if err := json.Unmarshal(data, &run); err != nil {
			logger.Warn("skipping corrupt run meta", "path", metaPath, "error", err)
			continue
		}
		if run.ID == "" {
			run.ID = ent.Name()
		}
		if run.WorkDir == "" {
			run.WorkDir = workDir
		}
		if run.PRDPath != "" && (strings.Contains(run.PRDPath, "..") || filepath.IsAbs(run.PRDPath)) {
			logger.Warn("skipping run with unsafe PRDPath", "id", run.ID, "prd_path", run.PRDPath)
			continue
		}
		stored := run
		r.mu.Lock()
		r.runs[run.ID] = &stored
		r.mu.Unlock()
	}
	return nil
}

func (r *Registry) Register(run *Run) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	stored := *run
	r.runs[run.ID] = &stored

	if err := persistRun(&stored); err != nil {
		delete(r.runs, run.ID)
		return err
	}
	return nil
}

func persistRun(run *Run) error {
	dir := filepath.Join(run.WorkDir, ".ralph", "runs", run.ID)
	if err := os.MkdirAll(dir, runDirPerm); err != nil {
		return fmt.Errorf("create run dir %q: %w", dir, err)
	}

	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal run meta: %w", err)
	}

	metaPath := filepath.Join(dir, "meta.json")
	if err := os.WriteFile(metaPath, data, 0600); err != nil {
		return fmt.Errorf("write run meta %q: %w", metaPath, err)
	}
	return nil
}

func (r *Registry) Get(id string) (*Run, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	run, ok := r.runs[id]
	if !ok {
		return nil, false
	}
	copy := *run
	return &copy, true
}

func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs = make(map[string]*Run)
}

func (r *Registry) List() []*Run {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*Run, 0, len(r.runs))
	for _, run := range r.runs {
		copy := *run
		list = append(list, &copy)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	return list
}

func IsTerminalStatus(status string) bool {
	switch status {
	case runstate.StatusCompleted, runstate.StatusFailed, runstate.StatusCancelled:
		return true
	default:
		return false
	}
}

var activeRunStatuses = map[string]bool{
	runstate.StatusRunning:           true,
	runstate.StatusWaitingClarify:    true,
	runstate.StatusWaitingReview:     true,
	runstate.StatusWaitingImplReview: true,
	runstate.StatusImplementing:      true,
}

func (r *Registry) ActiveForWorkDir(workDir string) (*Run, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, run := range r.runs {
		if run.WorkDir == workDir && activeRunStatuses[run.Status] {
			copy := *run
			return &copy, true
		}
	}
	return nil, false
}

func (r *Registry) UpdateStatus(id, status, phase string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, ok := r.runs[id]
	if !ok {
		return fmt.Errorf("run %q not found", id)
	}

	run.Status = status
	run.Phase = phase
	run.UpdatedAt = time.Now()

	return persistRun(run)
}

func (r *Registry) UpdateLifecycle(id string, u LifecycleUpdate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, ok := r.runs[id]
	if !ok {
		return fmt.Errorf("run %q not found", id)
	}

	if u.Status != "" {
		run.Status = u.Status
	}
	if u.Phase != "" {
		run.Phase = u.Phase
	}
	if u.Checkpoint != "" {
		run.Checkpoint = u.Checkpoint
	}
	run.UpdatedAt = time.Now()

	return persistRun(run)
}

func (r *Registry) UpdateCheckpoint(id, checkpoint string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, ok := r.runs[id]
	if !ok {
		return fmt.Errorf("run %q not found", id)
	}

	run.Checkpoint = checkpoint
	run.UpdatedAt = time.Now()

	return persistRun(run)
}

func (r *Registry) UpdateReviewLoop(id string, u ReviewLoopUpdate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, ok := r.runs[id]
	if !ok {
		return fmt.Errorf("run %q not found", id)
	}

	run.Checkpoint = u.Checkpoint
	run.ReviewIteration = u.ReviewIteration
	run.ReviewFingerprint = u.ReviewFingerprint
	run.ReviewElapsedMs = u.ReviewElapsedMs
	if u.StopReason != "" {
		run.StopReason = u.StopReason
	}
	run.LastReviewTranscriptPath = u.LastReviewTranscriptPath
	run.LastReviewChangedFilesHash = u.LastReviewChangedFilesHash
	if u.RecoveryAttempts > 0 {
		run.RecoveryAttempts = u.RecoveryAttempts
	}
	run.UpdatedAt = time.Now()

	return persistRun(run)
}
