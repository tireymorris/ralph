package status

import (
	"fmt"

	"ralph/internal/config"
	"ralph/internal/prd"
)

// Display loads and prints a formatted PRD status summary to stdout
func Display(cfg *config.Config) error {
	// Check if PRD file exists first
	if !prd.Exists(cfg) {
		fmt.Println("No PRD file found. Run ralph with a prompt to create one.")
		return nil
	}

	// Load the PRD file
	p, err := prd.Load(cfg)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}

	// Print project info
	fmt.Printf("Project: %s", p.ProjectName)
	if p.BranchName != "" {
		fmt.Printf(" (Branch: %s)", p.BranchName)
	}
	fmt.Println()

	// Calculate story counts
	total := len(p.Stories)
	completed := p.CompletedCount()
	failed := len(p.FailedStories(cfg.RetryAttempts))
	pending := total - completed - failed

	// Print story counts
	fmt.Printf("Stories: %d total, %d completed, %d pending, %d failed\n",
		total, completed, pending, failed)

	// Print individual stories
	for _, story := range p.Stories {
		status := "⏳"
		if story.Passes {
			status = "✓"
		} else if story.RetryCount >= cfg.RetryAttempts {
			status = "✗"
		}

		fmt.Printf("%s [%s] %s (priority: %d)\n",
			status, story.ID, story.Title, story.Priority)
	}

	return nil
}
