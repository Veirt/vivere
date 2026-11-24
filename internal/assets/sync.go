package assets

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/veirt/vivere/internal/addressables"
	"github.com/veirt/vivere/internal/boi"
	"github.com/veirt/vivere/internal/fsutil"
	"golang.org/x/sync/errgroup"
)

const providerID = "Ortega.Common.OrtegaAssestBundleProvider"

type Getter interface {
	Get(ctx context.Context, rawURL string) ([]byte, error)
}

type Syncer struct {
	logger      *slog.Logger
	getter      Getter
	root        string
	platform    string
	concurrency int
}

type Summary struct {
	CatalogURL      string
	Candidates      int
	Downloaded      int64
	Skipped         int64
	DownloadedPaths []string
}

func New(logger *slog.Logger, getter Getter, outputRoot, platform string, concurrency int) *Syncer {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Syncer{logger: logger, getter: getter, root: filepath.Join(outputRoot, "boi-bundles"), platform: platform, concurrency: concurrency}
}

func (s *Syncer) Sync(ctx context.Context, info boi.DataURIInfo) (Summary, error) {
	started := time.Now()
	catalogURL, err := boi.FormatURI(info.AssetCatalogFixedFmt, fmt.Sprintf("%s/%s.json", s.platform, info.AssetVersion))
	if err != nil {
		return Summary{}, err
	}
	catalogBytes, err := s.getter.Get(ctx, catalogURL)
	if err != nil {
		return Summary{}, err
	}
	catalog, err := addressables.Parse(catalogBytes)
	if err != nil {
		return Summary{}, err
	}
	if err := fsutil.EnsureDir(s.root); err != nil {
		return Summary{}, err
	}
	bundleRoot, err := os.OpenRoot(s.root)
	if err != nil {
		return Summary{}, err
	}
	defer bundleRoot.Close()

	g, workerCtx := errgroup.WithContext(ctx)
	g.SetLimit(s.concurrency)

	var downloaded atomic.Int64
	var skipped atomic.Int64
	var pathsMu sync.Mutex
	var downloadedPaths []string

	candidateCount := 0
	for _, loc := range catalog.Resources {
		if loc.ProviderID != providerID {
			continue
		}
		if workerCtx.Err() != nil {
			break
		}
		candidateCount++
		loc := loc
		g.Go(func() error {
			localPath, wasDownloaded, err := s.syncOne(workerCtx, bundleRoot, info, loc)
			if err != nil {
				return err
			}
			if wasDownloaded {
				downloaded.Add(1)
				pathsMu.Lock()
				downloadedPaths = append(downloadedPaths, localPath)
				pathsMu.Unlock()
			} else {
				skipped.Add(1)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return Summary{}, err
	}
	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}

	slices.Sort(downloadedPaths)

	downloadedCount := downloaded.Load()
	skippedCount := skipped.Load()
	s.logger.Info("asset sync finished", "catalog_url", catalogURL, "candidates", candidateCount, "downloaded", downloadedCount, "skipped", skippedCount, "elapsed", time.Since(started))
	return Summary{
		CatalogURL:      catalogURL,
		Candidates:      candidateCount,
		Downloaded:      downloadedCount,
		Skipped:         skippedCount,
		DownloadedPaths: downloadedPaths,
	}, nil
}

func (s *Syncer) syncOne(ctx context.Context, bundleRoot *os.Root, info boi.DataURIInfo, loc addressables.ResourceLocation) (string, bool, error) {
	bundleID := loc.PrimaryKey
	relPath, err := safeBundleRelPath(bundleID)
	if err != nil {
		return "", false, err
	}
	if expected, ok := addressables.BundleSize(loc); ok {
		if stat, err := bundleRoot.Stat(relPath); err == nil && stat.Size() == expected {
			return "", false, nil
		}
	} else if _, err := bundleRoot.Stat(relPath); err == nil {
		return "", false, nil
	}
	bundleURL, err := boi.FormatURI(info.AssetCatalogFixedFmt, fmt.Sprintf("%s/%s", s.platform, bundleID))
	if err != nil {
		return "", false, err
	}
	s.logger.Info("bundle download", "bundle_id", bundleID, "url", bundleURL)
	data, err := s.getter.Get(ctx, bundleURL)
	if err != nil {
		return "", false, err
	}
	if err := fsutil.WriteFileAtomicInRoot(bundleRoot, relPath, data, 0o644); err != nil {
		return "", false, err
	}
	return filepath.Join(s.root, relPath), true, nil
}

func safeBundleRelPath(bundleID string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(bundleID))
	if clean == "." || clean == ".." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid bundle id %q", bundleID)
	}
	return clean, nil
}
