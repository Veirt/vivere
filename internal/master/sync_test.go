package master

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/veirt/vivere/internal/boi"
	"github.com/vmihailenco/msgpack/v5"
)

type stubGetter struct {
	responses map[string][]byte
	calls     atomic.Int64
	onGet     func(ctx context.Context, rawURL string) error
}

func (g *stubGetter) Get(ctx context.Context, rawURL string) ([]byte, error) {
	g.calls.Add(1)
	if g.onGet != nil {
		if err := g.onGet(ctx, rawURL); err != nil {
			return nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, ok := g.responses[rawURL]
	if !ok {
		return nil, errors.New("not found: " + rawURL)
	}
	return data, nil
}

func md5Hex(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

func TestSyncDownloadsMissingBooksSkipsUpToDateAndFiltersLanguages(t *testing.T) {
	root := t.TempDir()
	masterDir := filepath.Join(root, "master")
	if err := os.MkdirAll(masterDir, 0o755); err != nil {
		t.Fatal(err)
	}

	existingData := []byte("already-here")
	if err := os.WriteFile(filepath.Join(masterDir, "ExistingMB"), existingData, 0o644); err != nil {
		t.Fatal(err)
	}

	charData := []byte("character-bytes")
	enData := []byte("en-bytes")
	frData := []byte("fr-ignored")

	catalogBytes, err := msgpack.Marshal(boi.MasterBookCatalog{
		MasterBookInfoMap: map[string]boi.MasterBookInfo{
			"ExistingMB":         {Name: "ExistingMB", Hash: md5Hex(existingData), Size: len(existingData)},
			"CharacterMB":        {Name: "CharacterMB", Hash: md5Hex(charData), Size: len(charData)},
			"TextResourceEnUsMB": {Name: "TextResourceEnUsMB", Hash: md5Hex(enData), Size: len(enData)},
			"TextResourceFrFrMB": {Name: "TextResourceFrFrMB", Hash: md5Hex(frData), Size: len(frData)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	getter := &stubGetter{
		responses: map[string][]byte{
			"https://cdn.example/1/master-catalog":     catalogBytes,
			"https://cdn.example/1/CharacterMB":        charData,
			"https://cdn.example/1/TextResourceEnUsMB": enData,
		},
	}

	syncer := NewWithConcurrency(slog.New(slog.NewTextHandler(io.Discard, nil)), getter, root, 4)
	summary, err := syncer.Sync(context.Background(), boi.DataURIInfo{
		MasterVersion: "1",
		MasterURIFmt:  "https://cdn.example/{0}/{1}",
	})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if summary.CatalogEntries != 3 || summary.Downloaded != 2 || summary.Skipped != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if getter.calls.Load() != 3 {
		t.Fatalf("getter calls = %d, want 3 (catalog + 2 books)", getter.calls.Load())
	}

	gotChar, err := os.ReadFile(filepath.Join(masterDir, "CharacterMB"))
	if err != nil || string(gotChar) != string(charData) {
		t.Fatalf("CharacterMB mismatch: data=%q err=%v", gotChar, err)
	}
	if _, err := os.Stat(filepath.Join(masterDir, "TextResourceFrFrMB")); !os.IsNotExist(err) {
		t.Fatalf("unexpected filtered text resource written: %v", err)
	}
}

func TestSyncPropagatesCancellation(t *testing.T) {
	root := t.TempDir()
	catalogBytes, err := msgpack.Marshal(boi.MasterBookCatalog{
		MasterBookInfoMap: map[string]boi.MasterBookInfo{
			"BookAMB": {Name: "BookAMB", Hash: "aaa"},
			"BookBMB": {Name: "BookBMB", Hash: "bbb"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	getter := &stubGetter{
		responses: map[string][]byte{
			"https://cdn.example/1/master-catalog": catalogBytes,
		},
		onGet: func(_ context.Context, rawURL string) error {
			if rawURL != "https://cdn.example/1/master-catalog" {
				cancel()
				return context.Canceled
			}
			return nil
		},
	}

	syncer := NewWithConcurrency(slog.New(slog.NewTextHandler(io.Discard, nil)), getter, root, 2)
	_, err = syncer.Sync(ctx, boi.DataURIInfo{
		MasterVersion: "1",
		MasterURIFmt:  "https://cdn.example/{0}/{1}",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Sync() error = %v, want context.Canceled", err)
	}
}
