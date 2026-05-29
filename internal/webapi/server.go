package webapi

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/veirt/vivere/internal/masterdata"
)

type Config struct {
	MasterRoot     string
	ExportsRoot    string
	AllowedOrigins []string
}

type Server struct {
	master         masterdata.Loader
	exportsRoot    string
	allowedOrigins map[string]struct{}
	mux            *http.ServeMux
}

func New(config Config) (*Server, error) {
	if config.MasterRoot == "" {
		return nil, fmt.Errorf("master root is required")
	}
	if config.ExportsRoot == "" {
		return nil, fmt.Errorf("exports root is required")
	}

	allowedOrigins := make(map[string]struct{}, len(config.AllowedOrigins))
	for _, origin := range config.AllowedOrigins {
		if origin != "" {
			allowedOrigins[origin] = struct{}{}
		}
	}

	server := &Server{
		master:         masterdata.New(config.MasterRoot),
		exportsRoot:    config.ExportsRoot,
		allowedOrigins: allowedOrigins,
		mux:            http.NewServeMux(),
	}
	server.routes()
	return server, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		if _, ok := s.allowedOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/assets/") && strings.Contains(filepath.ToSlash(r.URL.Path), "/../") {
		http.NotFound(w, r)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	api := NewAPI(s.master, s.exportsRoot)
	api.Routes(s.mux)
	s.mux.HandleFunc("GET /assets/", s.handleAsset)
}

func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	assetPath := strings.TrimPrefix(r.URL.Path, "/assets/")
	if assetPath == "" || strings.ContainsAny(assetPath, `\\`) {
		http.NotFound(w, r)
		return
	}
	cleanAssetPath := filepath.Clean(assetPath)
	if cleanAssetPath == "." || strings.HasPrefix(cleanAssetPath, "..") || filepath.IsAbs(cleanAssetPath) {
		http.NotFound(w, r)
		return
	}

	root, err := filepath.Abs(s.exportsRoot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fullPath := filepath.Join(root, cleanAssetPath)
	rel, err := filepath.Rel(root, fullPath)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, fullPath)
}
