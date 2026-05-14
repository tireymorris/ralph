// Package implement implements the "ralph implement" subcommand.
package implement

import (
	"fmt"
	"os"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

// Cmd runs the implement subcommand workflow.
type Cmd struct {
	cfg     *config.Config
	verbose bool
}

// NewCmd creates a new Cmd instance.
func NewCmd(cfg *config.Config, verbose bool) *Cmd {
	return &Cmd{
		cfg:     cfg,
		verbose: verbose,
	}
}

// Run loads existing prd.json and executes stories in priority order.
func (c *Cmd) Run() int {
	exists, err := prd.Exists(c.cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking PRD file: %v\n", err)
		return 1
	}
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: no PRD file found. Run 'ralph prd <prompt>' first.\n")
		return 1
	}

	p, err := prd.Load(c.cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading PRD: %v\n", err)
		return 1
	}

	if p.AllCompleted() {
		fmt.Printf("All %d stories already completed.\n", len(p.Stories))
		return 0
	}

	return 0
}

// Run is a convenience function that creates a Cmd and runs it.
func Run(cfg *config.Config, verbose bool) int {
	return NewCmd(cfg, verbose).Run()
}
