package workflow

import "testing"

func TestCleanupEventAliases(t *testing.T) {
	_ = EventCleanupStarted{}
	_ = EventCleanupCompleted{}
}
