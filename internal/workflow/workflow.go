package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ralph/internal/config"
	"ralph/internal/constants"
	"ralph/internal/logger"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

type Output struct {
	Text    string
	IsErr   bool
	Verbose bool
}

type Event interface {
	isEvent()
}

type EventPRDGenerating struct{}

func (EventPRDGenerating) isEvent() {}

type EventPRDGenerated struct {
	PRD *prd.PRD
}

func (EventPRDGenerated) isEvent() {}

type EventPRDLoaded struct {
	PRD *prd.PRD
}

func (EventPRDLoaded) isEvent() {}

type EventStoryStarted struct {
	Story     *prd.Story
	Iteration int
}

func (EventStoryStarted) isEvent() {}

type EventStoryCompleted struct {
	Story   *prd.Story
	Success bool
}

func (EventStoryCompleted) isEvent() {}

type EventOutput struct {
	Output
}

func (EventOutput) isEvent() {}

type EventError struct {
	Err error
}

func (EventError) isEvent() {}

type EventCompleted struct{}

func (EventCompleted) isEvent() {}

type EventFailed struct {
	FailedStories []*prd.Story
}

func (EventFailed) isEvent() {}

type Executor struct {
	cfg      *config.Config
	eventsCh chan Event
	runner   runner.RunnerInterface
}

func NewExecutor(cfg *config.Config, eventsCh chan Event) *Executor {
	return &Executor{
		cfg:      cfg,
		eventsCh: eventsCh,
		runner:   runner.New(cfg),
	}
}

func NewExecutorWithRunner(cfg *config.Config, eventsCh chan Event, r runner.RunnerInterface) *Executor {
	return &Executor{
		cfg:      cfg,
		eventsCh: eventsCh,
		runner:   r,
	}
}

func (e *Executor) RunGenerate(ctx context.Context, userPrompt string) (*prd.PRD, error) {
	logger.Debug("generating PRD", "prompt_length", len(userPrompt))
	e.emit(EventPRDGenerating{})

	isEmpty := isEmptyCodebase(e.cfg.WorkDir)
	if isEmpty {
		logger.Info("working directory has no source code, treating as new project", "work_dir", e.cfg.WorkDir)
		e.emit(EventOutput{Output{Text: "Warning: Working directory appears to have no source code. PRD will be generated for a new project."}})
	}

	e.emit(EventOutput{Output{Text: "Analyzing codebase and generating PRD..."}})

	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	go e.forwardOutput(outputCh)

	prdPrompt := prompt.PRDGeneration(userPrompt, e.cfg.PRDFile, "feature", isEmpty)
	err := e.runner.Run(ctx, prdPrompt, outputCh)
	close(outputCh)

	if err != nil {
		logger.Error("PRD generation failed", "error", err)
		e.emit(EventError{Err: fmt.Errorf("PRD generation failed with model %s: %w", e.cfg.Model, err)})
		return nil, fmt.Errorf("PRD generation failed with model %s: %w", e.cfg.Model, err)
	}

	if !prd.Exists(e.cfg) {
		err := fmt.Errorf("AI completed but did not generate %s — it may not have understood the request", e.cfg.PRDFile)
		logger.Error("AI did not generate PRD file", "file", e.cfg.PRDFile)
		e.emit(EventError{Err: err})
		return nil, err
	}

	p, err := prd.Load(e.cfg)
	if err != nil {
		logger.Error("failed to load generated PRD", "error", err)
		e.emit(EventError{Err: fmt.Errorf("failed to load generated PRD %s: %w", e.cfg.PRDFile, err)})
		return nil, fmt.Errorf("failed to load generated PRD %s: %w", e.cfg.PRDFile, err)
	}

	// Validate and improve PRD until actionable
	p, err = e.validateAndImprovePRD(ctx, p)
	if err != nil {
		logger.Warn("PRD validation failed, proceeding with original", "error", err)
	}

	logger.Debug("PRD generated and validated", "project", p.ProjectName, "stories", len(p.Stories))
	e.emit(EventPRDGenerated{PRD: p})
	return p, nil
}

func (e *Executor) RunLoad(ctx context.Context) (*prd.PRD, error) {
	p, err := prd.Load(e.cfg)
	if err != nil {
		e.emit(EventError{Err: fmt.Errorf("failed to load PRD %s: %w", e.cfg.PRDFile, err)})
		return nil, fmt.Errorf("failed to load PRD %s: %w", e.cfg.PRDFile, err)
	}

	logger.Debug("PRD loaded", "project", p.ProjectName, "stories", len(p.Stories))
	e.emit(EventPRDLoaded{PRD: p})
	return p, nil
}

