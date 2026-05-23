package apk

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/veirt/vivere/internal/config"
	"github.com/veirt/vivere/internal/fsutil"
	"github.com/veirt/vivere/internal/state"
)

type Syncer struct {
	logger *slog.Logger
	client *http.Client
	cfg    config.Config
	state  state.Store
}

type Summary struct {
	Version       string
	APKDownloaded bool
	FilesCopied   int
	ChangedFiles  []string
}

func New(logger *slog.Logger, cfg config.Config, store state.Store) *Syncer {
	return &Syncer{
		logger: logger,
		client: &http.Client{Timeout: cfg.APK.DownloadTimeout},
		cfg:    cfg,
		state:  store,
	}
}

func (s *Syncer) Sync(ctx context.Context, version string) (Summary, error) {
	started := time.Now()
	lastVersion, err := s.state.ReadString("apk-version.txt")
	if err != nil {
		return Summary{}, err
	}
	if version == lastVersion {
		s.logger.Info("apk version unchanged", "version", version)
		return Summary{Version: version}, nil
	}
	apkDir := filepath.Join(s.cfg.OutputRoot, "apk")
	extractDir := filepath.Join(s.cfg.OutputRoot, "apk-extract", "assets", "aa", s.cfg.Platform)
	apkName := fmt.Sprintf("mementomori_%s.apk", version)
	apkURL := fmt.Sprintf("https://mememori-game.com/apps/%s", apkName)
	apkPath := filepath.Join(apkDir, apkName)
	if err := fsutil.EnsureDir(apkDir); err != nil {
		return Summary{}, err
	}
	downloaded := false
	if _, err := os.Stat(apkPath); err != nil {
		if !os.IsNotExist(err) {
			return Summary{}, err
		}
		s.logger.Info("apk download start", "version", version, "url", apkURL, "path", apkPath)
		if err := s.downloadFile(ctx, apkURL, apkPath); err != nil {
			return Summary{}, err
		}
		if info, err := os.Stat(apkPath); err == nil {
			s.logger.Info("apk download complete", "version", version, "path", apkPath, "bytes", info.Size())
		}
		downloaded = true
	} else {
		s.logger.Info("apk already present", "version", version, "path", apkPath)
	}
	s.logger.Info("apk extract start", "version", version, "path", apkPath)
	platformSource, err := s.extractPlatformDir(apkPath)
	if err != nil {
		return Summary{}, err
	}
	copied, changedFiles, err := s.copyChangedFiles(platformSource, extractDir)
	if err != nil {
		return Summary{}, err
	}
	if err := s.state.WriteString("apk-version.txt", version); err != nil {
		return Summary{}, err
	}
	s.logger.Info("apk sync finished", "version", version, "files_copied", copied, "elapsed", time.Since(started))
	return Summary{Version: version, APKDownloaded: downloaded, FilesCopied: copied, ChangedFiles: changedFiles}, nil
}

func (s *Syncer) downloadFile(ctx context.Context, rawURL, dst string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download apk: unexpected status %s", resp.Status)
	}
	if err := fsutil.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".apk-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, dst)
}

func (s *Syncer) extractPlatformDir(apkPath string) (string, error) {
	zr, err := zip.OpenReader(apkPath)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	stagingRoot := filepath.Join(s.cfg.OutputRoot, "tmp", "apk-extract")
	if err := os.RemoveAll(stagingRoot); err != nil {
		return "", err
	}
	platformRoot := path.Join("assets", "aa", s.cfg.Platform)
	platformPrefix := platformRoot + "/"
	extracted := 0
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := path.Clean(filepath.ToSlash(f.Name))
		if !strings.HasPrefix(name, platformPrefix) {
			continue
		}
		rel := strings.TrimPrefix(name, platformPrefix)
		if rel == "." || rel == "" || strings.HasPrefix(rel, "../") || path.IsAbs(rel) {
			return "", fmt.Errorf("unsafe apk path %q", f.Name)
		}
		target := filepath.Join(stagingRoot, filepath.FromSlash(rel))
		if err := extractZipFile(f, target); err != nil {
			return "", err
		}
		extracted++
	}
	if extracted == 0 {
		return "", fmt.Errorf("apk contains no files under %s", platformRoot)
	}
	return stagingRoot, nil
}

func extractZipFile(f *zip.File, target string) error {
	if err := fsutil.EnsureDir(filepath.Dir(target)); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, rc)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func (s *Syncer) copyChangedFiles(srcDir, dstDir string) (int, []string, error) {
	if err := fsutil.EnsureDir(dstDir); err != nil {
		return 0, nil, err
	}
	copied := 0
	changedFiles := make([]string, 0)
	err := filepath.WalkDir(srcDir, func(src string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, src)
		if err != nil {
			return err
		}
		dst := filepath.Join(dstDir, rel)
		srcInfo, err := entry.Info()
		if err != nil {
			return err
		}
		if dstInfo, err := os.Stat(dst); err == nil && dstInfo.Size() == srcInfo.Size() {
			return nil
		}
		if err := copyFileAtomic(src, dst, 0o644); err != nil {
			return err
		}
		copied++
		changedFiles = append(changedFiles, dst)
		return nil
	})
	return copied, changedFiles, err
}

func copyFileAtomic(src, dst string, perm os.FileMode) error {
	if err := fsutil.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, dst)
}
