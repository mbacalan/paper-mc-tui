// Package state persists what build is currently installed and keeps a human-readable
// activity log, both alongside the jar in the target directory. It replaces the old
// bare-string logs/paper-ver.txt with a structured, atomically-written JSON file.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	stateFileName = "state.json"
	logFileName   = "paper-mc.log"
)

// State records the build last installed by this tool.
type State struct {
	Version     string    `json:"version"`      // e.g. "26.1.2"
	Build       int       `json:"build"`        // e.g. 70
	JarName     string    `json:"jar_name"`     // e.g. "paper-26.1.2-70.jar"
	SHA256      string    `json:"sha256"`       // verified checksum of the jar
	InstalledAt time.Time `json:"installed_at"` // when it was downloaded
}

// Store reads and writes State and the activity log in a directory.
type Store struct {
	dir       string
	statePath string
	logPath   string
}

// NewStore ensures dir exists and returns a Store rooted there.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("state: create dir %s: %w", dir, err)
	}
	return &Store{
		dir:       dir,
		statePath: filepath.Join(dir, stateFileName),
		logPath:   filepath.Join(dir, logFileName),
	}, nil
}

// Load returns the saved State. A missing file is not an error: it returns the zero
// State (Build == 0), which represents "nothing installed yet".
func (s *Store) Load() (State, error) {
	data, err := os.ReadFile(s.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("state: read %s: %w", s.statePath, err)
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return State{}, fmt.Errorf("state: parse %s: %w", s.statePath, err)
	}
	return st, nil
}

// Save writes State atomically (temp file + rename) so a crash mid-write never leaves
// a truncated state.json.
func (s *Store) Save(st State) error {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("state: marshal: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(s.dir, ".state-*.json.tmp")
	if err != nil {
		return fmt.Errorf("state: create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once renamed

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("state: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("state: close temp: %w", err)
	}
	if err := os.Rename(tmpName, s.statePath); err != nil {
		return fmt.Errorf("state: rename temp: %w", err)
	}
	return nil
}

// Log appends a timestamped line to the activity log. It is best-effort context for a
// human reading the log later, not the source of truth (that is state.json).
func (s *Store) Log(format string, args ...any) error {
	line := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	f, err := os.OpenFile(s.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("state: open log: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("state: write log: %w", err)
	}
	return nil
}
