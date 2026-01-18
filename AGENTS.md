# Ralph Agent Patterns

## Minimize Implementation Console Output

**Pattern:** Remove non-essential puts statements from implementation code to reduce execution time spent on printing.

**Context:** During story implementation, verbose console output from command execution was slowing down the process.

**Solution:**
- Remove puts statements for streaming stdout/stderr during opencode command execution
- Remove puts statements for prompt length and command execution details
- Remove puts statements for output summary information
- Keep critical progress messages (completion status, success/failure indicators)
- Keep critical error messages and backtraces

**Benefits:**
- Faster execution by reducing I/O operations
- Cleaner console output focused on essential progress
- Maintained visibility of critical events and errors

**Files Modified:**
- `lib/ralph/error_handler.rb` - Removed verbose puts from `capture_command_output` method