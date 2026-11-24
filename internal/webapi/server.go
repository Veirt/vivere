package webapi

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/cors"
	"github.com/veirt/vivere/internal/masterdata"
)

type Config struct {
	MasterRoot     string
	ExportsRoot    string
	AllowedOrigins []string
}

type Server struct {
	master      masterdata.Loader
	exportsRoot string
	mux         *http.ServeMux
	handler     http.Handler
}

func New(config Config) (*Server, error) {
	if config.MasterRoot == "" {
		return nil, fmt.Errorf("master root is required")
	}
	if config.ExportsRoot == "" {
		return nil, fmt.Errorf("exports root is required")
	}

	allowedOrigins := make([]string, 0, len(config.AllowedOrigins))
	for _, origin := range config.AllowedOrigins {
		if origin != "" {
			allowedOrigins = append(allowedOrigins, origin)
		}
	}

	server := &Server{
		master:      masterdata.New(config.MasterRoot),
		exportsRoot: config.ExportsRoot,
		mux:         http.NewServeMux(),
	}
	server.routes()

	corsOpts := cors.Options{
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
		AllowedHeaders: []string{"Content-Type"},
	}
	if len(allowedOrigins) > 0 {
		corsOpts.AllowedOrigins = allowedOrigins
	} else {
		// Prevent rs/cors from falling back to wildcard "*" when no origins are configured.
		corsOpts.AllowOriginFunc = func(string) bool { return false }
	}
	server.handler = cors.New(corsOpts).Handler(server.mux)
	return server, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/assets/") && strings.Contains(filepath.ToSlash(r.URL.Path), "/../") {
		http.NotFound(w, r)
		return
	}
	s.handler.ServeHTTP(w, r)
}

func (s *Server) routes() {
	api := NewAPI(s.master, s.exportsRoot)
	api.Routes(s.mux)
	s.mux.HandleFunc("GET /assets/", s.handleAsset)
}

func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	assetPath := strings.TrimPrefix(r.URL.Path, "/assets/")
	if assetPath == "" || strings.ContainsAny(assetPath, `\`) {
		http.NotFound(w, r)
		return
	}
	cleanAssetPath := filepath.Clean(assetPath)
	if cleanAssetPath == "." || strings.HasPrefix(cleanAssetPath, "..") || filepath.IsAbs(cleanAssetPath) {
		http.NotFound(w, r)
		return
	}

	root, err := os.OpenRoot(s.exportsRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer root.Close()

	info, err := root.Stat(cleanAssetPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	http.ServeFileFS(w, r, root.FS(), filepath.ToSlash(cleanAssetPath))
}
