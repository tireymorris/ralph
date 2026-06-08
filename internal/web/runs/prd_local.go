package runs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
)

// LocalPRDRunID is the stable API id for an in-progress TUI/CLI run backed only by prd.json.
const LocalPRDRunID = runstate.LocalRunID

// OngoingLocalPRD reports a synthetic run when prd.json exists, is incomplete, and no active web run
// is already tracked for the work directory.
func OngoingLocalPRD(cfg *config.Config, registry *Registry) (*Run, bool) {
	if _, ok := registry.ActiveForWorkDir(cfg.WorkDir); ok {
		return nil, false
	}

	exists, err := prd.Exists(cfg)
	if err != nil || !exists {
		return nil, false
	}

	p, err := prd.Load(cfg)
	if err != nil || p.AllCompleted() {
		return nil, false
	}

	prdPath := cfg.PRDPath()
	info, err := os.Stat(prdPath)
	if err != nil {
		return nil, false
	}
	mod := info.ModTime()
	if mod.IsZero() {
		mod = time.Now()
	}

	meta := loadLocalPRDMeta(cfg.WorkDir)
	status, phase := localPRDStatus(p, meta.Checkpoint)
	run := &Run{
		ID:                         LocalPRDRunID,
		WorkDir:                    cfg.WorkDir,
		Prompt:                     localPRDPrompt(p),
		Status:                     status,
		Phase:                      phase,
		CreatedAt:                  mod,
		UpdatedAt:                  mod,
		PRDPath:                    cfg.PRDFile,
		Checkpoint:                 meta.Checkpoint,
		ReviewIteration:            meta.ReviewIteration,
		ReviewFingerprint:          meta.ReviewFingerprint,
		ReviewElapsedMs:            meta.ReviewElapsedMs,
		StopReason:                 meta.StopReason,
		LastReviewTranscriptPath:   meta.LastReviewTranscriptPath,
		LastReviewChangedFilesHash: meta.LastReviewChangedFilesHash,
	}
	if !meta.UpdatedAt.IsZero() {
		run.UpdatedAt = meta.UpdatedAt
	}
	return run, true
}

type localPRDMeta struct {
	Checkpoint                 string    `json:"checkpoint,omitempty"`
	ReviewIteration            int       `json:"review_iteration,omitempty"`
	ReviewFingerprint          string    `json:"review_fingerprint,omitempty"`
	ReviewElapsedMs            int64     `json:"review_elapsed_ms,omitempty"`
	StopReason                 string    `json:"stop_reason,omitempty"`
	LastReviewTranscriptPath   string    `json:"last_review_transcript_path,omitempty"`
	LastReviewChangedFilesHash string    `json:"last_review_changed_files_hash,omitempty"`
	UpdatedAt                  time.Time `json:"updated_at,omitempty"`
}

func loadLocalPRDMeta(workDir string) localPRDMeta {
	path := filepath.Join(workDir, ".ralph", "runs", LocalPRDRunID, "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return localPRDMeta{}
	}
	var m localPRDMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return localPRDMeta{}
	}
	return m
}

func localPRDStatus(p *prd.PRD, checkpoint string) (status, phase string) {
	return runstate.LocalPRDStatusPhase(p, checkpoint)
}

func localPRDPrompt(p *prd.PRD) string {
	if name := strings.TrimSpace(p.ProjectName); name != "" {
		return name
	}
	ctx := strings.TrimSpace(p.Context)
	if ctx == "" {
		return "Local PRD run"
	}
	if i := strings.IndexByte(ctx, '\n'); i >= 0 {
		ctx = strings.TrimSpace(ctx[:i])
	}
	const maxLen = 120
	if len(ctx) > maxLen {
		return ctx[:maxLen] + "…"
	}
	return ctx
}
