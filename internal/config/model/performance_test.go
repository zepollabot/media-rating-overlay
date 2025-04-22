package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PerformanceTestSuite struct {
	suite.Suite
}

func TestPerformanceTestSuite(t *testing.T) {
	suite.Run(t, new(PerformanceTestSuite))
}

func (s *PerformanceTestSuite) TestDefaultPerformance() {
	cfg := DefaultPerformance()

	s.T().Run("Should return non-nil config", func(t *testing.T) {
		assert.NotNil(t, cfg)
	})
	s.T().Run("MaxThreads should be default", func(t *testing.T) {
		assert.Equal(t, 1, cfg.MaxThreads)
	})
	s.T().Run("LibraryProcessingTimeout should be default", func(t *testing.T) {
		assert.Equal(t, 600*time.Second, cfg.LibraryProcessingTimeout)
	})
}

func (s *PerformanceTestSuite) TestPerformance_Validate() {
	defaultCfg := DefaultPerformance()

	s.T().Run("Valid default config should pass", func(t *testing.T) {
		cfg := DefaultPerformance()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Negative MaxThreads should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.MaxThreads = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "performance.max_threads must be non-negative")
	})

	s.T().Run("Zero MaxThreads should pass", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.MaxThreads = 0
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Zero LibraryProcessingTimeout should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.LibraryProcessingTimeout = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "performance.library_processing_timeout must be positive")
	})

	s.T().Run("Negative LibraryProcessingTimeout should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.LibraryProcessingTimeout = -1 * time.Second
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "performance.library_processing_timeout must be positive")
	})
}
