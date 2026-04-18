package cli

import (
	"fmt"
	"os"

	"ralph/internal/config"
	"ralph/internal/status"
)

// RunStatus prints PRD progress to stdout.
func RunStatus(cfg *config.Config) error {
	if err := status.Display(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error displaying status: %v\n", err)
		return err
	}
	return nil
}
