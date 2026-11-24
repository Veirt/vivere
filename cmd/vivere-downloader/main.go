package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	"github.com/veirt/vivere/internal/app"
	"github.com/veirt/vivere/internal/config"
	"golang.org/x/term"
)

func main() {
	var configPath string
	var once bool
	flag.StringVar(&configPath, "config", "configs/downloader.yaml", "path to downloader config")
	flag.BoolVar(&once, "once", false, "run one sync cycle and exit")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if once {
		cfg.Once = true
	}

	var handler slog.Handler
	if term.IsTerminal(int(os.Stdout.Fd())) {
		handler = tint.NewHandler(os.Stdout, &tint.Options{Level: cfg.LogLevel.Level(), TimeFormat: time.Kitchen})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel.Level()})
	}
	logger := slog.New(handler)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, logger, cfg); err != nil {
		logger.Error("downloader exited with error", "error", err)
		os.Exit(1)
	}
}
