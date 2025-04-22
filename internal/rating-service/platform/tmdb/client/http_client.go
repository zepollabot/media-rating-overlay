package tmdb

import (
	"net/http"
	"time"

	common "github.com/zepollabot/media-rating-overlay/internal/httpclient"
)

// TMDBHTTPClient implements the HTTPClient interface for TMDB
type TMDBHTTPClient struct {
	client common.HTTPClient
}

// NewTMDBHTTPClient creates a new TMDB HTTP client
func NewTMDBHTTPClient(timeout time.Duration, maxRetries int) *TMDBHTTPClient {
	return &TMDBHTTPClient{
		client: common.NewRetryClient(timeout, maxRetries),
	}
}

// Do implements the HTTPClient interface
func (c *TMDBHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
