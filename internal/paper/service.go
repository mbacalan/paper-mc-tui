// Package paper is the application service the UI talks to. It orchestrates the Fill v3
// client, the downloader, and the local state store, keeping all network/disk/JSON
// concerns out of the views.
package paper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mbacalan/paper-mc-tui/internal/download"
	"github.com/mbacalan/paper-mc-tui/internal/papermc"
	"github.com/mbacalan/paper-mc-tui/internal/state"
)

const (
	jarName           = "paper.jar"
	defaultBackupName = "paper.backup.jar"
)

// Service ties together the API client, downloader, and state store for one target
// directory and set of allowed release channels.
type Service struct {
	client     *papermc.Client
	downloader *download.Downloader
	store      *state.Store
	dir        string
	channels   []papermc.Channel

	// cached holds the most recent resolution so Install need not query the API again
	// after CheckLatest. The UI drives these calls sequentially on one goroutine.
	cached *papermc.Release
}

// LatestInfo is a UI-friendly view of the newest available release.
type LatestInfo struct {
	Version  string
	Build    int
	JarName  string
	Channel  papermc.Channel
	Download papermc.Download
	UpToDate bool // true if the installed jar already matches this release
}

// NewService builds a Service. If no channels are given it defaults to STABLE.
func NewService(dir string, client *papermc.Client, dl *download.Downloader, store *state.Store, channels ...papermc.Channel) *Service {
	if len(channels) == 0 {
		channels = []papermc.Channel{papermc.ChannelStable}
	}
	return &Service{
		client:     client,
		downloader: dl,
		store:      store,
		dir:        dir,
		channels:   channels,
	}
}

func (s *Service) jarPath() string { return filepath.Join(s.dir, jarName) }

// CheckLatest resolves the newest available release and reports whether it is already
// installed. It refreshes the cached release used by Install.
func (s *Service) CheckLatest(ctx context.Context) (LatestInfo, error) {
	rel, err := s.client.Resolve(ctx, s.channels...)
	if err != nil {
		return LatestInfo{}, err
	}
	s.cached = &rel
	return s.infoFor(rel)
}

// Installed returns the build currently recorded as installed (zero State if none).
func (s *Service) Installed() (state.State, error) {
	return s.store.Load()
}

// JarExists reports whether a paper.jar is already present in the target directory.
func (s *Service) JarExists() bool {
	_, err := os.Stat(s.jarPath())
	return err == nil
}

// Backup renames the existing paper.jar to name (default "paper.backup.jar") within the
// target directory and returns the path it was moved to.
func (s *Service) Backup(name string) (string, error) {
	if name == "" {
		name = defaultBackupName
	}
	dest := filepath.Join(s.dir, name)
	if err := os.Rename(s.jarPath(), dest); err != nil {
		return "", fmt.Errorf("paper: backup existing jar: %w", err)
	}
	_ = s.store.Log("backed up existing %s to %s", jarName, name)
	return dest, nil
}

// Install downloads and verifies the latest release into the target directory, then
// records it in the state file. onProgress, if non-nil, receives transfer progress.
func (s *Service) Install(ctx context.Context, onProgress func(done, total int64)) error {
	rel, err := s.resolve(ctx)
	if err != nil {
		return err
	}

	_ = s.store.Log("downloading %s (build %d, %s)", rel.Download.Name, rel.Build.ID, rel.Build.Channel)
	if err := s.downloader.Download(ctx, rel.Download, s.jarPath(), onProgress); err != nil {
		_ = s.store.Log("download of %s failed: %v", rel.Download.Name, err)
		return err
	}

	st := state.State{
		Version:     rel.Version,
		Build:       rel.Build.ID,
		JarName:     rel.Download.Name,
		SHA256:      rel.Download.Checksums.SHA256,
		InstalledAt: time.Now(),
	}
	if err := s.store.Save(st); err != nil {
		return fmt.Errorf("paper: save state: %w", err)
	}
	_ = s.store.Log("installed %s", rel.Download.Name)
	return nil
}

// resolve returns the cached release if present (set by CheckLatest), otherwise queries
// the API.
func (s *Service) resolve(ctx context.Context) (papermc.Release, error) {
	if s.cached != nil {
		return *s.cached, nil
	}
	rel, err := s.client.Resolve(ctx, s.channels...)
	if err != nil {
		return papermc.Release{}, err
	}
	s.cached = &rel
	return rel, nil
}

// infoFor builds a LatestInfo and compares it against the installed state.
func (s *Service) infoFor(rel papermc.Release) (LatestInfo, error) {
	installed, err := s.store.Load()
	if err != nil {
		return LatestInfo{}, err
	}
	upToDate := installed.Version == rel.Version && installed.Build == rel.Build.ID && s.JarExists()
	return LatestInfo{
		Version:  rel.Version,
		Build:    rel.Build.ID,
		JarName:  rel.Download.Name,
		Channel:  rel.Build.Channel,
		Download: rel.Download,
		UpToDate: upToDate,
	}, nil
}
