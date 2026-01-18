# Agent Patterns

## Buffered Logging
- Logger buffers log messages in memory until flush() is called
- Reduces file I/O operations during execution
- Flush is called on completion of all stories or when errors occur
- Ensures no loss of logging functionality while improving performance

## Optimized Git Status Checks
- Use `git diff --quiet` and `git diff --staged --quiet` instead of `git status --porcelain` for change detection
- Avoids parsing output by checking exit codes directly
- More efficient as it can exit early without listing all files
- Maintains same functionality for detecting staged and unstaged changes