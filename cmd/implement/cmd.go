// Package implement implements the "ralph implement" subcommand.
package implement

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

type workflowExecutor interface {
	RunLoad(ctx context.Context) (*prd.PRD, error)
	RunImplementation(ctx context.Context, p *prd.PRD) error
}

// Cmd runs the implement subcommand workflow.
type Cmd struct {
	cfg      *config.Config
	verbose  bool
	executor workflowExecutor
	eventsCh chan workflow.Event
}

// NewCmd creates a new Cmd instance.
func NewCmd(cfg *config.Config, verbose bool) *Cmd {
	eventsCh := make(chan workflow.Event, constants.EventChannelBuffer)
	return &Cmd{
		cfg:      cfg,
		verbose:  verbose,
		eventsCh: eventsCh,
		executor: workflow.NewExecutor(cfg, eventsCh),
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	doneCh := make(chan int, 1)
	go c.handleEvents(c.eventsCh, doneCh)

	if err := c.executor.RunImplementation(ctx, p); err != nil {
		logger.Error("implementation failed", "error", err)
		close(c.eventsCh)
		<-doneCh
		return 1
	}

	close(c.eventsCh)
	return <-doneCh
}

func (c *Cmd) handleEvents(eventsCh <-chan workflow.Event, doneCh chan<- int) {
	exitCode := 0

	for event := range eventsCh {
		switch e := event.(type) {
		case events.EventStoryStarted:
			fmt.Printf("\nStarting story: %s - %s\n", e.Story.ID, e.Story.Title)

		case events.EventStoryCompleted:
			if e.Success {
				fmt.Printf("Completed story: %s - %s\n", e.Story.ID, e.Story.Title)
			} else {
				fmt.Printf("Story failed: %s - %s\n", e.Story.ID, e.Story.Title)
				exitCode = 1
			}

		case events.EventOutput:
			if e.Verbose && !c.verbose {
				continue
			}
			fmt.Printf("%s\n", e.Text)

		case events.EventError:
			fmt.Printf("Error: %v\n", e.Err)
			exitCode = 1

		case events.EventCompleted:
			fmt.Println("\nAll stories completed successfully.")
		}
	}

	doneCh <- exitCode
}

// Run is a convenience function that creates a Cmd and runs it.
func Run(cfg *config.Config, verbose bool) int {
	return NewCmd(cfg, verbose).Run()
}
