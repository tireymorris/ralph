package web

import (
	"strings"
	"testing"
)

func TestEmbeddedIndex(t *testing.T) {
	t.Helper()
	data, err := embeddedIndexHTML()
	if err != nil {
		t.Fatalf("embeddedIndexHTML: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("embedded index.html is empty")
	}
	body := string(data)
	if !strings.Contains(body, "/assets/") {
		t.Errorf("index.html should reference /assets/, got:\n%s", body)
	}
}
