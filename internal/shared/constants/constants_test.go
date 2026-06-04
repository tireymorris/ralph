package constants

import "testing"

func TestCleanupPassCountEqualsThree(t *testing.T) {
	if CleanupPassCount != 3 {
		t.Fatalf("CleanupPassCount = %d, want 3", CleanupPassCount)
	}
}
