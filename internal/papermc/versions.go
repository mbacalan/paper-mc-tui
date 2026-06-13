package papermc

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

// maxProbes bounds how many versions Resolve will query for a matching build, so a
// weird API response can't fan out into dozens of requests.
const maxProbes = 12

// Versions returns the raw grouped versions map (major version -> [versions...]).
func (c *Client) Versions(ctx context.Context) (map[string][]string, error) {
	var pr ProjectResponse
	if err := c.doJSON(ctx, "/projects/paper", &pr); err != nil {
		return nil, err
	}
	return pr.Versions, nil
}

// Resolve finds the newest version whose latest build is in one of the allowed
// channels and returns it together with that build and its server jar download.
// If no channels are given it defaults to STABLE.
//
// It walks versions newest-first using the per-version builds/latest shortcut. When
// only STABLE is allowed, pre-release versions (those with a "-rc"/"-pre" suffix) are
// skipped without a request, since they are never stable.
func (c *Client) Resolve(ctx context.Context, allowed ...Channel) (Release, error) {
	if len(allowed) == 0 {
		allowed = []Channel{ChannelStable}
	}
	stableOnly := len(allowed) == 1 && allowed[0] == ChannelStable

	grouped, err := c.Versions(ctx)
	if err != nil {
		return Release{}, err
	}

	probes := 0
	for _, version := range sortedVersions(grouped) {
		if stableOnly && isPrerelease(version) {
			continue
		}
		if probes >= maxProbes {
			break
		}
		probes++

		build, err := c.LatestBuild(ctx, version)
		if err != nil {
			// A listed version may not have any builds yet; skip those.
			var se *StatusError
			if errors.As(err, &se) && se.StatusCode == http.StatusNotFound {
				continue
			}
			return Release{}, err
		}

		if !slices.Contains(allowed, build.Channel) {
			continue
		}

		dl, ok := build.ServerDefault()
		if !ok {
			return Release{}, fmt.Errorf("papermc: version %s build %d: %w", version, build.ID, ErrNoServerDownload)
		}
		return Release{Version: version, Build: build, Download: dl}, nil
	}

	return Release{}, ErrNoStableBuild
}

// isPrerelease reports whether a version string is a release candidate or pre-release
// (e.g. "26.2-rc-2", "1.21.11-pre5"). Stable Paper versions never contain "-".
func isPrerelease(version string) bool {
	return strings.Contains(version, "-")
}

// sortedVersions flattens the grouped versions map into a single slice ordered newest
// first. We sort explicitly rather than trusting JSON/map ordering (Go maps are
// unordered).
func sortedVersions(grouped map[string][]string) []string {
	var all []string
	for _, vs := range grouped {
		all = append(all, vs...)
	}
	// Newest first: reverse the comparison arguments.
	slices.SortStableFunc(all, func(a, b string) int {
		return compareVersions(b, a)
	})
	return all
}

// compareVersions orders Paper version strings, returning -1, 0, or 1 for a<b, a==b,
// a>b. It handles the mixed numbering Paper uses today (CalVer-style "26.1.2" sorts
// above the legacy "1.21.11") and treats pre-release suffixes ("-rc"/"-pre") as lower
// precedence than the corresponding release.
func compareVersions(a, b string) int {
	aRelease, aPre, _ := strings.Cut(a, "-")
	bRelease, bPre, _ := strings.Cut(b, "-")
	return cmp.Or(
		compareRelease(aRelease, bRelease),
		comparePrerelease(aPre, bPre),
	)
}

// comparePrerelease orders pre-release suffixes: a final release (empty suffix)
// outranks any pre-release of the same version.
func comparePrerelease(a, b string) int {
	switch {
	case a == b:
		return 0
	case a == "":
		return 1
	case b == "":
		return -1
	default:
		return strings.Compare(a, b)
	}
}

// compareRelease compares the dot-separated, numeric release portion of two versions.
func compareRelease(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := range max(len(as), len(bs)) {
		if c := cmp.Compare(segment(as, i), segment(bs, i)); c != 0 {
			return c
		}
	}
	return 0
}

// segment parses the i-th release component as an int. Missing or non-numeric
// components count as 0; release components (before any "-") are numeric in practice.
func segment(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, _ := strconv.Atoi(parts[i])
	return n
}
