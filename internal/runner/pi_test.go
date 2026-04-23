package runner

import (
	"strings"
	"testing"
)

func TestPiProviderAndModel(t *testing.T) {
	tests := []struct {
		cfg    string
		wantP  string
		wantM  string
	}{
		{"pi/auto", "cursor", "auto"},
		{"pi/sonnet", "cursor", "sonnet"},
		{"pi/openai/gpt-4o", "openai", "gpt-4o"},
		{"pi/cursor/foo/bar", "cursor", "foo/bar"},
	}
	for _, tt := range tests {
		t.Run(tt.cfg, func(t *testing.T) {
			p, m := piProviderAndModel(tt.cfg)
			if p != tt.wantP || m != tt.wantM {
				t.Fatalf("piProviderAndModel(%q) = (%q,%q), want (%q,%q)", tt.cfg, p, m, tt.wantP, tt.wantM)
			}
		})
	}
}

func TestParsePiJSONLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "text_delta",
			line: `{"type":"message_update","assistantMessageEvent":{"type":"text_delta","delta":"hi"}}`,
			want: []string{"hi"},
		},
		{
			name: "tool start",
			line: `{"type":"tool_execution_start","toolCallId":"1","toolName":"read","args":{}}`,
			want: []string{"Using tool: read"},
		},
		{
			name: "tool end error",
			line: `{"type":"tool_execution_end","toolCallId":"1","toolName":"bash","result":null,"isError":true}`,
			want: []string{"Tool bash failed"},
		},
		{
			name: "session skipped",
			line: `{"type":"session","version":3,"id":"x"}`,
			want: nil,
		},
		{
			name: "invalid json",
			line: `not json`,
			want: []string{"not json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePiJSONLine(tt.line)
			if tt.want == nil {
				if len(got) != 0 {
					t.Fatalf("parsePiJSONLine() = %v, want nil", got)
				}
				return
			}
			var texts []string
			for _, o := range got {
				texts = append(texts, o.Text)
			}
			if strings.Join(texts, "|") != strings.Join(tt.want, "|") {
				t.Errorf("parsePiJSONLine() texts = %v, want %v", texts, tt.want)
			}
		})
	}
}
