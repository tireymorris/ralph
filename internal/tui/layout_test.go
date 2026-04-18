package tui

import (
	"testing"

	"ralph/internal/config"
)

func TestComputePaneHeightsFewLogsShrinksLogPane(t *testing.T) {
	mainMany, logMany := computePaneHeights(32, 200, 0, 0)
	mainFew, logFew := computePaneHeights(32, 2, 0, 0)
	if logFew >= logMany {
		t.Fatalf("expected fewer log lines with sparse logs: logFew=%d logMany=%d", logFew, logMany)
	}
	if mainFew <= mainMany {
		t.Fatalf("expected more main lines with sparse logs: mainFew=%d mainMany=%d", mainFew, mainMany)
	}
}

func TestComputePaneHeightsBiasExpandsLogs(t *testing.T) {
	_, log0 := computePaneHeights(40, 2, 0, 0)
	_, logPlus := computePaneHeights(40, 2, 5, 0)
	if logPlus <= log0 {
		t.Fatalf("positive bias should not shrink logs: log0=%d logPlus=%d", log0, logPlus)
	}
}

func TestComputePaneHeightsFullscreenMainReturnsFullHeight(t *testing.T) {
	termHeight := 40
	mainH, logH := computePaneHeights(termHeight, 10, 0, focusMain)
	if mainH != termHeight-scrollChrome {
		t.Errorf("expected main height %d with fullscreen focusMain, got %d", termHeight-scrollChrome, mainH)
	}
	if logH != 0 {
		t.Errorf("expected log height 0 with fullscreen focusMain, got %d", logH)
	}
}

func TestComputePaneHeightsFullscreenLogsReturnsFullHeight(t *testing.T) {
	termHeight := 40
	mainH, logH := computePaneHeights(termHeight, 10, 0, focusLogs)
	if logH != termHeight-scrollChrome {
		t.Errorf("expected log height %d with fullscreen focusLogs, got %d", termHeight-scrollChrome, logH)
	}
	if mainH != 0 {
		t.Errorf("expected main height 0 with fullscreen focusLogs, got %d", mainH)
	}
}

func TestApplyLayoutFullscreenMainHidesLogs(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.width = 120
	m.height = 40
	m.fullscreenPane = focusMain

	m.applyLayout(120, 40)

	if m.mainPane.Height <= 0 {
		t.Errorf("main pane height should be positive in fullscreen, got %d", m.mainPane.Height)
	}
}

func TestApplyLayoutFullscreenLogsHidesMain(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.width = 120
	m.height = 40
	m.fullscreenPane = focusLogs

	m.applyLayout(120, 40)

	if m.mainPane.Height != 0 {
		t.Errorf("main pane height should be 0 when logs are fullscreen, got %d", m.mainPane.Height)
	}
}

func TestFullscreenPersistsAcrossPhaseChange(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhasePRDReview
	m.scrollPane = focusMain
	m.fullscreenPane = focusMain

	// Simulate phase change
	m.phase = PhaseImplementation

	if m.fullscreenPane != focusMain {
		t.Errorf("fullscreenPane should persist across phase changes, got %v", m.fullscreenPane)
	}
}
