package tui

import "testing"

func TestComputePaneHeightsFewLogsShrinksLogPane(t *testing.T) {
	mainMany, logMany := computePaneHeights(32, 200, 0)
	mainFew, logFew := computePaneHeights(32, 2, 0)
	if logFew >= logMany {
		t.Fatalf("expected fewer log lines with sparse logs: logFew=%d logMany=%d", logFew, logMany)
	}
	if mainFew <= mainMany {
		t.Fatalf("expected more main lines with sparse logs: mainFew=%d mainMany=%d", mainFew, mainMany)
	}
}

func TestComputePaneHeightsBiasExpandsLogs(t *testing.T) {
	_, log0 := computePaneHeights(40, 2, 0)
	_, logPlus := computePaneHeights(40, 2, 5)
	if logPlus <= log0 {
		t.Fatalf("positive bias should not shrink logs: log0=%d logPlus=%d", log0, logPlus)
	}
}