func (e *Executor) RunImplementation(ctx context.Context, p *prd.PRD) error {
	logger.Debug("starting implementation",
		"project", p.ProjectName,
		"branch", p.BranchName,
		"total_stories", len(p.Stories),
		"completed", p.CompletedCount())

	iteration := 0

	for {
		select {
		case <-ctx.Done():
			logger.Debug("context cancelled")
			return ctx.Err()
		default:
		}

		p, err := prd.Load(e.cfg)
		if err != nil {
			logger.Error("failed to reload PRD", "error", err)
			wrappedErr := fmt.Errorf("failed to reload PRD %s: %w", e.cfg.PRDFile, err)
			e.emit(EventError{Err: fmt.Errorf("cannot continue without PRD: %w", wrappedErr)})
			return wrappedErr
		}

		if p.AllCompleted() {
			logger.Info("all stories completed successfully")
			prd.Delete(e.cfg)
			e.emit(EventCompleted{})
			return nil
		}

		next := p.NextPendingStory(e.cfg.RetryAttempts)
		if next == nil {
			failed := p.FailedStories(e.cfg.RetryAttempts)
			logger.Error("all remaining stories have failed", "failed_count", len(failed))
			e.emit(EventFailed{FailedStories: failed})
			return fmt.Errorf("all remaining stories have failed (%d stories)", len(failed))
		}

		iteration++
		if iteration > e.cfg.MaxIterations {
			logger.Error("max iterations reached", "iterations", iteration, "max", e.cfg.MaxIterations)
			e.emit(EventFailed{FailedStories: p.FailedStories(e.cfg.RetryAttempts)})
			return fmt.Errorf("max iterations (%d) reached after %d iterations", e.cfg.MaxIterations, iteration)
		}

		logger.Debug("starting story",
			"story_id", next.ID,
			"title", next.Title,
			"iteration", iteration,
			"retry_count", next.RetryCount)

		e.emit(EventStoryStarted{Story: next, Iteration: iteration})

		outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
		go e.forwardOutput(outputCh)

		storyPrompt := prompt.StoryImplementation(
			next.ID,
			next.Title,
			next.Description,
			next.AcceptanceCriteria,
			p.TestSpec,
			p.Context,
			e.cfg.PRDFile,
			iteration,
			p.CompletedCount(),
			len(p.Stories),
		)

		err = e.runner.Run(ctx, storyPrompt, outputCh)
		close(outputCh)

		if err != nil {
			logger.Debug("AI runner returned error", "story_id", next.ID, "model", e.cfg.Model, "error", err)
		}

		updatedPRD, loadErr := prd.Load(e.cfg)
		if loadErr != nil {
			logger.Error("failed to reload PRD after story, cannot continue", "error", loadErr, "story_id", next.ID)
			wrappedErr := fmt.Errorf("failed to reload PRD %s after story %s: %w", e.cfg.PRDFile, next.ID, loadErr)
			e.emit(EventError{Err: wrappedErr})
			return wrappedErr
		}

		// Check for version conflicts (unexpected jumps indicate concurrent modification)
		if p.Version > 0 && updatedPRD.Version > p.Version+1 {
			logger.Warn("PRD version jumped unexpectedly",
				"previous", p.Version,
				"current", updatedPRD.Version,
				"expected", p.Version+1,
				"story_id", next.ID)
			e.emit(EventOutput{Output{Text: fmt.Sprintf(
				"Warning: PRD was modified externally (version %d → %d)", p.Version, updatedPRD.Version)}})
		}

		updatedStory := updatedPRD.GetStory(next.ID)
		if updatedStory != nil && updatedStory.Passes {
			logger.Debug("story marked as completed", "story_id", next.ID)
			e.emit(EventStoryCompleted{Story: updatedStory, Success: true})
		} else {
			logger.Debug("story not completed", "story_id", next.ID)
			if updatedStory != nil && updatedStory.RetryCount == next.RetryCount {
				updatedStory.RetryCount++
				if saveErr := prd.Save(e.cfg, updatedPRD); saveErr != nil {
					logger.Warn("failed to save retry count", "error", saveErr, "story_id", next.ID)
				}
			}
			e.emit(EventStoryCompleted{Story: next, Success: false})
		}

		p = updatedPRD
	}
}

