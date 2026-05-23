package exporter

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/veirt/vivere/internal/config"
	"github.com/veirt/vivere/internal/fsutil"
)

var duplicateSuffixRegexp = regexp.MustCompile(`_#\d+$`)

type Exporter struct {
	logger *slog.Logger
	cfg    config.Config
}

type InputGroup struct {
	Name      string
	BaseDir   string
	Files     []string
	BatchSize int
}

type Summary struct {
	OutputDir         string
	Groups            int
	Batches           int
	FilesBefore       int
	FilesAfterCleanup int
	DeletedDuplicates int
	Elapsed           time.Duration
}

type manifestEntry struct {
	Group       string `json:"group"`
	Size        int64  `json:"size"`
	ModTimeUnix int64  `json:"mod_time_unix"`
}

type exportManifest map[string]manifestEntry

func New(logger *slog.Logger, cfg config.Config) *Exporter {
	return &Exporter{logger: logger, cfg: cfg}
}

func (e *Exporter) Run(ctx context.Context, groups []InputGroup) (Summary, error) {
	started := time.Now()
	manifestPath := e.resolveStateFile()
	manifest, err := loadManifest(manifestPath)
	if err != nil {
		return Summary{}, err
	}
	prepared, err := e.prepareGroups(groups, manifest)
	if err != nil {
		return Summary{}, err
	}
	outputDir := e.resolveOutputDir()
	if err := fsutil.EnsureDir(outputDir); err != nil {
		return Summary{}, err
	}
	stagingRoot := e.resolveStagingDir()
	if err := os.RemoveAll(stagingRoot); err != nil && !os.IsNotExist(err) {
		return Summary{}, fmt.Errorf("clear export staging dir: %w", err)
	}
	if err := fsutil.EnsureDir(stagingRoot); err != nil {
		return Summary{}, err
	}
	defer os.RemoveAll(stagingRoot)

	summary := Summary{OutputDir: outputDir, Groups: len(prepared)}
	if len(prepared) == 0 {
		e.logger.Info("asset export skipped", "reason", "no pending files")
		return summary, nil
	}
	for _, group := range prepared {
		summary.FilesBefore += len(group.Files)
		for batchIndex, batch := range splitBatches(group.Files, group.BatchSize) {
			summary.Batches++
			batchDir := filepath.Join(stagingRoot, fmt.Sprintf("%s-%04d", sanitizeName(group.Name), batchIndex+1))
			if err := stageBatch(group.BaseDir, batchDir, batch); err != nil {
				return summary, err
			}
			e.logger.Info("asset export batch start", "group", group.Name, "batch", batchIndex+1, "batch_size", len(batch), "output_dir", outputDir, "staging_dir", batchDir, "asset_types", strings.Join(e.cfg.Export.AssetTypes, ","))
			if err := e.runAssetStudio(ctx, batchDir, outputDir); err != nil {
				return summary, err
			}
			if err := recordBatchInManifest(manifest, group.Name, batch); err != nil {
				return summary, err
			}
			if err := saveManifest(manifestPath, manifest); err != nil {
				return summary, err
			}
			if err := os.RemoveAll(batchDir); err != nil {
				return summary, fmt.Errorf("cleanup export batch dir %s: %w", batchDir, err)
			}
		}
	}
	deleted, err := cleanupDuplicateExports(outputDir)
	if err != nil {
		return summary, err
	}
	summary.DeletedDuplicates = deleted
	after, err := countFiles(outputDir)
	if err != nil {
		return summary, err
	}
	summary.FilesAfterCleanup = after
	summary.Elapsed = time.Since(started)
	e.logger.Info("asset export finished", "output_dir", outputDir, "groups", summary.Groups, "batches", summary.Batches, "files_before", summary.FilesBefore, "files_after_cleanup", summary.FilesAfterCleanup, "deleted_duplicates", summary.DeletedDuplicates, "elapsed", summary.Elapsed)
	return summary, nil
}

func (e *Exporter) prepareGroups(groups []InputGroup, manifest exportManifest) ([]InputGroup, error) {
	prepared := make([]InputGroup, 0, len(groups))
	for _, group := range groups {
		baseDir := filepath.Clean(group.BaseDir)
		allFiles, err := listFiles(baseDir)
		if err != nil {
			if os.IsNotExist(err) {
				e.logger.Warn("export base dir missing, skipping", "group", group.Name, "base_dir", baseDir)
				continue
			}
			return nil, err
		}
		pendingSet := make(map[string]struct{})
		for _, file := range allFiles {
			pending, err := needsExport(file, manifest)
			if err != nil {
				return nil, err
			}
			if pending {
				pendingSet[file] = struct{}{}
			}
		}
		for _, file := range group.Files {
			if strings.TrimSpace(file) == "" {
				continue
			}
			clean := filepath.Clean(file)
			if !filepath.IsAbs(clean) {
				clean = filepath.Join(baseDir, clean)
			}
			if _, err := os.Stat(clean); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, fmt.Errorf("stat export file %s: %w", clean, err)
			}
			pendingSet[clean] = struct{}{}
		}
		if len(pendingSet) == 0 {
			continue
		}
		files := make([]string, 0, len(pendingSet))
		for file := range pendingSet {
			files = append(files, file)
		}
		sort.Strings(files)
		if group.BatchSize < 1 {
			group.BatchSize = e.cfg.Export.BatchSize
		}
		prepared = append(prepared, InputGroup{Name: group.Name, BaseDir: baseDir, Files: files, BatchSize: group.BatchSize})
	}
	return prepared, nil
}

