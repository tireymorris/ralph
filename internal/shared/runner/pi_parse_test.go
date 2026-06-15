package runner

import "testing"

func TestParsePiJSONLine_TextDeltaAppends(t *testing.T) {
	line := `{"type":"message_update","assistantMessageEvent":{"type":"text_delta","delta":"hello"}}`
	lines := parsePiJSONLine(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if lines[0].Text != "hello" {
		t.Errorf("Text = %q, want %q", lines[0].Text, "hello")
	}
	if !lines[0].Append {
		t.Error("Append should be true for text deltas")
	}
}
