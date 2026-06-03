package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"ralph/internal/shared/config"
)

var (
	serverMu sync.Mutex
	server   *http.Server
	baseURL  string
)

func listen(cfg *config.Config, addr string) (net.Listener, error) {
	_ = cfg
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", addr, err)
	}
	return ln, nil
}

func setBaseURL(ln net.Listener) error {
	host, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		return err
	}
	if host == "" || host == "::" {
		host = "127.0.0.1"
	}
	serverMu.Lock()
	baseURL = fmt.Sprintf("http://%s:%s", host, port)
	serverMu.Unlock()
	return nil
}

func Start(cfg *config.Config, addr string) error {
	ln, err := listen(cfg, addr)
	if err != nil {
		return err
	}
	if err := setBaseURL(ln); err != nil {
		_ = ln.Close()
		return err
	}
	h, err := NewHandler(cfg)
	if err != nil {
		_ = ln.Close()
		return err
	}
	serverMu.Lock()
	server = &http.Server{Handler: h}
	srv := server
	serverMu.Unlock()
	go func() {
		_ = srv.Serve(ln)
	}()
	return nil
}

func Run(cfg *config.Config, addr string) error {
	ln, err := listen(cfg, addr)
	if err != nil {
		return err
	}
	if err := setBaseURL(ln); err != nil {
		_ = ln.Close()
		return err
	}
	h, err := NewHandler(cfg)
	if err != nil {
		_ = ln.Close()
		return err
	}
	serverMu.Lock()
	server = &http.Server{Handler: h}
	srv := server
	serverMu.Unlock()
	return srv.Serve(ln)
}

func ServerURL() string {
	serverMu.Lock()
	defer serverMu.Unlock()
	return baseURL
}

func Shutdown() {
	serverMu.Lock()
	defer serverMu.Unlock()
	if server == nil {
		return
	}
	_ = server.Shutdown(context.Background())
	server = nil
	baseURL = ""
}
