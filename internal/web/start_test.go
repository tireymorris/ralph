package web

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
)

func TestRunServesHealth(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	done := make(chan error, 1)
	go func() {
		done <- Run(cfg, addr)
	}()

	var base string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		base = ServerURL()
		if base != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if base == "" {
		t.Fatal("server did not become ready")
	}
	defer Shutdown()

	resp, err := http.Get(base + "/health")
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
