package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mbacalan/paper-mc-tui/internal/papermc"
)

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// payloadServer serves a fixed body for any request.
func payloadServer(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func dl(srv *httptest.Server, body []byte, sha string) papermc.Download {
	return papermc.Download{
		Name:      "paper-test.jar",
		URL:       srv.URL + "/paper.jar",
		Size:      int64(len(body)),
		Checksums: papermc.Checksums{SHA256: sha},
	}
}

func noTempFiles(t *testing.T, dir string) {
	t.Helper()
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

func TestDownloadOK(t *testing.T) {
	body := []byte("this is a fake paper jar payload")
	srv := payloadServer(t, body)
	dir := t.TempDir()
	dest := filepath.Join(dir, "paper.jar")

	err := NewDownloader().Download(context.Background(), dl(srv, body, sha256Hex(body)), dest, nil)
	if err != nil {
		t.Fatalf("Download: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("content mismatch: got %q", got)
	}
	info, _ := os.Stat(dest)
	if info.Mode().Perm() != 0o644 {
		t.Errorf("perm = %v, want 0644", info.Mode().Perm())
	}
	noTempFiles(t, dir)
}

func TestDownloadChecksumMismatch(t *testing.T) {
	body := []byte("real content")
	srv := payloadServer(t, body)
	dir := t.TempDir()
	dest := filepath.Join(dir, "paper.jar")
	if err := os.WriteFile(dest, []byte("OLD JAR"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := NewDownloader().Download(context.Background(), dl(srv, body, sha256Hex([]byte("different"))), dest, nil)
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("err = %v, want ErrChecksumMismatch", err)
	}

	// Pre-existing jar must be untouched, and no temp file left behind.
	got, _ := os.ReadFile(dest)
	if string(got) != "OLD JAR" {
		t.Errorf("existing jar was modified: %q", got)
	}
	noTempFiles(t, dir)
}

func TestDownloadSizeMismatch(t *testing.T) {
	body := []byte("twelve bytes")
	srv := payloadServer(t, body)
	dir := t.TempDir()
	dest := filepath.Join(dir, "paper.jar")

	d := papermc.Download{URL: srv.URL + "/x", Size: 9999, Checksums: papermc.Checksums{SHA256: sha256Hex(body)}}
	err := NewDownloader().Download(context.Background(), d, dest, nil)
	if !errors.Is(err, ErrSizeMismatch) {
		t.Fatalf("err = %v, want ErrSizeMismatch", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Error("dest should not exist after a size mismatch")
	}
	noTempFiles(t, dir)
}

func TestDownloadContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(make([]byte, 1024))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		<-r.Context().Done() // hold the connection open until the client cancels
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	dest := filepath.Join(dir, "paper.jar")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	d := papermc.Download{URL: srv.URL + "/x", Size: 10_000_000}
	err := NewDownloader().Download(ctx, d, dest, nil)
	if err == nil {
		t.Fatal("expected an error from a cancelled download")
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Error("dest should not exist after a cancelled download")
	}
	noTempFiles(t, dir)
}

func TestDownloadProgress(t *testing.T) {
	body := make([]byte, 100_000)
	srv := payloadServer(t, body)
	dir := t.TempDir()
	dest := filepath.Join(dir, "paper.jar")

	var lastDone, lastTotal int64
	calls := 0
	onProgress := func(done, total int64) {
		calls++
		lastDone, lastTotal = done, total
	}
	if err := NewDownloader().Download(context.Background(), dl(srv, body, sha256Hex(body)), dest, onProgress); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if calls == 0 {
		t.Error("expected onProgress to be called")
	}
	if lastDone != int64(len(body)) || lastTotal != int64(len(body)) {
		t.Errorf("final progress = %d/%d, want %d/%d", lastDone, lastTotal, len(body), len(body))
	}
}
