package metacritic

import (
	"net/http"
	"time"

	common "github.com/zepollabot/media-rating-overlay/internal/httpclient"
)

// MetacriticHTTPClient implements the HTTPClient interface for Metacritic
type MetacriticHTTPClient struct {
	client common.HTTPClient
}

// NewMetacriticHTTPClient creates a new Metacritic HTTP client
func NewMetacriticHTTPClient(timeout time.Duration, maxRetries int) *MetacriticHTTPClient {
	return &MetacriticHTTPClient{
		client: common.NewRetryClient(timeout, maxRetries),
	}
}

// Do implements the HTTPClient interface
func (c *MetacriticHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
