package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	want := State{
		Version:     "26.1.2",
		Build:       70,
		JarName:     "paper-26.1.2-70.jar",
		SHA256:      "bbbb",
		InstalledAt: time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC),
	}
	if err := s.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Version != want.Version || got.Build != want.Build || got.JarName != want.JarName || got.SHA256 != want.SHA256 {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, want)
	}
	if !got.InstalledAt.Equal(want.InstalledAt) {
		t.Errorf("InstalledAt = %v, want %v", got.InstalledAt, want.InstalledAt)
	}
}

func TestLoadMissingReturnsZero(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load on missing file should not error: %v", err)
	}
	if got.Build != 0 || got.Version != "" {
		t.Errorf("expected zero State, got %+v", got)
	}
}

func TestSaveLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	if err := s.Save(State{Version: "1.21.10", Build: 130}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

func TestLogAppends(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	if err := s.Log("downloaded %s", "paper-1.21.10-130.jar"); err != nil {
		t.Fatalf("Log: %v", err)
	}
	if err := s.Log("second line"); err != nil {
		t.Fatalf("Log: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, logFileName))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "paper-1.21.10-130.jar") || !strings.Contains(got, "second line") {
		t.Errorf("log missing expected content:\n%s", got)
	}
	if n := strings.Count(strings.TrimSpace(got), "\n"); n != 1 { // 2 lines => 1 interior newline
		t.Errorf("expected 2 log lines, got content:\n%s", got)
	}
}
