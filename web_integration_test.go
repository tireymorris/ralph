//go:build integration

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func startWebServer(t *testing.T, workDir string) (baseURL string, stop func()) {
	t.Helper()
	binaryPath := buildTestBinary(t)

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, binaryPath, "web", "--port", "0")
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "RALPH_RUNNER=mock")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start web server: %v", err)
	}

	urlCh := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			line := sc.Text()
			const prefix = "ralph web listening on "
			if strings.HasPrefix(line, prefix) {
				urlCh <- strings.TrimSpace(strings.TrimPrefix(line, prefix))
				return
			}
		}
	}()

	var listenURL string
	deadline := time.Now().Add(10 * time.Second)
	for listenURL == "" && time.Now().Before(deadline) {
		select {
		case u := <-urlCh:
			listenURL = u
		default:
			time.Sleep(20 * time.Millisecond)
		}
	}
	if listenURL == "" {
		cancel()
		_ = cmd.Wait()
		t.Fatal("timed out waiting for web server listen URL on stdout")
	}

	healthDeadline := time.Now().Add(5 * time.Second)
	client := &http.Client{Timeout: 2 * time.Second}
	healthy := false
	for time.Now().Before(healthDeadline) {
		resp, err := client.Get(listenURL + "/health")
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				healthy = true
				_ = resp.Body.Close()
				break
			}
			_ = resp.Body.Close()
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !healthy {
		cancel()
		_ = cmd.Wait()
		t.Fatalf("/health did not return 200 within 5s of server start (baseURL=%s)", listenURL)
	}

	stop = func() {
		cancel()
		_ = cmd.Wait()
	}
	return listenURL, stop
}

func TestWebIntegrationHealth(t *testing.T) {
	workDir := t.TempDir()
	initGitRepo(t, workDir)

	baseURL, stop := startWebServer(t, workDir)
	defer stop()

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /health status = %d, want 200", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode health JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("health status = %q, want ok", body["status"])
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
}

func TestWebIntegrationCreateRunSSE(t *testing.T) {
	workDir := t.TempDir()
	initGitRepo(t, workDir)

	baseURL, stop := startWebServer(t, workDir)
	defer stop()

	body := []byte(`{"prompt":"integration test goal"}`)
	resp, err := http.Post(baseURL+"/api/runs", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/runs: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST /api/runs status = %d, want 201, body: %s", resp.StatusCode, b)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.ID == "" {
		t.Fatal("create response missing id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/runs/"+created.ID+"/events", nil)
	if err != nil {
		t.Fatalf("new SSE request: %v", err)
	}
	sseResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET events: %v", err)
	}
	defer sseResp.Body.Close()
	if sseResp.StatusCode != http.StatusOK {
		t.Fatalf("GET events status = %d, want 200", sseResp.StatusCode)
	}

	sc := bufio.NewScanner(sseResp.Body)
	dataLines := 0
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		dataLines++
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		var envelope struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
			t.Fatalf("decode SSE event: %v", err)
		}
		if envelope.Type != "" {
			break
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("read SSE stream: %v", err)
	}
	if dataLines == 0 {
		t.Fatal("expected at least one SSE data event within 30s, got none")
	}
}
