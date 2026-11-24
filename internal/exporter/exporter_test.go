package exporter

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/veirt/vivere/internal/config"
)

func TestCleanupDuplicateExports(t *testing.T) {
	root := t.TempDir()
	keep := filepath.Join(root, "foo.png")
	dup := filepath.Join(root, "foo_#1.png")
	nestedDup := filepath.Join(root, "nested", "bar_#2.txt")
	for _, path := range []string{keep, dup, nestedDup} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	deleted, remaining, err := cleanupAndCountExports(root)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 2 || remaining != 1 {
		t.Fatalf("cleanupAndCountExports = (%d, %d), want (2, 1)", deleted, remaining)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("keep file missing: %v", err)
	}
	if _, err := os.Stat(dup); !os.IsNotExist(err) {
		t.Fatalf("dup file still exists: %v", err)
	}
}

func TestResolveOutputDir(t *testing.T) {
	ex := New(slog.Default(), config.Config{OutputRoot: "/tmp/root", Export: config.ExportConfig{OutputDir: "exports", StagingDir: "tmp/export-staging", StateFile: "state/export-manifest.json"}})
	if got := ex.resolveOutputDir(); got != "/tmp/root/exports" {
		t.Fatalf("resolveOutputDir()=%q", got)
	}
	if got := ex.resolveStagingDir(); got != "/tmp/root/tmp/export-staging" {
		t.Fatalf("resolveStagingDir()=%q", got)
	}
	if got := ex.resolveStateFile(); got != "/tmp/root/state/export-manifest.json" {
		t.Fatalf("resolveStateFile()=%q", got)
	}
}

func TestSplitBatches(t *testing.T) {
	batches := splitBatches([]string{"a", "b", "c", "d", "e"}, 2)
	if len(batches) != 3 {
		t.Fatalf("len(batches)=%d want 3", len(batches))
	}
	if len(batches[0]) != 2 || len(batches[1]) != 2 || len(batches[2]) != 1 {
		t.Fatalf("unexpected batch sizes: %#v", batches)
	}
}

func TestNeedsExport(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file.bundle")
	if err := os.WriteFile(path, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := exportManifest{}
	pending, err := needsExport(path, manifest)
	if err != nil || !pending {
		t.Fatalf("needsExport empty manifest = (%v, %v), want (true, nil)", pending, err)
	}
	if err := recordBatchInManifest(manifest, "test", []string{path}); err != nil {
		t.Fatal(err)
	}
	pending, err = needsExport(path, manifest)
	if err != nil || pending {
		t.Fatalf("needsExport recorded manifest = (%v, %v), want (false, nil)", pending, err)
	}
}
