// Package download streams a build's jar to disk reliably: it writes to a temp file in
// the destination directory, verifies the SHA256 from the API while streaming, and only
// then atomically renames it into place. A failed, mismatched, or cancelled download
// never touches an existing jar.
package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mbacalan/paper-mc-tui/internal/papermc"
)

var (
	// ErrChecksumMismatch means the downloaded bytes did not match the expected SHA256.
	ErrChecksumMismatch = errors.New("download: sha256 mismatch")
	// ErrSizeMismatch means the downloaded byte count did not match the expected size.
	ErrSizeMismatch = errors.New("download: size mismatch")
)

// Downloader fetches build artifacts. Its HTTP client has no timeout: the jar is large
// (~55 MB) and the caller bounds the transfer with a context deadline instead. This is
// deliberately separate from the papermc API client's short timeout.
type Downloader struct {
	httpClient *http.Client
	userAgent  string
}

// Option configures a Downloader.
type Option func(*Downloader)

// WithHTTPClient sets the underlying HTTP client (useful for tests).
func WithHTTPClient(c *http.Client) Option {
	return func(d *Downloader) {
		if c != nil {
			d.httpClient = c
		}
	}
}

// WithUserAgent sets the User-Agent header sent with the download request.
func WithUserAgent(ua string) Option {
	return func(d *Downloader) {
		if ua != "" {
			d.userAgent = ua
		}
	}
}

// NewDownloader returns a Downloader with sensible defaults, overridden by opts.
func NewDownloader(opts ...Option) *Downloader {
	d := &Downloader{
		httpClient: &http.Client{}, // no timeout; the caller's context bounds the transfer
		userAgent:  papermc.DefaultUserAgent,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// Download streams dl.URL to destPath. onProgress, if non-nil, is called with the bytes
// downloaded so far and the total expected (dl.Size, 0 if unknown); it fires on each 1%
// change and once at completion.
func (d *Downloader) Download(ctx context.Context, dl papermc.Download, destPath string, onProgress func(done, total int64)) (err error) {
	dir := filepath.Dir(destPath)
	tmp, err := os.CreateTemp(dir, ".paper-*.jar.tmp")
	if err != nil {
		return fmt.Errorf("download: create temp file: %w", err)
	}
	tmpName := tmp.Name()

	committed := false
	defer func() {
		tmp.Close()
		if !committed {
			os.Remove(tmpName)
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dl.URL, nil)
	if err != nil {
		return fmt.Errorf("download: build request: %w", err)
	}
	req.Header.Set("User-Agent", d.userAgent)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download: request %s: %w", dl.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: unexpected status %d for %s", resp.StatusCode, dl.URL)
	}

	hasher := sha256.New()
	pr := &progressReader{r: resp.Body, total: dl.Size, onProgress: onProgress}
	written, err := io.Copy(io.MultiWriter(tmp, hasher), pr)
	if err != nil {
		return fmt.Errorf("download: copy body: %w", err)
	}
	if onProgress != nil {
		onProgress(written, dl.Size) // ensure a final tick
	}

	if dl.Size > 0 && written != dl.Size {
		return fmt.Errorf("%w: got %d bytes, want %d", ErrSizeMismatch, written, dl.Size)
	}

	if dl.Checksums.SHA256 != "" {
		got := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(got, dl.Checksums.SHA256) {
			return fmt.Errorf("%w: got %s, want %s", ErrChecksumMismatch, got, dl.Checksums.SHA256)
		}
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("download: sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("download: close temp file: %w", err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("download: chmod temp file: %w", err)
	}
	if err := os.Rename(tmpName, destPath); err != nil {
		return fmt.Errorf("download: rename into place: %w", err)
	}
	committed = true
	return nil
}

// progressReader counts bytes read and reports progress on each whole-percent change.
type progressReader struct {
	r          io.Reader
	total      int64
	read       int64
	lastPct    int
	onProgress func(done, total int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.read += int64(n)
	if pr.onProgress != nil && pr.total > 0 {
		pct := int(pr.read * 100 / pr.total)
		if pct != pr.lastPct {
			pr.lastPct = pct
			pr.onProgress(pr.read, pr.total)
		}
	}
	return n, err
}
