package papermc

import (
	"context"
	"net/url"
)

// Builds returns all builds for a version, newest first (as the API orders them).
func (c *Client) Builds(ctx context.Context, version string) ([]Build, error) {
	var builds []Build
	path := "/projects/paper/versions/" + url.PathEscape(version) + "/builds"
	if err := c.doJSON(ctx, path, &builds); err != nil {
		return nil, err
	}
	return builds, nil
}

// LatestBuild returns the most recent build for a version, regardless of channel.
func (c *Client) LatestBuild(ctx context.Context, version string) (Build, error) {
	var b Build
	path := "/projects/paper/versions/" + url.PathEscape(version) + "/builds/latest"
	if err := c.doJSON(ctx, path, &b); err != nil {
		return Build{}, err
	}
	return b, nil
}
