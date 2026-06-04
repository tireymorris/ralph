package review

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"sort"
)

type fingerprintPayload struct {
	Findings []Finding `json:"Findings"`
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
