package review

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"ralph/internal/workflow/events"
)

type Finding = events.ImplementationFinding

const (
	findingsBlockStart = "===ralph-findings==="
	findingsBlockEnd   = "===/ralph-findings==="
)

type fingerprintPayload struct {
	Findings []Finding `json:"Findings"`
}

type rawFinding struct {
	Category string `json:"category"`
	Path     string `json:"path"`
	Line     *int   `json:"line,omitempty"`
	Summary  string `json:"summary"`
}

func ParseFindings(transcript string) ([]Finding, error) {
	block, ok := extractFindingsBlock(transcript)
	if !ok {
		return nil, nil
	}
	var raw []rawFinding
	if err := json.Unmarshal([]byte(block), &raw); err != nil {
		return nil, fmt.Errorf("parse findings JSON: %w", err)
	}
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]Finding, 0, len(raw))
	for _, r := range raw {
		line := 0
		if r.Line != nil {
			line = *r.Line
		}
		out = append(out, Finding{
			ID:       findingID(r.Category, r.Path, line, r.Summary),
			Category: r.Category,
			Path:     r.Path,
			Line:     line,
			Summary:  r.Summary,
		})
	}
	return out, nil
}

func extractFindingsBlock(transcript string) (string, bool) {
	start := strings.Index(transcript, findingsBlockStart)
	if start < 0 {
		return "", false
	}
	rest := transcript[start+len(findingsBlockStart):]
	end := strings.Index(rest, findingsBlockEnd)
	if end < 0 {
		return "", false
	}
	return strings.TrimSpace(rest[:end]), true
}

func findingID(category, path string, line int, summary string) string {
	key := fmt.Sprintf("%s\x00%s\x00%d\x00%s", category, path, line, summary)
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:8])
}

func Fingerprint(findings []Finding) string {
	if len(findings) == 0 {
		return ""
	}
	sorted := slices.Clone(findings)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})
	data, err := json.Marshal(fingerprintPayload{Findings: sorted})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
