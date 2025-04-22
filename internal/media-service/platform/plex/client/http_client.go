package plex

import (
	"net/http"
	"time"

	common "github.com/zepollabot/media-rating-overlay/internal/httpclient"
)

// PlexHTTPClient implements the HTTPClient interface for Plex
type PlexHTTPClient struct {
	client common.HTTPClient
}

// NewPlexHTTPClient creates a new Plex HTTP client
func NewPlexHTTPClient(timeout time.Duration, maxRetries int) *PlexHTTPClient {
	return &PlexHTTPClient{
		client: common.NewRetryClient(timeout, maxRetries),
	}
}

// Do implements the HTTPClient interface
func (c *PlexHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
