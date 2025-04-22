package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LoggerTestSuite struct {
	suite.Suite
}

func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

func (s *LoggerTestSuite) TestDefaultLogger() {
	cfg := DefaultLogger()

	s.T().Run("Should return non-nil config", func(t *testing.T) {
		assert.NotNil(t, cfg)
	})
	s.T().Run("LogFilePath should be default", func(t *testing.T) {
		assert.Equal(t, "logs/media-rating-overlay.log", cfg.LogFilePath)
	})
	s.T().Run("LogLevel should be default", func(t *testing.T) {
		assert.Equal(t, "info", cfg.LogLevel)
	})
	s.T().Run("MaxSize should be default", func(t *testing.T) {
		assert.Equal(t, 100, cfg.MaxSize)
	})
	s.T().Run("MaxBackups should be default", func(t *testing.T) {
		assert.Equal(t, 5, cfg.MaxBackups)
	})
	s.T().Run("MaxAge should be default", func(t *testing.T) {
		assert.Equal(t, 30, cfg.MaxAge)
	})
	s.T().Run("Compress should be default", func(t *testing.T) {
		assert.True(t, cfg.Compress)
	})
	s.T().Run("ServiceName should be default", func(t *testing.T) {
		assert.Equal(t, "media-rating-overlay", cfg.ServiceName)
	})
	s.T().Run("ServiceVersion should be default", func(t *testing.T) {
		assert.Equal(t, "1.0.0", cfg.ServiceVersion)
	})
	s.T().Run("UseJSON should be default", func(t *testing.T) {
		assert.True(t, cfg.UseJSON)
	})
	s.T().Run("UseStdout should be default", func(t *testing.T) {
		assert.True(t, cfg.UseStdout)
	})
}

func (s *LoggerTestSuite) TestLogger_Validate() {
	defaultCfg := DefaultLogger()

	s.T().Run("Valid default config should pass", func(t *testing.T) {
		cfg := DefaultLogger() // Use a fresh default for each sub-test
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Empty LogFilePath should fail", func(t *testing.T) {
		cfg := *defaultCfg // Create a mutable copy
		cfg.LogFilePath = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "logger.log_file_path is required")
	})

	s.T().Run("Zero MaxSize should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.MaxSize = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "logger.max_size must be positive")
	})

	s.T().Run("Negative MaxSize should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.MaxSize = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "logger.max_size must be positive")
	})

	s.T().Run("Negative MaxBackups should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.MaxBackups = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "logger.max_backups must be non-negative")
	})

	s.T().Run("Zero MaxBackups should pass", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.MaxBackups = 0
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Negative MaxAge should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.MaxAge = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "logger.max_age must be non-negative")
	})

	s.T().Run("Zero MaxAge should pass", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.MaxAge = 0
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Empty ServiceName should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.ServiceName = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "logger.service_name is required")
	})

	s.T().Run("Empty ServiceVersion should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.ServiceVersion = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "logger.service_version is required")
	})
}
