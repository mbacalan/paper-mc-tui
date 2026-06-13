package paper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mbacalan/paper-mc-tui/internal/download"
	"github.com/mbacalan/paper-mc-tui/internal/papermc"
	"github.com/mbacalan/paper-mc-tui/internal/state"
)

// newServiceFixture spins up an httptest server that serves the Fill v3 endpoints and
// the jar object, wired to a Service rooted in a temp dir.
func newServiceFixture(t *testing.T) (*Service, string, []byte) {
	t.Helper()

	payload := []byte("pretend this is a 55MB paper server jar")
	sum := sha256.Sum256(payload)
	sha := hex.EncodeToString(sum[:])

	mux := http.NewServeMux()
	mux.HandleFunc("/projects/paper", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"project":{"id":"paper","name":"Paper"},"versions":{"26.2":["26.2-rc-2"],"26.1":["26.1.2"]}}`)
	})
	mux.HandleFunc("/jar", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	// builds/latest for 26.1.2 references the jar object on the same server.
	mux.HandleFunc("/projects/paper/versions/26.1.2/builds/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":70,"time":"2026-05-20T10:00:00Z","channel":"STABLE","commits":[],
			"downloads":{"server:default":{"name":"paper-26.1.2-70.jar","size":%d,
			"checksums":{"sha256":"%s"},"url":"%s/jar"}}}`, len(payload), sha, srv.URL)
	})

	dir := t.TempDir()
	store, err := state.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	client := papermc.NewClient(
		papermc.WithBaseURL(srv.URL),
		papermc.WithHTTPClient(srv.Client()),
		papermc.WithUserAgent("paper-mc-tui-test (+https://example.test)"),
	)
	dl := download.NewDownloader(download.WithHTTPClient(srv.Client()))

	return NewService(dir, client, dl, store, papermc.ChannelStable), dir, payload
}

func TestServiceInstallThenUpToDate(t *testing.T) {
	svc, dir, payload := newServiceFixture(t)
	ctx := context.Background()

	info, err := svc.CheckLatest(ctx)
	if err != nil {
		t.Fatalf("CheckLatest: %v", err)
	}
	if info.Version != "26.1.2" || info.Build != 70 {
		t.Errorf("got %s build %d, want 26.1.2 build 70", info.Version, info.Build)
	}
	if info.UpToDate {
		t.Error("nothing installed yet; UpToDate should be false")
	}

	if err := svc.Install(ctx, nil); err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Jar written with the right content.
	got, err := os.ReadFile(filepath.Join(dir, "paper.jar"))
	if err != nil {
		t.Fatalf("read jar: %v", err)
	}
	if string(got) != string(payload) {
		t.Errorf("jar content mismatch")
	}

	// State recorded.
	st, _ := svc.Installed()
	if st.Version != "26.1.2" || st.Build != 70 || st.JarName != "paper-26.1.2-70.jar" || st.SHA256 == "" {
		t.Errorf("unexpected state: %+v", st)
	}

	// Now it should report up to date.
	info2, err := svc.CheckLatest(ctx)
	if err != nil {
		t.Fatalf("CheckLatest #2: %v", err)
	}
	if !info2.UpToDate {
		t.Error("expected UpToDate after install")
	}
}

func TestServiceBackup(t *testing.T) {
	svc, dir, _ := newServiceFixture(t)
	jar := filepath.Join(dir, "paper.jar")
	if err := os.WriteFile(jar, []byte("old jar"), 0o644); err != nil {
		t.Fatal(err)
	}

	dest, err := svc.Backup("")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if filepath.Base(dest) != "paper.backup.jar" {
		t.Errorf("backup path = %s, want .../paper.backup.jar", dest)
	}
	if _, err := os.Stat(jar); !os.IsNotExist(err) {
		t.Error("original paper.jar should be gone after backup")
	}
	got, _ := os.ReadFile(dest)
	if string(got) != "old jar" {
		t.Errorf("backup content = %q, want 'old jar'", got)
	}
}
