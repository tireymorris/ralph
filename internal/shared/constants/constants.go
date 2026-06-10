package constants

const (
	// EventChannelBuffer is sized for burst output from AI runners.
	EventChannelBuffer = 10000

	// PipeReaderCount covers subprocess stdout and stderr.
	PipeReaderCount = 2

	// PipeReaderBufferSize avoids immediate resizing for typical lines.
	PipeReaderBufferSize = 64 * 1024

	// VerbosePatternMinLength is the minimum length needed before checking verbose prefixes.
	VerbosePatternMinLength = 4

	// VerboseTimestampMinLength is the minimum length needed before checking timestamp prefixes.
	VerboseTimestampMinLength = 10

	// TimestampContextLength bounds the substring inspected for timestamp patterns.
	TimestampContextLength = 30

	// FileLockTimeout is in seconds.
	FileLockTimeout = 30

	// FileLockRetryDelay is in milliseconds.
	FileLockRetryDelay = 100

	// TempFileRandomRange provides entropy for temporary file names.
	TempFileRandomRange = 100000

	// MaxPRDSelfReviewRounds caps agent self-review rounds after PRD generation.
	MaxPRDSelfReviewRounds = 3
)
