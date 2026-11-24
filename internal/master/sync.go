package master

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/veirt/vivere/internal/boi"
	"github.com/veirt/vivere/internal/fsutil"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/sync/errgroup"
)

const defaultConcurrency = 8

type Getter interface {
	Get(ctx context.Context, rawURL string) ([]byte, error)
}

type Syncer struct {
	logger      *slog.Logger
	getter      Getter
	root        string
	concurrency int
}

type Summary struct {
	CatalogEntries int
	Downloaded     int
	Skipped        int
}

type syncOutcome int

const (
	outcomeSkipped syncOutcome = iota
	outcomeDownloaded
)

func New(logger *slog.Logger, getter Getter, outputRoot string) *Syncer {
	return NewWithConcurrency(logger, getter, outputRoot, defaultConcurrency)
}

func NewWithConcurrency(logger *slog.Logger, getter Getter, outputRoot string, concurrency int) *Syncer {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Syncer{
		logger:      logger,
		getter:      getter,
		root:        filepath.Join(outputRoot, "master"),
		concurrency: concurrency,
	}
}

func (s *Syncer) Sync(ctx context.Context, info boi.DataURIInfo) (Summary, error) {
	started := time.Now()
	catalogURL, err := boi.FormatURI(info.MasterURIFmt, info.MasterVersion, "master-catalog")
	if err != nil {
		return Summary{}, err
	}
	catalogBytes, err := s.getter.Get(ctx, catalogURL)
	if err != nil {
		return Summary{}, err
	}
	var catalog boi.MasterBookCatalog
	if err := msgpack.Unmarshal(catalogBytes, &catalog); err != nil {
		return Summary{}, fmt.Errorf("decode master catalog: %w", err)
	}
	if err := fsutil.EnsureDir(s.root); err != nil {
		return Summary{}, err
	}

	keys := make([]string, 0, len(catalog.MasterBookInfoMap))
	for name := range catalog.MasterBookInfoMap {
		if allowMaster(name) {
			keys = append(keys, name)
		}
	}
	slices.Sort(keys)

	summary := Summary{CatalogEntries: len(keys)}
	if len(keys) == 0 {
		s.logger.Info("master sync finished", "entries", summary.CatalogEntries, "downloaded", 0, "skipped", 0, "elapsed", time.Since(started))
		return summary, nil
	}

	g, workerCtx := errgroup.WithContext(ctx)
	g.SetLimit(min(s.concurrency, len(keys)))
	outcomes := make([]syncOutcome, len(keys))

	for index, name := range keys {
		if workerCtx.Err() != nil {
			break
		}
		index := index
		name := name
		entry := catalog.MasterBookInfoMap[name]
		g.Go(func() error {
			outcome, err := s.syncOne(workerCtx, info, name, entry)
			if err != nil {
				return err
			}
			outcomes[index] = outcome
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return summary, err
	}
	if err := ctx.Err(); err != nil {
		return summary, err
	}

	for _, outcome := range outcomes {
		if outcome == outcomeDownloaded {
			summary.Downloaded++
		} else {
			summary.Skipped++
		}
	}
	s.logger.Info("master sync finished", "entries", summary.CatalogEntries, "downloaded", summary.Downloaded, "skipped", summary.Skipped, "elapsed", time.Since(started))
	return summary, nil
}

func (s *Syncer) syncOne(ctx context.Context, info boi.DataURIInfo, name string, entry boi.MasterBookInfo) (syncOutcome, error) {
	localPath, err := safeMasterPath(s.root, name)
	if err != nil {
		return outcomeSkipped, err
	}
	if hash, err := fsutil.FileMD5(localPath); err == nil && strings.EqualFold(hash, entry.Hash) {
		return outcomeSkipped, nil
	}
	masterURL, err := boi.FormatURI(info.MasterURIFmt, info.MasterVersion, name)
	if err != nil {
		return outcomeSkipped, err
	}
	s.logger.Info("master download", "name", name, "url", masterURL)
	data, err := s.getter.Get(ctx, masterURL)
	if err != nil {
		return outcomeSkipped, err
	}
	if err := fsutil.WriteFileAtomic(localPath, data, 0o644); err != nil {
		return outcomeSkipped, err
	}
	return outcomeDownloaded, nil
}

func safeMasterPath(root, name string) (string, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("invalid master book name %q", name)
	}
	return filepath.Join(root, name), nil
}

func allowMaster(name string) bool {
	if !strings.HasPrefix(name, "TextResource") {
		return true
	}
	switch name {
	case "TextResourceJaJpMB", "TextResourceZhTwMB", "TextResourceEnUsMB", "TextResourceKoKrMB":
		return true
	default:
		return false
	}
}
