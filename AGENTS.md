# Agent Patterns

## Buffered Logging
- Logger buffers log messages in memory until flush() is called
- Reduces file I/O operations during execution
- Flush is called on completion of all stories or when errors occur
- Ensures no loss of logging functionality while improving performance