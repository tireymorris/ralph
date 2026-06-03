package web

import (
	"strings"
	"testing"
)

func TestListenAddrLoopbackOnly(t *testing.T) {
	addr := ListenAddr(8080)
	if strings.HasPrefix(addr, "0.0.0.0:") {
		t.Fatalf("ListenAddr binds all interfaces: %q", addr)
	}
	host, _, ok := strings.Cut(addr, ":")
	if !ok {
		t.Fatalf("ListenAddr = %q, want host:port", addr)
	}
	if host != "127.0.0.1" && host != "localhost" {
		t.Errorf("host = %q, want 127.0.0.1 or localhost", host)
	}
}

func TestListenAddrEphemeralPort(t *testing.T) {
	addr := ListenAddr(0)
	if addr != "127.0.0.1:0" {
		t.Fatalf("ListenAddr(0) = %q, want 127.0.0.1:0", addr)
	}
}
