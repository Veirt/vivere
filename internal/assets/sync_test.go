package assets

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/veirt/vivere/internal/boi"
)

type stubAssetGetter struct {
	responses map[string][]byte
	onGet     func(ctx context.Context, rawURL string) error
}

func (g *stubAssetGetter) Get(ctx context.Context, rawURL string) ([]byte, error) {
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
		return nil, errors.New("missing: " + rawURL)
	}
	return data, nil
}

func TestSyncDownloadsBundlesInDeterministicOrderAndSkipsMatchingSizes(t *testing.T) {
	root := t.TempDir()
	bundlesDir := filepath.Join(root, "boi-bundles")
	if err := os.MkdirAll(bundlesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	existing := []byte("12345")
	if err := os.WriteFile(filepath.Join(bundlesDir, "bundle_existing"), existing, 0o644); err != nil {
		t.Fatal(err)
	}

	catalogJSON := buildTestCatalogJSON(t, []testBundleSpec{
		{provider: providerID, primaryKey: "bundle_z", size: 3},
		{provider: providerID, primaryKey: "bundle_existing", size: int64(len(existing))},
		{provider: providerID, primaryKey: "bundle_a", size: 4},
		{provider: "OtherProvider", primaryKey: "ignored_bundle", size: 9},
	})

	getter := &stubAssetGetter{
		responses: map[string][]byte{
			"https://cdn.example/Android/v1.json":         catalogJSON,
			"https://cdn.example/Android/bundle_a":        []byte("aaaa"),
			"https://cdn.example/Android/bundle_z":        []byte("zzz"),
			"https://cdn.example/Android/bundle_existing": existing,
		},
	}

	syncer := New(slog.New(slog.NewTextHandler(io.Discard, nil)), getter, root, "Android", 3)
	summary, err := syncer.Sync(context.Background(), boi.DataURIInfo{
		AssetVersion:         "v1",
		AssetCatalogFixedFmt: "https://cdn.example/{0}",
	})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if summary.Candidates != 3 || summary.Downloaded != 2 || summary.Skipped != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	wantPaths := []string{
		filepath.Join(bundlesDir, "bundle_a"),
		filepath.Join(bundlesDir, "bundle_z"),
	}
	if len(summary.DownloadedPaths) != len(wantPaths) || summary.DownloadedPaths[0] != wantPaths[0] || summary.DownloadedPaths[1] != wantPaths[1] {
		t.Fatalf("DownloadedPaths = %v, want %v", summary.DownloadedPaths, wantPaths)
	}
}

func TestSyncRejectsTraversalAndPropagatesCancellation(t *testing.T) {
	root := t.TempDir()
	badCatalog := buildTestCatalogJSON(t, []testBundleSpec{
		{provider: providerID, primaryKey: "../escape", size: 4},
	})

	syncer := New(slog.New(slog.NewTextHandler(io.Discard, nil)), &stubAssetGetter{
		responses: map[string][]byte{
			"https://cdn.example/Android/v1.json": badCatalog,
		},
	}, root, "Android", 2)

	if _, err := syncer.Sync(context.Background(), boi.DataURIInfo{
		AssetVersion:         "v1",
		AssetCatalogFixedFmt: "https://cdn.example/{0}",
	}); err == nil {
		t.Fatal("expected error for traversal bundle id, got nil")
	}

	cancelCatalog := buildTestCatalogJSON(t, []testBundleSpec{
		{provider: providerID, primaryKey: "bundle_1", size: 4},
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancelSyncer := New(slog.New(slog.NewTextHandler(io.Discard, nil)), &stubAssetGetter{
		responses: map[string][]byte{
			"https://cdn.example/Android/v1.json": cancelCatalog,
		},
		onGet: func(_ context.Context, rawURL string) error {
			if rawURL != "https://cdn.example/Android/v1.json" {
				cancel()
				return context.Canceled
			}
			return nil
		},
	}, root, "Android", 2)

	if _, err := cancelSyncer.Sync(ctx, boi.DataURIInfo{
		AssetVersion:         "v1",
		AssetCatalogFixedFmt: "https://cdn.example/{0}",
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Sync() error = %v, want context.Canceled", err)
	}
}

type testBundleSpec struct {
	provider   string
	primaryKey string
	size       int64
}

func buildTestCatalogJSON(t *testing.T, specs []testBundleSpec) []byte {
	t.Helper()
	providerIndex := make(map[string]int32)
	var providers []string
	var internalIDs []string

	keyBuf := new(bytes.Buffer)
	mustWrite(t, keyBuf, int32(len(specs)))

	bucketBuf := new(bytes.Buffer)
	mustWrite(t, bucketBuf, int32(len(specs)))

	extraBuf := new(bytes.Buffer)
	entryBuf := new(bytes.Buffer)
	mustWrite(t, entryBuf, int32(len(specs)))

	for i, spec := range specs {
		pIdx, ok := providerIndex[spec.provider]
		if !ok {
			pIdx = int32(len(providers))
			providers = append(providers, spec.provider)
			providerIndex[spec.provider] = pIdx
		}
		internalIDs = append(internalIDs, "internal/"+spec.primaryKey)

		keyOffset := int32(keyBuf.Len())
		keyBuf.WriteByte(0)
		mustWrite(t, keyBuf, int32(len(spec.primaryKey)))
		keyBuf.WriteString(spec.primaryKey)

		mustWrite(t, bucketBuf, keyOffset)
		mustWrite(t, bucketBuf, int32(1))
		mustWrite(t, bucketBuf, int32(i))

		dataOffset := int32(extraBuf.Len())
		extraBuf.WriteByte(7)
		extraBuf.WriteByte(uint8(len("asm")))
		extraBuf.WriteString("asm")
		extraBuf.WriteByte(uint8(len("cls")))
		extraBuf.WriteString("cls")
		utf16JSON := utf16leString(t, `{"m_BundleSize":`+jsonNumber(spec.size)+`}`)
		mustWrite(t, extraBuf, int32(len(utf16JSON)))
		extraBuf.Write(utf16JSON)

		for _, field := range []int32{int32(i), pIdx, -1, 0, dataOffset, int32(i), 0} {
			mustWrite(t, entryBuf, field)
		}
	}

	raw, err := json.Marshal(map[string]any{
		"m_ProviderIds":      providers,
		"m_InternalIds":      internalIDs,
		"m_KeyDataString":    base64.StdEncoding.EncodeToString(keyBuf.Bytes()),
		"m_BucketDataString": base64.StdEncoding.EncodeToString(bucketBuf.Bytes()),
		"m_EntryDataString":  base64.StdEncoding.EncodeToString(entryBuf.Bytes()),
		"m_ExtraDataString":  base64.StdEncoding.EncodeToString(extraBuf.Bytes()),
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func utf16leString(t *testing.T, s string) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	for _, r := range s {
		mustWrite(t, buf, uint16(r))
	}
	return buf.Bytes()
}

func jsonNumber(v int64) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func mustWrite(t *testing.T, buf *bytes.Buffer, v any) {
	t.Helper()
	if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
		t.Fatal(err)
	}
}
