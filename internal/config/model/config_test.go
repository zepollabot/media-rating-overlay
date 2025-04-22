package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) TestDefaultConfig() {
	cfg := DefaultConfig()

	s.T().Run("Should return a non-nil config", func(t *testing.T) {
		assert.NotNil(t, cfg)
	})

	s.T().Run("Plex should be default", func(t *testing.T) {
		assert.Equal(t, DefaultPlex(), &cfg.Plex)
	})

	s.T().Run("TMDB should be default", func(t *testing.T) {
		assert.Equal(t, DefaultTMDB(), &cfg.TMDB)
	})

	s.T().Run("Performance should be default", func(t *testing.T) {
		assert.Equal(t, DefaultPerformance(), &cfg.Performance)
	})

	s.T().Run("HTTPClient should be default", func(t *testing.T) {
		assert.Equal(t, DefaultHTTPClient(), &cfg.HTTPClient)
	})

	s.T().Run("Logger should be default", func(t *testing.T) {
		assert.Equal(t, DefaultLogger(), &cfg.Logger)
	})

	s.T().Run("Processor should be default", func(t *testing.T) {
		assert.Equal(t, DefaultProcessorConfig(), &cfg.Processor)
	})
}

func (s *ConfigTestSuite) TestConfig_Validate() {
	s.T().Run("Valid default config should pass", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Invalid Plex config should fail", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Plex.Enabled = true
		cfg.Plex.Url = "" // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plex config: plex.url is required when plex is enabled")
	})

	s.T().Run("Invalid TMDB config should fail", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.TMDB.Enabled = true
		cfg.TMDB.ApiKey = "" // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tmdb config: tmdb.api_key is required when tmdb is enabled")
	})

	s.T().Run("Invalid Performance config should fail (MaxThreads)", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Performance.MaxThreads = -1 // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "performance config: performance.max_threads must be non-negative")
	})

	s.T().Run("Invalid Performance config should fail (LibraryProcessingTimeout)", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Performance.LibraryProcessingTimeout = 0 // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "performance config: performance.library_processing_timeout must be positive")
	})

	s.T().Run("Invalid HTTPClient config should fail (Timeout)", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.HTTPClient.Timeout = 0 // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http_client config: http_client.timeout must be positive")
	})

	s.T().Run("Invalid HTTPClient config should fail (MaxRetries)", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.HTTPClient.MaxRetries = -1 // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http_client config: http_client.max_retries must be non-negative")
	})

	s.T().Run("Invalid Logger config should fail (LogFilePath)", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Logger.LogFilePath = "" // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "logger config: logger.log_file_path is required")
	})

	s.T().Run("Invalid Logger config should fail (MaxSize)", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Logger.MaxSize = 0 // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "logger config: logger.max_size must be positive")
	})

	s.T().Run("Invalid ProcessorConfig should fail (RatingBuilder.Timeout)", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Processor.ItemProcessor.RatingBuilder.Timeout = 0 // Invalid state
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "processor config: rating_builder.timeout must be greater than 0")
	})
}
