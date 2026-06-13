package papermc

import (
	"errors"
	"fmt"
)

var (
	// ErrNoStableBuild means no version with a build in an allowed channel was found.
	ErrNoStableBuild = errors.New("papermc: no build found in an allowed channel")
	// ErrNoServerDownload means the chosen build has no "server:default" artifact.
	ErrNoServerDownload = errors.New("papermc: build has no server:default download")
	// ErrUnexpectedStatus is matched by StatusError via errors.Is.
	ErrUnexpectedStatus = errors.New("papermc: unexpected status code")
)

// StatusError is returned when the API responds with a non-200 status. It carries
// the code and URL for diagnostics and matches ErrUnexpectedStatus.
type StatusError struct {
	StatusCode int
	URL        string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("papermc: unexpected status %d for %s", e.StatusCode, e.URL)
}

func (e *StatusError) Is(target error) bool {
	return target == ErrUnexpectedStatus
}
