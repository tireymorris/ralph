package errors

import "fmt"

// PRDError represents errors related to PRD operations
type PRDError struct {
	Op  string
	Err error
}

func (e PRDError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("PRD %s failed", e.Op)
	}
	return fmt.Sprintf("PRD %s failed: %v", e.Op, e.Err)
}

func (e PRDError) Unwrap() error { return e.Err }

// OpencodeError represents errors related to opencode execution
type OpencodeError struct {
	Op  string
	Err error
}

func (e OpencodeError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("opencode %s failed", e.Op)
	}
	return fmt.Sprintf("opencode %s failed: %v", e.Op, e.Err)
}

func (e OpencodeError) Unwrap() error { return e.Err }

// GitError represents errors related to git operations
type GitError struct {
	Op  string
	Err error
}

func (e GitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("git %s failed", e.Op)
	}
	return fmt.Sprintf("git %s failed: %v", e.Op, e.Err)
}

func (e GitError) Unwrap() error { return e.Err }
