# Ralph - Autonomous AI Agent CLI

Ralph automates software development through an autonomous agent loop that implements PRD requirements iteratively until completion.

## About

This is a Ruby implementation of the **Ralph Wiggum pattern** originally created by [Geoffrey Huntley](https://ghuntley.com/). The Ralph pattern is a simple but powerful concept: run an AI coding agent repeatedly until all tasks are complete, starting fresh with each iteration.

**Original Concept:** [Geoffrey Huntley](https://ghuntley.com/)  
**Learn More:** [everything is a ralph loop](https://ghuntley.com/loop/)

## Quick Start

```bash
./ralph init                    # Initialize project structure
./ralph prd:create "Build user auth"  # Create PRD from description  
./ralph prd:convert tasks/prd-*.md       # Convert to JSON format
./ralph run                      # Run autonomously until completion
./ralph run 10                  # Run with max 10 iterations
./ralph status                    # Show progress
./ralph debug                    # Show debug info
```

## Commands

### Project Setup
- `ralph init` - Initialize Ralph project with directories and template files
- `ralph prd:create "description"` - Create PRD markdown from feature description
- `ralph prd:convert <file.md>` - Convert PRD to structured JSON

### Autonomous Development
- `ralph run` - Run autonomous agent until all stories complete (default)
- `ralph run <number>` - Run with maximum iteration limit
- `ralph status` - Show current progress and completion percentage
- `ralph debug` - Full debugging information and state details

### Interactive Mode
- `ralph -i` - Interactive REPL with autocomplete
- `ralph --help` - Show all available commands

## Workflow

1. **Initialize**: `ralph init` creates project structure
2. **Plan**: `ralph prd:create "feature description"` generates requirements
3. **Convert**: `ralph prd:convert tasks/prd-*.md` creates actionable stories
4. **Execute**: `ralph run` starts autonomous development loop
5. **Monitor**: `ralph status` tracks progress

## Ralph Pattern Features

- **LLM-Agnostic**: Works with OpenAI, Anthropic, or Ollama models
- **Fresh Context**: Each iteration spawns a new agent instance with clean state (core Ralph principle)
- **Intelligent Quality Gates**: LLM determines appropriate project-specific quality checks
- **Smart Story Parsing**: LLM extracts and prioritizes user stories from PRDs
- **Priority-Based**: Implements stories in priority order
- **Progress Tracking**: Updates `prd.json`, `progress.txt`, and `AGENTS.md`
- **Git Integration**: Automatic branching and committing
- **Learning Persistence**: Documents patterns and discoveries for future iterations
- **Deterministic Context Allocation**: Avoids context rot through clean iteration boundaries

## Current Implementation Status

✅ **Implemented:**
- Full CLI with command registry
- Interactive mode and help system
- Project structure initialization
- PRD creation and conversion (basic)
- Autonomous agent loop framework
- Progress tracking and git integration
- AGENTS.md learning persistence

⚠️ **Requirements:**
- **LLM Provider**: OpenAI, Anthropic, or Ollama
- **API Keys**: Set `OPENAI_API_KEY` or `ANTHROPIC_API_KEY` environment variables
- **Ruby Gems**: `gem install openai anthropic` (if using those providers)

## LLM Configuration

Configure your LLM provider via environment variables:

```bash
# OpenAI (default)
export RALPH_LLM_PROVIDER=openai
export RALPH_LLM_MODEL=gpt-4
export OPENAI_API_KEY=your_key_here

# Anthropic Claude
export RALPH_LLM_PROVIDER=anthropic
export RALPH_LLM_MODEL=claude-3-sonnet-20241022
export ANTHROPIC_API_KEY=your_key_here

# Local Ollama
export RALPH_LLM_PROVIDER=ollama
export RALPH_LLM_MODEL=codellama
```

## How It Works

1. **LLM Integration**: Ralph uses any configured LLM for:
   - Extracting user stories from PRD documents
   - Determining project-appropriate quality checks  
   - Implementing code based on story requirements
   - Learning from previous iterations

2. **Project Discovery**: The LLM scans your project to understand:
   - Technology stack (package.json, requirements.txt, etc.)
   - Testing frameworks and commands
   - Build and lint configurations
   - Code patterns and conventions

3. **Iterative Implementation**: Each iteration:
   - Spawns a fresh LLM context (no conversation accumulation)
   - Implements one user story completely
   - Runs project-specific quality checks
   - Commits if all checks pass
   - Updates progress and learning files

## Project Structure

```
project/
├── prd.json           # Structured requirements with story status
├── progress.txt        # Iteration logs and learnings  
├── AGENTS.md          # Patterns discovered during development
├── tasks/             # PRD markdown files
├── archive/           # Completed feature archives
└── scripts/ralph/     # Ralph automation scripts
```

## Example Usage

```bash
# Start a new feature
./ralph init
./ralph prd:create "Add user authentication with login and registration"

# Edit the generated PRD file with detailed requirements
# vim tasks/prd-20240117-143022.md

# Convert to actionable stories
./ralph prd:convert tasks/prd-20240117-143022.md

# Run autonomously (will continue until all stories pass)
./ralph run

# Or with iteration limit for testing
./ralph run 5

# Check progress anytime
./ralph status
```

## Adding Custom Commands

Create files in `commands/` that register commands:

```ruby
Ralph::Registry.register('myapp:greet', 'Say hello') do |name = 'World'|
  puts "Hello, #{name}!"
end
```

Commands are auto-loaded from `commands/*.rb`.

## File Exclusions

Generated files are excluded from git via `.gitignore`:
- `tasks/` - PRD markdown files
- `prd.json` - Current project requirements
- `progress.txt` - Iteration logs
- `AGENTS.md` - Learning documentation
- `archive/` - Completed feature archives

Only the Ralph CLI framework code is version controlled.

## References

- **Original Pattern:** [Geoffrey Huntley](https://ghuntley.com/)
- **Ralph Philosophy:** [everything is a ralph loop](https://ghuntley.com/loop/)
- **History:** [A brief history of ralph](https://www.humanlayer.dev/blog/brief-history-of-ralph)
- **Technical Deep Dive:** [Ralph Wiggum, explained](https://jpcaparas.medium.com/ralph-wiggum-explained-the-claude-code-loop-that-keeps-going-3250dcc30809)

## Acknowledgments

This implementation is based on the Ralph Wiggum pattern by [Geoffrey Huntley](https://ghuntley.com/). The pattern fundamentally changed how we approach autonomous AI development by recognizing that clean iteration boundaries are more effective than accumulating conversation history.

The name "Ralph" comes from Ralph Wiggum of The Simpsons, known for his persistent and optimistic approach - a perfect metaphor for autonomous agents that keep trying until they succeed.