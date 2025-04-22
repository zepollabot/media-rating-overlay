package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TMDBTestSuite struct {
	suite.Suite
}

func TestTMDBTestSuite(t *testing.T) {
	suite.Run(t, new(TMDBTestSuite))
}

func (s *TMDBTestSuite) TestDefaultTMDB() {
	cfg := DefaultTMDB()

	s.T().Run("Should return non-nil config", func(t *testing.T) {
		assert.NotNil(t, cfg)
	})
	s.T().Run("Enabled should be false by default", func(t *testing.T) {
		assert.False(t, cfg.Enabled)
	})
	s.T().Run("ApiKey should be empty by default", func(t *testing.T) {
		assert.Empty(t, cfg.ApiKey)
	})
	s.T().Run("Language should be en-US by default", func(t *testing.T) {
		assert.Equal(t, "en-US", cfg.Language)
	})
	s.T().Run("Region should be US by default", func(t *testing.T) {
		assert.Equal(t, "US", cfg.Region)
	})
}

func (s *TMDBTestSuite) TestTMDB_Validate() {
	defaultCfg := DefaultTMDB()

	s.T().Run("Valid default config (disabled) should pass", func(t *testing.T) {
		cfg := DefaultTMDB()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Enabled with empty ApiKey should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.Enabled = true
		cfg.ApiKey = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "tmdb.api_key is required when tmdb is enabled")
	})

	s.T().Run("Enabled with ApiKey should pass", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.Enabled = true
		cfg.ApiKey = "an-api-key"
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Disabled with ApiKey should pass", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.Enabled = false
		cfg.ApiKey = "an-api-key"
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
