package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProcessorConfigTestSuite struct {
	suite.Suite
}

func TestProcessorConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorConfigTestSuite))
}

func (s *ProcessorConfigTestSuite) TestDefaultProcessorConfig() {
	cfg := DefaultProcessorConfig()

	s.T().Run("Should return non-nil config", func(t *testing.T) {
		assert.NotNil(t, cfg)
	})

	// ItemProcessor defaults
	s.T().Run("ItemProcessor.RatingBuilder.Timeout should be default", func(t *testing.T) {
		assert.Equal(t, 30*time.Second, cfg.ItemProcessor.RatingBuilder.Timeout)
	})

	// LibraryProcessor defaults
	s.T().Run("LibraryProcessor.DefaultTimeout should be default", func(t *testing.T) {
		assert.Equal(t, 600*time.Second, cfg.LibraryProcessor.DefaultTimeout)
	})
}

func (s *ProcessorConfigTestSuite) TestProcessorConfig_Validate() {
	newDefaultConfig := func() ProcessorConfig {
		return *DefaultProcessorConfig()
	}

	s.T().Run("Valid default config should pass", func(t *testing.T) {
		cfg := DefaultProcessorConfig()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	// ItemProcessor validation tests
	s.T().Run("ItemProcessor.RatingBuilder.Timeout zero should fail", func(t *testing.T) {
		cfg := newDefaultConfig()
		cfg.ItemProcessor.RatingBuilder.Timeout = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "rating_builder.timeout must be greater than 0")
	})
	s.T().Run("ItemProcessor.RatingBuilder.Timeout negative should fail", func(t *testing.T) {
		cfg := newDefaultConfig()
		cfg.ItemProcessor.RatingBuilder.Timeout = -1 * time.Second
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "rating_builder.timeout must be greater than 0")
	})

	// LibraryProcessor validation tests
	s.T().Run("LibraryProcessor.DefaultTimeout zero should fail", func(t *testing.T) {
		cfg := newDefaultConfig()
		cfg.LibraryProcessor.DefaultTimeout = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "default_timeout must be greater than 0")
	})
	s.T().Run("LibraryProcessor.DefaultTimeout negative should fail", func(t *testing.T) {
		cfg := newDefaultConfig()
		cfg.LibraryProcessor.DefaultTimeout = -1 * time.Second
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "default_timeout must be greater than 0")
	})

}
