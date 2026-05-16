package status

import (
	"fmt"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

func Display(cfg *config.Config) error {

	exists, err := prd.Exists(cfg)
	if err != nil {
		return fmt.Errorf("checking PRD file: %w", err)
	}
	if !exists {
		fmt.Println("No PRD file found. Run ralph with a prompt to create one.")
		return nil
	}

	p, err := prd.Load(cfg)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	fmt.Printf("Project: %s", p.ProjectName)
	if p.BranchName != "" {
		fmt.Printf(" (Branch: %s)", p.BranchName)
	}
	fmt.Println()

	total := len(p.Stories)
	completed := p.CompletedCount()
	pending := total - completed

	fmt.Printf("Stories: %d total, %d completed, %d pending\n",
		total, completed, pending)

	for _, story := range p.Stories {
		status := "⏳"
		if story.Passes {
			status = "✓"
		}

		fmt.Printf("%s [%s] %s (priority: %d)\n",
			status, story.ID, story.Title, story.Priority)
	}

	return nil
}
