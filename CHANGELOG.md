# Changelog

All notable changes to Ralph will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **CRITICAL**: Fixed race condition bug in PRD file handling that caused data corruption and lost updates
  - Implemented atomic file writes (write to temp, atomic rename)
  - Added file locking using github.com/gofrs/flock (shared locks for reads, exclusive for writes)
  - Added optimistic locking with version field to detect concurrent modifications
  - Fixed silent error handling in workflow that continued with stale data
  - All file operations now properly synchronized and error-checked
  - See commit `f3f4a4e` for details

### Changed
- PRD file permissions changed from 0644 to 0600 (user-only access) for better security
- Workflow now fails fast on PRD reload errors instead of silently continuing with stale data
- Reverted default model back to `opencode/big-pickle`

### Added
- `Version` field to PRD struct for optimistic locking and conflict detection
- Comprehensive test suite for PRD storage with race condition detection
- `LockTimeoutError` and `VersionConflictError` types for better error handling
- Version conflict warnings in workflow when external modifications detected
- Backward compatibility for old PRD files without version field (defaults to 0)

### Fixed (CLI)
- Corrected Claude CLI command name from `claude-code` to `claude`
- Updated Claude model names to use valid aliases (`sonnet`, `haiku`, `opus`)

## [0.1.0] - 2026-01-23

### Added
- Initial release of Ralph
- PRD generation from natural language prompts
- Iterative story-based implementation
- Interactive TUI mode using Charmbracelet Bubbletea
- Headless CLI mode for CI/CD pipelines
- Support for OpenCode and Claude Code AI backends
- Git integration for commits and branch management
- Configurable retry attempts and iteration limits
