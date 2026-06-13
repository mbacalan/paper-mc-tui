// Package buildinfo holds build metadata stamped into the binary at build time.
//
// The Makefile sets these via -ldflags, e.g.:
//
//	-X github.com/mbacalan/paper-mc-tui/internal/buildinfo.Version=v1.2.3
//
// When built without ldflags (e.g. `go run`), the defaults below apply.
package buildinfo

var (
	// Version is the release version, typically from `git describe --tags`.
	Version = "dev"
	// Commit is the short git commit hash.
	Commit = "none"
	// Date is the UTC build timestamp.
	Date = "unknown"
)
