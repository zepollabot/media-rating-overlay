package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MetacriticTestSuite struct {
	suite.Suite
}

func TestMetacriticTestSuite(t *testing.T) {
	suite.Run(t, new(MetacriticTestSuite))
}

func (s *MetacriticTestSuite) TestDefaultMetacritic() {
	cfg := DefaultMetacritic()

	s.T().Run("Should return non-nil config", func(t *testing.T) {
		assert.NotNil(t, cfg)
	})
	s.T().Run("Enabled should be false by default", func(t *testing.T) {
		assert.False(t, cfg.Enabled)
	})
	s.T().Run("ApiKey should be empty by default", func(t *testing.T) {
		assert.Empty(t, cfg.APIKey)
	})
}

func (s *MetacriticTestSuite) TestMetacritic_Validate() {
	defaultCfg := DefaultMetacritic()

	s.T().Run("Valid default config (disabled) should pass", func(t *testing.T) {
		cfg := DefaultMetacritic()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Enabled with empty ApiKey should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.Enabled = true
		cfg.APIKey = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "metacritic.api_key is required when metacritic is enabled")
	})

	s.T().Run("Enabled with ApiKey should pass", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.Enabled = true
		cfg.APIKey = "an-api-key"
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Disabled with ApiKey should pass", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.Enabled = false
		cfg.APIKey = "an-api-key"
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
