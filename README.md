# Ralph 

Ralph implements the Ralph Wiggum pattern - autonomous agents that learn and improve through clean iteration boundaries.

**Original Pattern**: [Geoffrey Huntley](https://ghuntley.com/)  
**Ralph Philosophy**: [everything is a ralph loop](https://ghuntley.com/loop/)  
**History**: [A brief history of ralph](https://www.humanlayer.dev/blog/brief-history-of-ralph)

Ralph transforms natural language requirements into working code through autonomous user story implementation.

## Usage

```bash
# Full implementation
./ralph "Add user authentication with login and registration"

# Generate PRD for review
./ralph "Add user authentication" --dry-run
```

## Generated Files

- `prd.json` - User stories and project state (cleaned up on completion)

