package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	"github.com/veirt/vivere/internal/webapi"
	"golang.org/x/term"
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

	var handler slog.Handler
	if term.IsTerminal(int(os.Stderr.Fd())) {
		handler = tint.NewHandler(os.Stderr, &tint.Options{TimeFormat: time.Kitchen})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, nil)
	}
	logger := slog.New(handler)

	if *masterRoot == "" {
		*masterRoot = filepath.Join(*root, "master")
	}
	if *exportsRoot == "" {
		*exportsRoot = filepath.Join(*root, "exports")
	}

	var allowedOrigins []string
	if *corsOrigin != "" {
		allowedOrigins = append(allowedOrigins, *corsOrigin)
	}

	server, err := webapi.New(webapi.Config{
		MasterRoot:     *masterRoot,
		ExportsRoot:    *exportsRoot,
		AllowedOrigins: allowedOrigins,
	})
	if err != nil {
		logger.Error("create API server", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runServer(ctx, logger, *addr, *masterRoot, *exportsRoot, server); err != nil {
		logger.Error("serve", "err", err)
		os.Exit(1)
	}
}

func runServer(ctx context.Context, logger *slog.Logger, addr, masterRoot, exportsRoot string, handler http.Handler) error {
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Go(func() {
		errCh <- httpServer.ListenAndServe()
	})
	defer wg.Wait()

	logger.Info("vivere API server listening", "addr", addr, "master", masterRoot, "exports", exportsRoot)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		logger.Info("vivere API server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
