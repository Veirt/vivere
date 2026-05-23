package app

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/veirt/vivere/internal/apk"
	"github.com/veirt/vivere/internal/assets"
	"github.com/veirt/vivere/internal/boi"
	"github.com/veirt/vivere/internal/config"
	"github.com/veirt/vivere/internal/exporter"
	"github.com/veirt/vivere/internal/fsutil"
	"github.com/veirt/vivere/internal/master"
	"github.com/veirt/vivere/internal/state"
)

func Run(ctx context.Context, logger *slog.Logger, cfg config.Config) error {
	if err := fsutil.EnsureDir(cfg.OutputRoot); err != nil {
		return fmt.Errorf("ensure output root: %w", err)
	}
	store := state.New(cfg.OutputRoot)
	client := boi.NewClient(logger.With("component", "boi"), cfg)
	apkSyncer := apk.New(logger.With("component", "apk"), cfg, store)
	masterSyncer := master.New(logger.With("component", "master"), client, cfg.OutputRoot)
	assetSyncer := assets.New(logger.With("component", "assets"), client, cfg.OutputRoot, cfg.Platform, cfg.Assets.Concurrency)
	exporterRunner := exporter.New(logger.With("component", "export"), cfg)

	runOnce := func() error {
		cycleStarted := time.Now()
		logger.Info("sync cycle start")
		apkVersion, err := client.FetchAPKVersion(ctx, cfg.APK.VarsURL)
		if err != nil {
			return fmt.Errorf("fetch apk version: %w", err)
		}
		info, err := client.ResolveDataURI(ctx, apkVersion)
		if err != nil {
			return fmt.Errorf("resolve data uri: %w", err)
		}
		if err := store.WriteString("app-version.txt", info.AppVersion); err != nil {
			return err
		}
		logger.Info("resolved boi versions", "apk_version", apkVersion, "app_version", info.AppVersion, "master_version", info.MasterVersion, "asset_version", info.AssetVersion)

		apkSummary, err := apkSyncer.Sync(ctx, apkVersion)
		if err != nil {
			return fmt.Errorf("apk sync: %w", err)
		}
		logger.Info("apk sync completed", "version", apkSummary.Version, "downloaded_apk", apkSummary.APKDownloaded, "files_copied", apkSummary.FilesCopied)

		if cfg.Master.Enabled {
			masterSummary, err := masterSyncer.Sync(ctx, info)
			if err != nil {
				return fmt.Errorf("master sync: %w", err)
			}
			logger.Info("master sync completed", "entries", masterSummary.CatalogEntries, "downloaded", masterSummary.Downloaded, "skipped", masterSummary.Skipped)
		}

		var assetSummary assets.Summary
		if cfg.Assets.Enabled {
			assetSummary, err = assetSyncer.Sync(ctx, info)
			if err != nil {
				return fmt.Errorf("asset sync: %w", err)
			}
			logger.Info("asset sync completed", "catalog_url", assetSummary.CatalogURL, "candidates", assetSummary.Candidates, "downloaded", assetSummary.Downloaded, "skipped", assetSummary.Skipped)
		}
		if cfg.Export.Enabled {
			groups := []exporter.InputGroup{
				{Name: "boi-bundles", BaseDir: filepath.Join(cfg.OutputRoot, "boi-bundles"), Files: assetSummary.DownloadedPaths, BatchSize: cfg.Export.BatchSize},
				{Name: "apk-assets", BaseDir: filepath.Join(cfg.OutputRoot, "apk-extract", "assets", "aa", cfg.Platform), Files: apkSummary.ChangedFiles, BatchSize: cfg.Export.BatchSize},
			}
			exportSummary, err := exporterRunner.Run(ctx, groups)
			if err != nil {
				return fmt.Errorf("export assets: %w", err)
			}
			logger.Info("asset export completed", "output_dir", exportSummary.OutputDir, "groups", exportSummary.Groups, "batches", exportSummary.Batches, "files_before", exportSummary.FilesBefore, "files_after_cleanup", exportSummary.FilesAfterCleanup, "deleted_duplicates", exportSummary.DeletedDuplicates, "elapsed", exportSummary.Elapsed)
		}
		logger.Info("sync cycle finished", "elapsed", time.Since(cycleStarted))
		return nil
	}

	if cfg.Once {
		return runOnce()
	}
	if err := runOnce(); err != nil {
		logger.Error("sync cycle failed", "error", err)
	}
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := runOnce(); err != nil {
				logger.Error("sync cycle failed", "error", err)
			}
		}
	}
}
