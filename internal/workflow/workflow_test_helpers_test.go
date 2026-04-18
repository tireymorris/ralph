package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
)

func setupTestPRDFile(t *testing.T, dir string, p *prd.PRD) *config.Config {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.WorkDir = dir
	cfg.PRDFile = "prd.json"

	if p != nil {
		prdPath := filepath.Join(dir, "prd.json")
		data := `{"project_name":"` + p.ProjectName + `","stories":[`
		for i, s := range p.Stories {
			if i > 0 {
				data += ","
			}
			passesStr := "false"
			if s.Passes {
				passesStr = "true"
			}
			data += `{"id":"` + s.ID + `","title":"` + s.Title + `","description":"` + s.Description + `","acceptance_criteria":["AC"],"priority":` + string(rune('0'+s.Priority)) + `,"passes":` + passesStr + `,"retry_count":` + string(rune('0'+s.RetryCount)) + `}`
		}
		data += `]}`
		if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
			t.Fatalf("failed to write test PRD: %v", err)
		}
	}

	return cfg
}
