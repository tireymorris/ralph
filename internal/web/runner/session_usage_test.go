package runner

import (
	"os"
	"strings"
	"testing"
)

func TestRunControllerUsesSessionFacadeForWorkflowStarts(t *testing.T) {
	files := []string{"controller.go", "resume.go"}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", file, err)
		}
		if strings.Contains(string(src), "c.Driver.Start") {
			t.Fatalf("%s calls workflow driver start methods directly; use session facade methods", file)
		}
	}
}
