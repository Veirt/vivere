package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OutputRoot string        `yaml:"output_root"`
	Platform   string        `yaml:"platform"`
	Interval   time.Duration `yaml:"interval"`
	Once       bool          `yaml:"once"`
	LogLevel   LogLevel      `yaml:"log_level"`
	HTTP       HTTPConfig    `yaml:"http"`
	BOI        BOIConfig     `yaml:"boi"`
	APK        APKConfig     `yaml:"apk"`
	Master     SyncConfig    `yaml:"master"`
	Assets     AssetsConfig  `yaml:"assets"`
	Export     ExportConfig  `yaml:"export"`
}

type HTTPConfig struct {
	UserAgent string `yaml:"user_agent"`
}

type BOIConfig struct {
	AuthBaseURL     string        `yaml:"auth_base_url"`
	CountryCode     string        `yaml:"country_code"`
	RequestTimeout  time.Duration `yaml:"request_timeout"`
	DownloadTimeout time.Duration `yaml:"download_timeout"`
	MaxRetries      int           `yaml:"max_retries"`
}

type APKConfig struct {
	VarsURL         string        `yaml:"vars_url"`
	RequestTimeout  time.Duration `yaml:"request_timeout"`
	DownloadTimeout time.Duration `yaml:"download_timeout"`
}

type SyncConfig struct {
	Enabled bool `yaml:"enabled"`
}

type AssetsConfig struct {
	Enabled     bool `yaml:"enabled"`
	Concurrency int  `yaml:"concurrency"`
}

type ExportConfig struct {
	Enabled            bool     `yaml:"enabled"`
	AssetStudioCLIPath string   `yaml:"assetstudio_cli_path"`
	OutputDir          string   `yaml:"output_dir"`
	AssetTypes         []string `yaml:"asset_types"`
	BatchSize          int      `yaml:"batch_size"`
	StagingDir         string   `yaml:"staging_dir"`
	StateFile          string   `yaml:"state_file"`
}

type LogLevel string

func (l LogLevel) Level() slog.Level {
	switch strings.ToLower(string(l)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Load(path string) (Config, error) {
	cfg := Config{
		OutputRoot: "./data",
		Platform:   "Android",
		Interval:   time.Hour,
		Once:       false,
		LogLevel:   "info",
		HTTP: HTTPConfig{
			UserAgent: "vivere-downloader/0.1",
		},
		BOI: BOIConfig{
			AuthBaseURL:     "https://prd1-auth.mememori-boi.com/api/",
			CountryCode:     "CN",
			RequestTimeout:  30 * time.Second,
			DownloadTimeout: 10 * time.Minute,
			MaxRetries:      10,
		},
		APK: APKConfig{
			VarsURL:         "https://mememori-game.com/apps/vars.js",
			RequestTimeout:  30 * time.Second,
			DownloadTimeout: 20 * time.Minute,
		},
		Master: SyncConfig{Enabled: true},
		Assets: AssetsConfig{Enabled: true, Concurrency: 8},
		Export: ExportConfig{
			Enabled:            false,
			AssetStudioCLIPath: "",
			OutputDir:          "exports",
			AssetTypes:         []string{"tex2d", "audio", "video", "textAsset", "sprite"},
			BatchSize:          100,
			StagingDir:         "tmp/export-staging",
			StateFile:          "state/export-manifest.json",
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	cfg.OutputRoot = filepath.Clean(cfg.OutputRoot)
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error
	if strings.TrimSpace(c.OutputRoot) == "" {
		errs = append(errs, errors.New("output_root is required"))
	}
	if strings.TrimSpace(c.Platform) == "" {
		errs = append(errs, errors.New("platform is required"))
	}
	if strings.TrimSpace(c.BOI.AuthBaseURL) == "" {
		errs = append(errs, errors.New("boi.auth_base_url is required"))
	}
	if strings.TrimSpace(c.BOI.CountryCode) == "" {
		errs = append(errs, errors.New("boi.country_code is required"))
	}
	if c.BOI.RequestTimeout <= 0 {
		errs = append(errs, errors.New("boi.request_timeout must be > 0"))
	}
	if c.APK.RequestTimeout <= 0 {
		errs = append(errs, errors.New("apk.request_timeout must be > 0"))
	}
	if c.BOI.DownloadTimeout <= 0 {
		errs = append(errs, errors.New("boi.download_timeout must be > 0"))
	}
	if c.APK.DownloadTimeout <= 0 {
		errs = append(errs, errors.New("apk.download_timeout must be > 0"))
	}
	if c.BOI.MaxRetries < 1 {
		errs = append(errs, errors.New("boi.max_retries must be >= 1"))
	}
	if c.Assets.Concurrency < 1 {
		errs = append(errs, errors.New("assets.concurrency must be >= 1"))
	}
	if c.Export.Enabled {
		if strings.TrimSpace(c.Export.AssetStudioCLIPath) == "" {
			errs = append(errs, errors.New("export.assetstudio_cli_path is required when export.enabled=true"))
		}
		if strings.TrimSpace(c.Export.OutputDir) == "" {
			errs = append(errs, errors.New("export.output_dir is required when export.enabled=true"))
		}
		if len(c.Export.AssetTypes) == 0 {
			errs = append(errs, errors.New("export.asset_types must not be empty when export.enabled=true"))
		}
		if c.Export.BatchSize < 1 {
			errs = append(errs, errors.New("export.batch_size must be >= 1 when export.enabled=true"))
		}
		if strings.TrimSpace(c.Export.StagingDir) == "" {
			errs = append(errs, errors.New("export.staging_dir is required when export.enabled=true"))
		}
		if strings.TrimSpace(c.Export.StateFile) == "" {
			errs = append(errs, errors.New("export.state_file is required when export.enabled=true"))
		}
	}
	if !c.Once && c.Interval <= 0 {
		errs = append(errs, errors.New("interval must be > 0 when once=false"))
	}
	return errors.Join(errs...)
}
