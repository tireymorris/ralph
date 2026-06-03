package app

import (
	"net/http"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/web"
)

func TestRunWebServesHealth(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	done := make(chan int, 1)
	go func() {
		done <- RunWeb(cfg, 0)
	}()

	deadline := time.Now().Add(2 * time.Second)
	var baseURL string
	for time.Now().Before(deadline) {
		baseURL = web.ServerURL()
		if baseURL != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if baseURL == "" {
		t.Fatal("web server did not start")
	}

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	web.Shutdown()
	if code := <-done; code != 0 {
		t.Errorf("RunWeb() = %d, want 0", code)
	}
}
