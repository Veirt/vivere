package apk

import (
	"archive/zip"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/veirt/vivere/internal/config"
	"github.com/veirt/vivere/internal/state"
)

func TestExtractPlatformDirCopiesNestedFiles(t *testing.T) {
	root := t.TempDir()
	apkPath := filepath.Join(root, "test.apk")
	writeTestAPK(t, apkPath, map[string]string{
		"assets/aa/Android/root.bundle":        "root",
		"assets/aa/Android/nested/file.bundle": "nested",
		"assets/aa/iOS/ignored.bundle":         "ignored",
	})

	s := New(slog.Default(), config.Config{OutputRoot: root, Platform: "Android"}, state.New(root))
	staging, err := s.extractPlatformDir(apkPath)
	if err != nil {
		t.Fatalf("extractPlatformDir() error = %v", err)
	}
	copied, err := s.copyChangedFiles(staging, filepath.Join(root, "out"))
	if err != nil {
		t.Fatalf("copyChangedFiles() error = %v", err)
	}
	if copied != 2 {
		t.Fatalf("copyChangedFiles() copied = %d, want 2", copied)
	}
	for _, name := range []string{"root.bundle", filepath.Join("nested", "file.bundle")} {
		if _, err := os.Stat(filepath.Join(root, "out", name)); err != nil {
			t.Fatalf("expected copied file %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "out", "ignored.bundle")); !os.IsNotExist(err) {
		t.Fatalf("ignored platform file exists or stat failed unexpectedly: %v", err)
	}
}

func writeTestAPK(t *testing.T, apkPath string, files map[string]string) {
	t.Helper()
	f, err := os.Create(apkPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
