package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/veirt/vivere/internal/webapi"
)

func main() {
	var (
		root        = flag.String("root", "/mnt/storage3/vivere", "vivere data root containing master/ and exports/")
		masterRoot  = flag.String("master", "", "master data directory; defaults to <root>/master")
		exportsRoot = flag.String("exports", "", "exported assets directory; defaults to <root>/exports")
		addr        = flag.String("addr", ":8080", "HTTP listen address")
		corsOrigin  = flag.String("cors-origin", "http://localhost:3000", "allowed browser origin for direct API requests; empty disables CORS")
	)
	flag.Parse()

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{TimeFormat: time.Kitchen}))
	if *masterRoot == "" {
		*masterRoot = filepath.Join(*root, "master")
	}
	if *exportsRoot == "" {
		*exportsRoot = filepath.Join(*root, "exports")
	}

	allowedOrigins := []string{}
	if *corsOrigin != "" {
		allowedOrigins = append(allowedOrigins, *corsOrigin)
	}

	server, err := webapi.New(webapi.Config{MasterRoot: *masterRoot, ExportsRoot: *exportsRoot, AllowedOrigins: allowedOrigins})
	if err != nil {
		logger.Error("create API server", "err", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           server,
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("vivere API server listening", "addr", *addr, "master", *masterRoot, "exports", *exportsRoot)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("serve", "err", err)
		os.Exit(1)
	}
}
