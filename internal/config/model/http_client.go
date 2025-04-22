package config

import (
	"fmt"
	"time"
)

type HTTPClient struct {
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
}

func DefaultHTTPClient() *HTTPClient {
	return &HTTPClient{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}
}

// Validate validates the HTTPClient configuration
func (c *HTTPClient) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("http_client.timeout must be positive")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("http_client.max_retries must be non-negative")
	}
	return nil
}
