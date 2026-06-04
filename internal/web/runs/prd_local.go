package runs

import (
	"os"
	"strings"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

// LocalPRDRunID is the stable API id for an in-progress TUI/CLI run backed only by prd.json.
const LocalPRDRunID = "prd-local"

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

	status, phase := localPRDStatus(p)
	run := &Run{
		ID:        LocalPRDRunID,
		WorkDir:   cfg.WorkDir,
		Prompt:    localPRDPrompt(p),
		Status:    status,
		Phase:     phase,
		CreatedAt: mod,
		UpdatedAt: mod,
		PRDPath:   cfg.PRDFile,
	}
	return run, true
}

func localPRDStatus(p *prd.PRD) (status, phase string) {
	if len(p.Stories) == 0 {
		return "running", "generate"
	}
	return "implementing", "implement"
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
