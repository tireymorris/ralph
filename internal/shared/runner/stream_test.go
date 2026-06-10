package runner

import (
	"strings"
	"testing"

	"ralph/internal/shared/constants"
)

func passthroughTransform(line string) []OutputLine {
	return []OutputLine{{Text: line}}
}

func collectPipeLines(t *testing.T, input string) []string {
	t.Helper()
	outputCh := make(chan OutputLine, 16)
	if err := readPipeLines(strings.NewReader(input), outputCh, passthroughTransform); err != nil {
		t.Fatalf("readPipeLines() error = %v", err)
	}
	close(outputCh)
	var lines []string
	for out := range outputCh {
		lines = append(lines, out.Text)
	}
	return lines
}

func TestReadPipeLinesHandlesLinesBeyondScannerLimit(t *testing.T) {
	longLine := strings.Repeat("x", 2*1024*1024)
	lines := collectPipeLines(t, "before\n"+longLine+"\nafter\n")

	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[1] != longLine {
		t.Fatalf("long line corrupted: got %d bytes, want %d", len(lines[1]), len(longLine))
	}
}

func TestReadPipeLinesTrimsCRLF(t *testing.T) {
	lines := collectPipeLines(t, "one\r\ntwo\r\n")

	if len(lines) != 2 || lines[0] != "one" || lines[1] != "two" {
		t.Fatalf("got %q, want [one two]", lines)
	}
}

func TestReadPipeLinesEmitsFinalLineWithoutNewline(t *testing.T) {
	lines := collectPipeLines(t, "first\nlast")

	if len(lines) != 2 || lines[1] != "last" {
		t.Fatalf("got %q, want final line %q", lines, "last")
	}
}

func TestReadPipeLinesRejectsOversizedLine(t *testing.T) {
	oversized := strings.Repeat("x", constants.MaxPipeLineSize+1) + "\n"
	outputCh := make(chan OutputLine, 1)
	err := readPipeLines(strings.NewReader(oversized), outputCh, passthroughTransform)
	if err == nil {
		t.Fatal("readPipeLines() error = nil, want line size error")
	}
	if !strings.Contains(err.Error(), "line exceeds") {
		t.Fatalf("readPipeLines() error = %v, want line exceeds message", err)
	}
}
