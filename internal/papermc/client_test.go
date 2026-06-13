package papermc

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const testUserAgent = "paper-mc-tui-test (+https://example.test)"

// newTestServer serves the testdata fixtures for the Fill v3 endpoints the client
// uses. It records the User-Agent of the last request via gotUA.
func newTestServer(t *testing.T, gotUA *string) *httptest.Server {
	t.Helper()

	serve := func(file string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if gotUA != nil {
				*gotUA = r.Header.Get("User-Agent")
			}
			if r.Header.Get("User-Agent") == "" {
				http.Error(w, "missing user-agent", http.StatusBadRequest)
				return
			}
			data, err := os.ReadFile(filepath.Join("testdata", file))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/projects/paper", serve("project.json"))
	mux.HandleFunc("/projects/paper/versions/26.2-rc-2/builds/latest", serve("build_latest_beta.json"))
	mux.HandleFunc("/projects/paper/versions/26.1.2/builds/latest", serve("build_latest_stable.json"))
	mux.HandleFunc("/projects/paper/versions/1.21.10/builds", serve("builds_list.json"))
	mux.HandleFunc("/projects/paper/versions/1.21.10/builds/latest", serve("builds_list.json"))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	return NewClient(
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithUserAgent(testUserAgent),
	)
}

func TestResolveStable(t *testing.T) {
	srv := newTestServer(t, nil)
	c := newTestClient(t, srv)

	rel, err := c.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// 26.2-rc-2 is a pre-release (skipped without a request); 26.1.2 is the newest
	// version whose latest build is STABLE.
	if rel.Version != "26.1.2" {
		t.Errorf("version = %q, want 26.1.2", rel.Version)
	}
	if rel.Build.Channel != ChannelStable {
		t.Errorf("channel = %q, want STABLE", rel.Build.Channel)
	}
	if rel.Download.Name != "paper-26.1.2-70.jar" {
		t.Errorf("download name = %q, want paper-26.1.2-70.jar", rel.Download.Name)
	}
	if rel.Download.Checksums.SHA256 == "" {
		t.Error("expected a sha256 on the resolved download")
	}
}

func TestResolveExperimental(t *testing.T) {
	srv := newTestServer(t, nil)
	c := newTestClient(t, srv)

	rel, err := c.Resolve(context.Background(), ChannelStable, ChannelBeta, ChannelAlpha)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// With experimental channels allowed, the newest version (a BETA pre-release) wins.
	if rel.Version != "26.2-rc-2" {
		t.Errorf("version = %q, want 26.2-rc-2", rel.Version)
	}
	if rel.Build.Channel != ChannelBeta {
		t.Errorf("channel = %q, want BETA", rel.Build.Channel)
	}
}

func TestUserAgentSent(t *testing.T) {
	var gotUA string
	srv := newTestServer(t, &gotUA)
	c := newTestClient(t, srv)

	if _, err := c.Versions(context.Background()); err != nil {
		t.Fatalf("Versions: %v", err)
	}
	if gotUA != testUserAgent {
		t.Errorf("User-Agent = %q, want %q", gotUA, testUserAgent)
	}
}

func TestBuildsDecode(t *testing.T) {
	srv := newTestServer(t, nil)
	c := newTestClient(t, srv)

	builds, err := c.Builds(context.Background(), "1.21.10")
	if err != nil {
		t.Fatalf("Builds: %v", err)
	}
	if len(builds) != 2 {
		t.Fatalf("got %d builds, want 2", len(builds))
	}
	b := builds[0]
	if b.ID != 130 {
		t.Errorf("ID = %d, want 130", b.ID)
	}
	if b.Time.IsZero() {
		t.Error("expected build time to parse")
	}
	dl, ok := b.ServerDefault()
	if !ok {
		t.Fatal("expected a server:default download")
	}
	const wantSHA = "158703f75a26f842ea656b3dc6d75bf3d1ec176b97a2c36384d0b80b3871af53"
	if dl.Checksums.SHA256 != wantSHA {
		t.Errorf("sha256 = %q, want %q", dl.Checksums.SHA256, wantSHA)
	}
	if dl.Size != 54475623 {
		t.Errorf("size = %d, want 54475623", dl.Size)
	}
}

func TestServerDefaultAbsent(t *testing.T) {
	b := Build{Downloads: map[string]Download{"other:thing": {Name: "x"}}}
	if _, ok := b.ServerDefault(); ok {
		t.Error("expected ServerDefault to report absent")
	}
}

func TestStatusError(t *testing.T) {
	srv := newTestServer(t, nil)
	c := newTestClient(t, srv)

	// An unregistered version returns 404 from the mux.
	_, err := c.LatestBuild(context.Background(), "0.0.0")
	if err == nil {
		t.Fatal("expected an error for an unknown version")
	}
	if !errors.Is(err, ErrUnexpectedStatus) {
		t.Errorf("error %v does not match ErrUnexpectedStatus", err)
	}
	var se *StatusError
	if !errors.As(err, &se) || se.StatusCode != http.StatusNotFound {
		t.Errorf("expected a 404 StatusError, got %v", err)
	}
}