func (e *Exporter) resolveOutputDir() string {
	if filepath.IsAbs(e.cfg.Export.OutputDir) {
		return e.cfg.Export.OutputDir
	}
	return filepath.Join(e.cfg.OutputRoot, filepath.FromSlash(e.cfg.Export.OutputDir))
}

func (e *Exporter) resolveStagingDir() string {
	if filepath.IsAbs(e.cfg.Export.StagingDir) {
		return e.cfg.Export.StagingDir
	}
	return filepath.Join(e.cfg.OutputRoot, filepath.FromSlash(e.cfg.Export.StagingDir))
}

func (e *Exporter) resolveStateFile() string {
	if filepath.IsAbs(e.cfg.Export.StateFile) {
		return e.cfg.Export.StateFile
	}
	return filepath.Join(e.cfg.OutputRoot, filepath.FromSlash(e.cfg.Export.StateFile))
}

func (e *Exporter) runAssetStudio(ctx context.Context, inputDir, outputDir string) error {
	args := []string{inputDir, "-t", strings.Join(e.cfg.Export.AssetTypes, ","), "-o", outputDir}
	cmd := exec.CommandContext(ctx, e.cfg.Export.AssetStudioCLIPath, args...)
	e.logger.Info("assetstudio exec", "command", e.cfg.Export.AssetStudioCLIPath, "args", strings.Join(args, " "))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start AssetStudioCLI: %w", err)
	}
	go scanPipe(e.logger, stdout, "assetstudio stdout")
	go scanPipe(e.logger, stderr, "assetstudio stderr")
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("AssetStudioCLI failed for %s: %w", inputDir, err)
	}
	return nil
}

func scanPipe(logger *slog.Logger, pipe io.Reader, msg string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		logger.Info(msg, "line", scanner.Text())
	}
}

func splitBatches(files []string, batchSize int) [][]string {
	if batchSize < 1 {
		batchSize = len(files)
	}
	batches := make([][]string, 0, (len(files)+batchSize-1)/batchSize)
	for start := 0; start < len(files); start += batchSize {
		end := start + batchSize
		if end > len(files) {
			end = len(files)
		}
		batches = append(batches, files[start:end])
	}
	return batches
}

func stageBatch(baseDir, batchDir string, files []string) error {
	if err := fsutil.EnsureDir(batchDir); err != nil {
		return err
	}
	baseDir = filepath.Clean(baseDir)
	for _, src := range files {
		rel, err := filepath.Rel(baseDir, src)
		if err != nil {
			return err
		}
		if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("file %s is outside base dir %s", src, baseDir)
		}
		dst := filepath.Join(batchDir, rel)
		if err := linkOrCopyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func linkOrCopyFile(src, dst string) error {
	if err := fsutil.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	if err := os.Link(src, dst); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrExist) {
		if err := os.Symlink(src, dst); err == nil {
			return nil
		}
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func loadManifest(path string) (exportManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return exportManifest{}, nil
		}
		return nil, fmt.Errorf("read export manifest: %w", err)
	}
	var manifest exportManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse export manifest: %w", err)
	}
	if manifest == nil {
		manifest = exportManifest{}
	}
	return manifest, nil
}

func saveManifest(path string, manifest exportManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal export manifest: %w", err)
	}
	return fsutil.WriteFileAtomic(path, data, 0o644)
}

func recordBatchInManifest(manifest exportManifest, group string, files []string) error {
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("stat exported source file %s: %w", file, err)
		}
		manifest[file] = manifestEntry{Group: group, Size: info.Size(), ModTimeUnix: info.ModTime().Unix()}
	}
	return nil
}

func needsExport(path string, manifest exportManifest) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	entry, ok := manifest[path]
	if !ok {
		return true, nil
	}
	return entry.Size != info.Size() || entry.ModTimeUnix != info.ModTime().Unix(), nil
}

func listFiles(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func cleanupDuplicateExports(root string) (int, error) {
	deleted := 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if !duplicateSuffixRegexp.MatchString(name) {
			return nil
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		deleted++
		return nil
	})
	return deleted, err
}

func countFiles(root string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			count++
		}
		return nil
	})
	if os.IsNotExist(err) {
		return 0, nil
	}
	return count, err
}

func sanitizeName(name string) string {
	replacer := strings.NewReplacer(string(os.PathSeparator), "-", " ", "-", "/", "-")
	return replacer.Replace(strings.ToLower(strings.TrimSpace(name)))
}
