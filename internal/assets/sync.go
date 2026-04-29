package assets

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/veirt/vivere/internal/addressables"
	"github.com/veirt/vivere/internal/boi"
	"github.com/veirt/vivere/internal/fsutil"
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
	CatalogURL string
	Candidates int
	Downloaded int64
	Skipped    int64
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan addressables.ResourceLocation)
	errCh := make(chan error, 1)
	var downloaded int64
	var skipped int64
	var wg sync.WaitGroup
	for i := 0; i < s.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for loc := range jobs {
				if err := s.syncOne(ctx, info, loc, &downloaded, &skipped); err != nil {
					select {
					case errCh <- err:
					default:
					}
					cancel()
					return
				}
			}
		}()
	}
	candidateCount := 0
sendLoop:
	for _, loc := range catalog.Resources {
		if loc.ProviderID != providerID {
			continue
		}
		candidateCount++
		select {
		case <-ctx.Done():
			break sendLoop
		case jobs <- loc:
		}
	}
	close(jobs)
	wg.Wait()
	select {
	case err := <-errCh:
		return Summary{}, err
	default:
	}
	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}
	s.logger.Info("asset sync finished", "catalog_url", catalogURL, "candidates", candidateCount, "downloaded", downloaded, "skipped", skipped, "elapsed", time.Since(started))
	return Summary{CatalogURL: catalogURL, Candidates: candidateCount, Downloaded: downloaded, Skipped: skipped}, nil
}

func (s *Syncer) syncOne(ctx context.Context, info boi.DataURIInfo, loc addressables.ResourceLocation, downloaded, skipped *int64) error {
	bundleID := loc.PrimaryKey
	localPath, err := safeBundlePath(s.root, bundleID)
	if err != nil {
		return err
	}
	if expected, ok := addressables.BundleSize(loc); ok {
		if stat, err := os.Stat(localPath); err == nil && stat.Size() == expected {
			atomic.AddInt64(skipped, 1)
			return nil
		}
	} else if _, err := os.Stat(localPath); err == nil {
		atomic.AddInt64(skipped, 1)
		return nil
	}
	bundleURL, err := boi.FormatURI(info.AssetCatalogFixedFmt, fmt.Sprintf("%s/%s", s.platform, bundleID))
	if err != nil {
		return err
	}
	s.logger.Info("bundle download", "bundle_id", bundleID, "url", bundleURL)
	data, err := s.getter.Get(ctx, bundleURL)
	if err != nil {
		return err
	}
	if err := fsutil.WriteFileAtomic(localPath, data, 0o644); err != nil {
		return err
	}
	atomic.AddInt64(downloaded, 1)
	return nil
}

func safeBundlePath(root, bundleID string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(bundleID))
	if clean == "." || clean == ".." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid bundle id %q", bundleID)
	}
	return filepath.Join(root, clean), nil
}
