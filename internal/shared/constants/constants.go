package constants

const (
	// EventChannelBuffer is sized for burst output from AI runners.
	EventChannelBuffer = 10000

	// ScannerBufferSize allows reading long AI output lines.
	ScannerBufferSize = 1024 * 1024

	// PipeReaderCount covers subprocess stdout and stderr.
	PipeReaderCount = 2

	// InitialScannerBufferCapacity avoids immediate resizing for typical lines.
	InitialScannerBufferCapacity = 64 * 1024

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
)
