# Ralph - Autonomous AI Agent CLI

Ralph automates software development through an autonomous agent loop that implements PRD requirements iteratively until completion.

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

## Autonomous Agent Features

- **Fresh Context**: Each iteration starts with clean state
- **Priority-Based**: Implements stories in priority order
- **Quality Gates**: Runs type checking, tests, and linting
- **Progress Tracking**: Updates `prd.json`, `progress.txt`, and `AGENTS.md`
- **Git Integration**: Automatic branching and committing
- **Learning Persistence**: Documents patterns and discoveries

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