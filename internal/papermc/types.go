package papermc

import "time"

// Channel is the release channel of a build, as reported by Fill v3.
type Channel string

const (
	ChannelStable Channel = "STABLE"
	ChannelBeta   Channel = "BETA"
	ChannelAlpha  Channel = "ALPHA"
)

// downloadKeyServerDefault is the key under a build's "downloads" map that holds
// the standard server jar. Modeling downloads as a map keeps us forward-compatible
// if Paper adds more artifacts (e.g. mappings).
const downloadKeyServerDefault = "server:default"

// ProjectResponse is the body of GET /v3/projects/{project}.
type ProjectResponse struct {
	Project  ProjectInfo         `json:"project"`
	Versions map[string][]string `json:"versions"` // major version -> [versions...]
}

type ProjectInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Build is one element of GET .../builds, or the body of GET .../builds/latest.
type Build struct {
	ID        int                 `json:"id"`
	Channel   Channel             `json:"channel"`
	Time      time.Time           `json:"time"`
	Commits   []Commit            `json:"commits"`
	Downloads map[string]Download `json:"downloads"`
}

type Commit struct {
	SHA     string    `json:"sha"`
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

// Download describes a single downloadable artifact of a build.
type Download struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	Checksums Checksums `json:"checksums"`
}

type Checksums struct {
	SHA256 string `json:"sha256"`
}

// ServerDefault returns the standard server jar download for the build, if present.
func (b Build) ServerDefault() (Download, bool) {
	d, ok := b.Downloads[downloadKeyServerDefault]
	return d, ok
}

// Release is a fully resolved "what to install": a version, the build chosen for it,
// and that build's server jar download.
type Release struct {
	Version  string
	Build    Build
	Download Download
}
