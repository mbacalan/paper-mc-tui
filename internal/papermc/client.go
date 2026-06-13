// Package papermc is a small client for the PaperMC Fill v3 download API
// (https://fill.papermc.io/v3). It is pure HTTP + JSON with no disk I/O, takes a
// context on every call, and is configured via functional options so it can be
// pointed at an httptest server in tests.
package papermc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	// DefaultBaseURL is the Fill v3 API root. The old api.papermc.io/v2 was sunset
	// and stopped receiving builds on 2025-12-31.
	DefaultBaseURL = "https://fill.papermc.io/v3"
	// DefaultUserAgent is a fallback identifier. Fill v3 requires a descriptive,
	// non-generic User-Agent; callers should override it with a real version via
	// WithUserAgent.
	DefaultUserAgent = "paper-mc-tui (+https://github.com/mbacalan/paper-mc-tui)"
	// defaultTimeout bounds individual JSON API calls. The large jar transfer is
	// handled by the download package, which uses its own timeout.
	defaultTimeout = 15 * time.Second
)

// Client talks to the Fill v3 API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets the underlying HTTP client (useful for tests).
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) {
		if c != nil {
			cl.httpClient = c
		}
	}
}

// WithBaseURL overrides the API root (no trailing slash).
func WithBaseURL(u string) Option {
	return func(cl *Client) {
		if u != "" {
			cl.baseURL = u
		}
	}
}

// WithUserAgent sets the User-Agent header sent on every request.
func WithUserAgent(ua string) Option {
	return func(cl *Client) {
		if ua != "" {
			cl.userAgent = ua
		}
	}
}

// NewClient returns a Client with sensible defaults, overridden by opts.
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    DefaultBaseURL,
		userAgent:  DefaultUserAgent,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// doJSON performs a GET against baseURL+path and decodes the JSON body into out.
func (c *Client) doJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("papermc: build request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("papermc: request %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &StatusError{StatusCode: resp.StatusCode, URL: req.URL.String()}
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("papermc: decode %s: %w", path, err)
	}
	return nil
}
