package master

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/veirt/vivere/internal/boi"
	"github.com/veirt/vivere/internal/fsutil"
	"github.com/vmihailenco/msgpack/v5"
)

type Getter interface {
	Get(ctx context.Context, rawURL string) ([]byte, error)
}

type Syncer struct {
	logger *slog.Logger
	getter Getter
	root   string
}

type Summary struct {
	CatalogEntries int
	Downloaded     int
	Skipped        int
}

func New(logger *slog.Logger, getter Getter, outputRoot string) *Syncer {
	return &Syncer{logger: logger, getter: getter, root: filepath.Join(outputRoot, "master")}
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
	sort.Strings(keys)
	summary := Summary{CatalogEntries: len(keys)}
	for _, name := range keys {
		if ctx.Err() != nil {
			return summary, ctx.Err()
		}
		entry := catalog.MasterBookInfoMap[name]
		localPath := filepath.Join(s.root, name)
		if hash, err := fsutil.FileMD5(localPath); err == nil && strings.EqualFold(hash, entry.Hash) {
			summary.Skipped++
			continue
		}
		masterURL, err := boi.FormatURI(info.MasterURIFmt, info.MasterVersion, name)
		if err != nil {
			return summary, err
		}
		s.logger.Info("master download", "name", name, "url", masterURL)
		bytes, err := s.getter.Get(ctx, masterURL)
		if err != nil {
			return summary, err
		}
		if err := fsutil.WriteFileAtomic(localPath, bytes, 0o644); err != nil {
			return summary, err
		}
		summary.Downloaded++
	}
	s.logger.Info("master sync finished", "entries", summary.CatalogEntries, "downloaded", summary.Downloaded, "skipped", summary.Skipped, "elapsed", time.Since(started))
	return summary, nil
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
