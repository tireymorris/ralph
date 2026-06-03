package web

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
)

func TestStartServesHealth(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	done := make(chan struct{})
	go func() {
		_ = Start(cfg, "127.0.0.1:0")
		close(done)
	}()

	var baseURL string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		baseURL = ServerURL()
		if baseURL != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if baseURL == "" {
		t.Fatal("server did not become ready")
	}
	defer func() {
		Shutdown()
		<-done
	}()

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !strings.Contains(string(body), "ok") {
		t.Errorf("body = %q, want JSON containing ok", body)
	}
}
