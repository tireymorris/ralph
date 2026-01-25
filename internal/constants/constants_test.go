package constants

import (
	"testing"
)

func TestConstantsValues(t *testing.T) {
	// Test that constants have expected values
	tests := []struct {
		name     string
		got      int
		expected int
	}{
		{"EventChannelBuffer", EventChannelBuffer, 10000},
		{"ScannerBufferSize", ScannerBufferSize, 1024 * 1024},
		{"PipeReaderCount", PipeReaderCount, 2},
		{"MaxJSONRepairAttempts", MaxJSONRepairAttempts, 2},
		{"InitialScannerBufferCapacity", InitialScannerBufferCapacity, 64 * 1024},
		{"VerbosePatternMinLength", VerbosePatternMinLength, 4},
		{"VerboseTimestampMinLength", VerboseTimestampMinLength, 10},
		{"TimestampContextLength", TimestampContextLength, 30},
		{"FileLockTimeout", FileLockTimeout, 30},
		{"FileLockRetryDelay", FileLockRetryDelay, 100},
		{"TempFileRandomRange", TempFileRandomRange, 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.expected)
			}
		})
	}
}
