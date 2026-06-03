package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static/dist/*
var staticFS embed.FS

func embeddedDistFS() (fs.FS, error) {
	return fs.Sub(staticFS, "static/dist")
}

func embeddedIndexHTML() ([]byte, error) {
	dist, err := embeddedDistFS()
	if err != nil {
		return nil, err
	}
	data, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		return nil, fmt.Errorf("read index.html: %w", err)
	}
	return data, nil
}

func registerStatic(mux *http.ServeMux) {
	dist, err := embeddedDistFS()
	if err != nil {
		panic(fmt.Sprintf("web static dist: %v", err))
	}
	assetsDir, err := fs.Sub(dist, "assets")
	if err != nil {
		panic(fmt.Sprintf("web static assets dir: %v", err))
	}
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsDir))))
	mux.HandleFunc("GET /{$}", serveIndexHTML(dist))
	mux.HandleFunc("GET /{path...}", spaFallback(dist))
}

func serveIndexHTML(dist fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeIndexHTML(w, dist)
	}
}

func spaFallback(dist fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		writeIndexHTML(w, dist)
	}
}

func writeIndexHTML(w http.ResponseWriter, dist fs.FS) {
	data, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}
