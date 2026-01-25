package constants

// Common constants used across the Ralph codebase.

const (
	// EventChannelBuffer controls how many events can be queued before blocking.
	// Set to 10000 to handle burst output from AI runners.
	EventChannelBuffer = 10000

	// ScannerBufferSize is maximum line size for reading AI output.
	// Set to 1MB to handle very long output lines.
	ScannerBufferSize = 1024 * 1024

	// PipeReaderCount is the number of goroutines to spawn for reading
	// subprocess stdout and stderr pipes (one for stdout, one for stderr).
	PipeReaderCount = 2

	// MaxJSONRepairAttempts is how many times we'll try to fix corrupted PRD JSON
	// before giving up.
	MaxJSONRepairAttempts = 2

	// InitialScannerBufferCapacity is the initial capacity for scanner buffers.
	// Set to 64KB to handle moderate line sizes without immediate resizing.
	InitialScannerBufferCapacity = 64 * 1024

	// VerbosePatternMinLength is the minimum line length to check for verbose patterns.
	// Prevents unnecessary pattern matching on very short lines.
	VerbosePatternMinLength = 4

	// VerboseTimestampMinLength is the minimum line length to check for timestamp patterns.
	// Ensures line is long enough to contain a timestamp format.
	VerboseTimestampMinLength = 10

	// TimestampContextLength is the maximum substring length to examine when looking
	// for timestamp patterns in verbose lines.
	TimestampContextLength = 30

	// FileLockTimeout is the timeout for acquiring file locks.
	// Set to 30 seconds to handle temporary contention without hanging indefinitely.
	FileLockTimeout = 30

	// FileLockRetryDelay is the delay between file lock acquisition attempts.
	// Set to 100ms to allow rapid retries without overwhelming the system.
	FileLockRetryDelay = 100

	// TempFileRandomRange is the range for random numbers in temporary filenames.
	// Set to 100000 to provide sufficient entropy for temporary file names.
	TempFileRandomRange = 100000
)
