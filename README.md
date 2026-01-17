# Ralph - Autonomous Software Development Agent

Minimalist implementation of the Ralph Wiggum pattern - an autonomous agent that implements software features through a single command.

## Usage

```bash
./ralph "Add user authentication with login and registration"
```

## How It Works

1. **Analysis**: Scans codebase and creates user stories
2. **Implementation**: Implements stories one by one until complete  
3. **Learning**: Each iteration learns from previous work via AGENTS.md  

## Requirements

- opencode CLI tool (https://opencode.ai)
- Git repository

## Architecture

- **Single Command**: One file contains all logic
- **opencode-Powered**: AI intelligence via opencode CLI
- **Git-Native**: Uses git for state management

## Files Created

- `prd.json` - Project state and stories
- `AGENTS.md` - Patterns discovered during development  
- `progress.txt` - Iteration logs

All excluded from git via `.gitignore`.

## References

**Original Pattern:** [Geoffrey Huntley](https://ghuntley.com/)  
**Ralph Philosophy:** [everything is a ralph loop](https://ghuntley.com/loop/)  
**History:** [A brief history of ralph](https://www.humanlayer.dev/blog/brief-history-of-ralph)

## Acknowledgments

This implementation honors Geoffrey Huntley's Ralph Wiggum pattern by focusing on clean iteration boundaries and autonomous operation. The name "Ralph" comes from Ralph Wiggum's persistent, optimistic approach - perfect for autonomous agents.