func (e *Executor) emit(event Event) {
	if e.eventsCh != nil {
		select {
		case e.eventsCh <- event:
		default:
			logger.Warn("event channel full, dropping event", "event_type", fmt.Sprintf("%T", event))
		}
	}
}

func (e *Executor) forwardOutput(outputCh <-chan runner.OutputLine) {
	for line := range outputCh {
		e.emit(EventOutput{Output{Text: line.Text, IsErr: line.IsErr, Verbose: line.Verbose}})
	}
}

func (e *Executor) validateAndImprovePRD(ctx context.Context, p *prd.PRD) (*prd.PRD, error) {
	const maxValidationIterations = 3

	for iteration := 0; iteration < maxValidationIterations; iteration++ {
		if e.isPRDActionable(p) {
			logger.Debug("PRD is actionable", "iteration", iteration)
			return p, nil
		}

		logger.Debug("improving PRD", "iteration", iteration+1)
		e.emit(EventOutput{Output{Text: fmt.Sprintf("Improving PRD for actionability (iteration %d)...", iteration+1)}})

		outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
		go e.forwardOutput(outputCh)

		validationPrompt := prompt.PRDValidation(p.ToJSON(), e.cfg.PRDFile, p.Context)
		err := e.runner.Run(ctx, validationPrompt, outputCh)
		close(outputCh)

		if err != nil {
			if ctx.Err() != nil {
				return p, ctx.Err()
			}
			logger.Warn("PRD validation attempt failed", "iteration", iteration+1, "error", err)
			continue
		}

		updatedPRD, loadErr := prd.Load(e.cfg)
		if loadErr == nil {
			p = updatedPRD
			logger.Debug("PRD improved", "iteration", iteration+1)
		} else {
			logger.Warn("failed to load improved PRD", "iteration", iteration+1, "error", loadErr)
		}
	}

	if !e.isPRDActionable(p) {
		logger.Warn("PRD still not fully actionable after validation, proceeding anyway")
		e.emit(EventOutput{Output{Text: "Warning: PRD may contain vague requirements, proceeding with best effort."}})
	}

	return p, nil
}

// isPRDActionable checks for clearly vague story descriptions and acceptance
// criteria that lack any quantification. This is intentionally conservative —
// it only flags stories that use vague terms with zero quantifying context.
func (e *Executor) isPRDActionable(p *prd.PRD) bool {
	vagueVerbs := []string{"simplify", "optimize", "reduce", "improve", "enhance", "streamline", "refactor"}
	vagueAdjectives := []string{"proper", "appropriate", "comprehensive", "good", "correct", "consistent", "clean", "robust"}
	quantifiers := []string{"%", "lines", "words", "bytes", "functions", "from", "to", "remove", "delete", "replace", "rename", "move", "extract", "inline", "split", "merge"}

	for _, story := range p.Stories {
		if hasVagueTerms(story.Description, vagueVerbs, quantifiers) {
			return false
		}
		for _, ac := range story.AcceptanceCriteria {
			if hasVagueTerms(ac, vagueVerbs, quantifiers) {
				return false
			}
			if hasVagueTerms(ac, vagueAdjectives, quantifiers) {
				return false
			}
		}
	}
	return true
}

// hasVagueTerms checks if text contains any of the vague terms without
// any quantifying context.
func hasVagueTerms(text string, vagueTerms, quantifiers []string) bool {
	lower := strings.ToLower(text)
	for _, term := range vagueTerms {
		if !strings.Contains(lower, term) {
			continue
		}
		hasQuantification := false
		for _, q := range quantifiers {
			if strings.Contains(lower, q) {
				hasQuantification = true
				break
			}
		}
		if !hasQuantification {
			return true
		}
	}
	return false
}

// isEmptyCodebase checks whether the working directory contains any source
// code files. Returns true if no files with common source code extensions
// are found (skipping hidden directories).
func isEmptyCodebase(workDir string) bool {
	if workDir == "" {
		return true
	}

	sourceExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".rb": true, ".java": true, ".rs": true, ".c": true, ".cpp": true, ".cs": true,
		".php": true, ".swift": true, ".kt": true, ".ex": true, ".hs": true, ".scala": true,
		".sh": true, ".ml": true, ".r": true, ".pl": true, ".lua": true, ".dart": true,
		".vue": true, ".svelte": true, ".html": true, ".css": true, ".scss": true,
	}

	found := false
	filepath.WalkDir(workDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable dirs
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if sourceExts[ext] {
			found = true
			return filepath.SkipAll
		}
		return nil
	})

	return !found
}
