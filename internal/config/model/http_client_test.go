package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HTTPClientTestSuite struct {
	suite.Suite
}

func TestHTTPClientTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPClientTestSuite))
}

func (s *HTTPClientTestSuite) TestDefaultHTTPClient() {
	cfg := DefaultHTTPClient()

	s.T().Run("Should return a non-nil config", func(t *testing.T) {
		assert.NotNil(t, cfg)
	})

	s.T().Run("Timeout should be default", func(t *testing.T) {
		assert.Equal(t, 30*time.Second, cfg.Timeout)
	})

	s.T().Run("MaxRetries should be default", func(t *testing.T) {
		assert.Equal(t, 3, cfg.MaxRetries)
	})
}

func (s *HTTPClientTestSuite) TestHTTPClient_Validate() {
	s.T().Run("Valid default config should pass", func(t *testing.T) {
		cfg := DefaultHTTPClient()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Invalid Timeout (zero) should fail", func(t *testing.T) {
		cfg := DefaultHTTPClient()
		cfg.Timeout = 0 // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "http_client.timeout must be positive")
	})

	s.T().Run("Invalid Timeout (negative) should fail", func(t *testing.T) {
		cfg := DefaultHTTPClient()
		cfg.Timeout = -1 * time.Second // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "http_client.timeout must be positive")
	})

	s.T().Run("Invalid MaxRetries (negative) should fail", func(t *testing.T) {
		cfg := DefaultHTTPClient()
		cfg.MaxRetries = -1 // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "http_client.max_retries must be non-negative")
	})

	s.T().Run("Valid MaxRetries (zero) should pass", func(t *testing.T) {
		cfg := DefaultHTTPClient()
		cfg.MaxRetries = 0
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
