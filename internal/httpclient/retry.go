package common

import (
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/httpclient"
)

const (
	backoffInterval       = 2 * time.Millisecond
	maximumJitterInterval = 100 * time.Millisecond
)

// NewRetryClient creates a new HTTP client with retry mechanism
func NewRetryClient(timeout time.Duration, maxRetries int) *httpclient.Client {
	backoff := heimdall.NewConstantBackoff(backoffInterval, maximumJitterInterval)

	// Create a new retry mechanism with the backoff
	retrier := heimdall.NewRetrier(backoff)

	// Create a new client, sets the retry mechanism, and the number of times you would like to retry
	return httpclient.NewClient(
		httpclient.WithHTTPTimeout(timeout),
		httpclient.WithRetrier(retrier),
		httpclient.WithRetryCount(maxRetries),
	)
}
