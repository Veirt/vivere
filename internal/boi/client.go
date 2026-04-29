package boi

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/veirt/vivere/internal/config"
	"github.com/vmihailenco/msgpack/v5"
)

type Client struct {
	logger         *slog.Logger
	httpClient     *http.Client
	downloadClient *http.Client
	apkClient      *http.Client
	authBaseURL    string
	countryCode    string
	maxRetries     int
}

func NewClient(logger *slog.Logger, cfg config.Config) *Client {
	return &Client{
		logger:         logger,
		httpClient:     &http.Client{Timeout: cfg.BOI.RequestTimeout},
		downloadClient: &http.Client{Timeout: cfg.BOI.DownloadTimeout},
		apkClient:      &http.Client{Timeout: cfg.APK.RequestTimeout},
		authBaseURL:    cfg.BOI.AuthBaseURL,
		countryCode:    cfg.BOI.CountryCode,
		maxRetries:     cfg.BOI.MaxRetries,
	}
}

func (c *Client) FetchAPKVersion(ctx context.Context, varsURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, varsURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.apkClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch vars.js: unexpected status %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return ParseAPKVersion(string(body))
}

func (c *Client) ResolveDataURI(ctx context.Context, candidateVersion string) (DataURIInfo, error) {
	candidates, err := VersionCandidates(candidateVersion)
	if err != nil {
		return DataURIInfo{}, err
	}
	var lastErr error
	for _, version := range candidates {
		info, err := c.getDataURI(ctx, version)
		if err == nil {
			return info, nil
		}
		var updateErr *clientUpdateError
		if ok := AsClientUpdate(err, &updateErr); ok {
			c.logger.Debug("candidate app version rejected", "app_version", version, "error", err)
			lastErr = err
			continue
		}
		return DataURIInfo{}, err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no valid app version found from candidate %s", candidateVersion)
	}
	return DataURIInfo{}, lastErr
}

func (c *Client) getDataURI(ctx context.Context, appVersion string) (DataURIInfo, error) {
	payload, err := marshalMessagePack(getDataURIRequest{CountryCode: c.countryCode, UserID: 0})
	if err != nil {
		return DataURIInfo{}, err
	}
	endpoint, err := url.JoinPath(c.authBaseURL, "auth/getDataUri")
	if err != nil {
		return DataURIInfo{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return DataURIInfo{}, err
	}
	req.Header.Set("content-type", "application/x-msgpack")
	req.Header.Set("ortegaappversion", appVersion)
	req.Header.Set("ortegadevicetype", "2")
	req.Header.Set("ortegauuid", randomHex(16))
	req.Header.Set("user-agent", "BestHTTP/2 v2.3.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return DataURIInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return DataURIInfo{}, fmt.Errorf("getDataUri status %s", resp.Status)
	}
	statusCode := resp.Header.Get("ortegastatuscode")
	if statusCode == "" {
		return DataURIInfo{}, fmt.Errorf("getDataUri missing ortegastatuscode header")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DataURIInfo{}, err
	}
	if statusCode != "0" {
		var apiErr apiErrorResponse
		if err := msgpack.Unmarshal(body, &apiErr); err != nil {
			return DataURIInfo{}, fmt.Errorf("decode api error: %w", err)
		}
		if apiErr.ErrorCode == commonRequireClientUpdate {
			return DataURIInfo{}, &clientUpdateError{AppVersion: appVersion, ErrorCode: apiErr.ErrorCode, Message: apiErr.Message}
		}
		return DataURIInfo{}, fmt.Errorf("getDataUri error code=%d message=%q", apiErr.ErrorCode, apiErr.Message)
	}
	var decoded dataURIResponse
	if err := msgpack.Unmarshal(body, &decoded); err != nil {
		return DataURIInfo{}, fmt.Errorf("decode getDataUri: %w", err)
	}
	return DataURIInfo{
		AppVersion:           firstNonEmpty(decoded.AppAssetVersionInfo.Version, appVersion),
		MasterVersion:        resp.Header.Get("ortegamasterversion"),
		AssetVersion:         resp.Header.Get("ortegaassetversion"),
		AssetCatalogFixedFmt: decoded.AssetCatalogFixedURIFormat,
		MasterURIFmt:         decoded.MasterURIFormat,
		RawDataURIFmt:        decoded.RawDataURIFormat,
		CatalogURIFmt:        decoded.AssetCatalogURIFormat,
		NoticeBannerImageFmt: decoded.NoticeBannerImageURIFormat,
		Response:             decoded,
	}, nil
}

func (c *Client) Get(ctx context.Context, rawURL string) ([]byte, error) {
	start := time.Now()
	c.logger.Debug("download start", "url", rawURL)
	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		attemptStart := time.Now()
		data, err := c.getOnce(ctx, rawURL)
		if err == nil {
			c.logger.Info("download complete", "url", rawURL, "bytes", len(data), "attempt", attempt, "elapsed", time.Since(start), "attempt_elapsed", time.Since(attemptStart))
			return data, nil
		}
		lastErr = err
		c.logger.Warn("download attempt failed", "url", rawURL, "attempt", attempt, "elapsed", time.Since(attemptStart), "error", err)
		if attempt == c.maxRetries {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return nil, fmt.Errorf("GET %s failed after retries: %w", rawURL, lastErr)
}

func (c *Client) getOnce(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.downloadClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return strings.ToLower(hex.EncodeToString(buf))
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

type clientUpdateError struct {
	AppVersion string
	ErrorCode  int
	Message    string
}

func (e *clientUpdateError) Error() string {
	return fmt.Sprintf("client update required for app version %s: code=%d message=%q", e.AppVersion, e.ErrorCode, e.Message)
}

func AsClientUpdate(err error, target **clientUpdateError) bool {
	if err == nil {
		return false
	}
	cu, ok := err.(*clientUpdateError)
	if ok && target != nil {
		*target = cu
	}
	return ok
}
