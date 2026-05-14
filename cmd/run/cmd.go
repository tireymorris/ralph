// Package run implements the "ralph run" subcommand (headless/stdout mode).
package run

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"ralph/internal/shared/cli"
	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
	"ralph/internal/prompt"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

type workflowExecutor interface {
	RunClarify(ctx context.Context, userPrompt string) ([]prompt.QuestionAnswer, error)
	RunLoad(ctx context.Context) (*prd.PRD, error)
	RunGenerateWithAnswers(ctx context.Context, userPrompt string, qas []prompt.QuestionAnswer) (*prd.PRD, error)
	RunImplementation(ctx context.Context, p *prd.PRD) error
}

// Cmd runs the headless (stdout/stdin) workflow.
type Cmd struct {
	cfg              *config.Config
	prompt           string
	dryRun           bool
	resume           bool
	verbose          bool
	executor         workflowExecutor
	eventsCh         chan workflow.Event
	reviewResponseCh chan bool
	cancelFunc       context.CancelFunc
}

// NewCmd creates a new Cmd instance.
func NewCmd(cfg *config.Config, userPrompt string, dryRun, resume, verbose bool) *Cmd {
	eventsCh := make(chan workflow.Event, constants.EventChannelBuffer)
	reviewResponseCh := make(chan bool, 1)
	return &Cmd{
		cfg:              cfg,
		prompt:           userPrompt,
		dryRun:           dryRun,
		resume:           resume,
		verbose:          verbose,
		eventsCh:         eventsCh,
		executor:         workflow.NewExecutor(cfg, eventsCh),
		reviewResponseCh: reviewResponseCh,
	}
}

// Run executes the headless workflow and returns an exit code.
func (c *Cmd) Run() int {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel
	defer cancel()

	logger.Debug("run command starting", "prompt", c.prompt, "dry_run", c.dryRun, "resume", c.resume)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Debug("received interrupt signal")
		fmt.Println("\nInterrupted, shutting down...")
		cancel()
	}()

	doneCh := make(chan int, 1)
	go c.handleEvents(c.eventsCh, doneCh)

	var p *prd.PRD
	var err error

	if c.resume {
		p, err = c.executor.RunLoad(ctx)
	} else {
		qas, clarifyErr := c.executor.RunClarify(ctx, c.prompt)
		if clarifyErr != nil {
			logger.Error("clarification step failed", "error", clarifyErr)
			close(c.eventsCh)
			<-doneCh
			return 1
		}
		p, err = c.executor.RunGenerateWithAnswers(ctx, c.prompt, qas)
	}

	if err != nil {
		logger.Error("operation failed", "error", err)
		close(c.eventsCh)
		<-doneCh
		return 1
	}

	if c.dryRun {
		fmt.Println("Dry run complete - PRD saved, no implementation performed")
		close(c.eventsCh)
		<-doneCh
		return 0
	}

	if implErr := c.executor.RunImplementation(ctx, p); implErr != nil {
		logger.Error("implementation failed", "error", implErr)
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
		case events.EventClarifyingQuestions:
			answers := c.collectAnswers(e.Questions)
			e.AnswersCh <- answers

		case events.EventPRDGenerating:
			fmt.Println("Generating PRD...")

		case events.EventPRDGenerated:
			fmt.Printf("PRD generated: %s (%d stories)\n", e.PRD.ProjectName, len(e.PRD.Stories))
			fmt.Printf("Saved to: %s\n\n", c.cfg.PRDFile)
			cli.PrintStoryList(os.Stdout, e.PRD)

		case events.EventPRDLoaded:
			fmt.Printf("Loaded PRD: %s (%d stories, %d completed)\n\n",
				e.PRD.ProjectName, len(e.PRD.Stories), e.PRD.CompletedCount())
			cli.PrintStoryList(os.Stdout, e.PRD)

		case events.EventPRDReview:
			if c.dryRun {
				fmt.Println("PRD ready for review")
				fmt.Println("(Dry run - you can edit prd.json, then run with --resume to proceed)")
				continue
			}
			c.promptPRDReview(e.PRD)
			proceed := <-c.reviewResponseCh
			if !proceed {
				fmt.Println("Review not confirmed, exiting.")
				if c.cancelFunc != nil {
					c.cancelFunc()
				}
				return
			}

		case events.EventStoryStarted:
			fmt.Printf("Story: %s\n", e.Story.Title)

		case events.EventStoryCompleted:
			fmt.Printf("  Completed\n\n")

		case events.EventOutput:
			if e.Verbose && !c.verbose {
				continue
			}
			fmt.Printf("%s %s\n", cli.OutputPrefix(e.IsErr), e.Text)
			os.Stdout.Sync()

		case events.EventError:
			fmt.Printf("Error: %v\n", e.Err)
			exitCode = 1

		case events.EventCompleted:
			fmt.Println("All stories completed successfully!")
			exitCode = 0
		}
	}

	doneCh <- exitCode
}

func (c *Cmd) collectAnswers(questions []string) []prompt.QuestionAnswer {
	fmt.Println()
	fmt.Println("Before generating your PRD, please answer a few clarifying questions.")
	fmt.Println("(Press Enter to skip a question, or Ctrl+C to abort.)")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	qas := make([]prompt.QuestionAnswer, 0, len(questions))

	for i, q := range questions {
		fmt.Printf("Q%d: %s\n> ", i+1, q)
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())
		qas = append(qas, prompt.QuestionAnswer{Question: q, Answer: answer})
	}

	fmt.Println()
	return qas
}

func (c *Cmd) promptPRDReview(p *prd.PRD) {
	fmt.Println("PRD ready for review")
	fmt.Println()
	cli.PrintStoryDetails(os.Stdout, p)
	fmt.Println("Please choose an action:")
	fmt.Println("  1. Edit in editor ($EDITOR)")
	fmt.Println("  2. Proceed without changes")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter choice (1-2): ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		if err := cli.RunEditor(c.cfg.PRDFile); err != nil {
			fmt.Printf("Editor exited with error: %v\n", err)
		}
		fmt.Println("\nEditor closed.")
		fmt.Print("Proceed to implementation? [Y/n]: ")
		scanner.Scan()
		confirm := strings.TrimSpace(scanner.Text())
		c.reviewResponseCh <- confirm == "" || confirm == "y" || confirm == "Y"
	case "2":
		c.reviewResponseCh <- true
	default:
		fmt.Println("Invalid choice, proceeding without changes.")
		c.reviewResponseCh <- true
	}
}

// Run is a convenience function that creates a Cmd and runs it.
func Run(cfg *config.Config, userPrompt string, dryRun, resume, verbose bool) int {
	return NewCmd(cfg, userPrompt, dryRun, resume, verbose).Run()
}
