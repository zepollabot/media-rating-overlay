package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PlexTestSuite struct {
	suite.Suite
}

func TestPlexTestSuite(t *testing.T) {
	suite.Run(t, new(PlexTestSuite))
}

func (s *PlexTestSuite) TestDefaultPlex() {
	cfg := DefaultPlex()

	s.T().Run("Should return non-nil config", func(t *testing.T) {
		assert.NotNil(t, cfg)
	})
	s.T().Run("Enabled should be false by default", func(t *testing.T) {
		assert.False(t, cfg.Enabled)
	})
	s.T().Run("Libraries should be empty by default", func(t *testing.T) {
		assert.Empty(t, cfg.Libraries)
	})
	s.T().Run("URL should be empty by default", func(t *testing.T) {
		assert.Empty(t, cfg.Url)
	})
	s.T().Run("Token should be empty by default", func(t *testing.T) {
		assert.Empty(t, cfg.Token)
	})
}

func (s *PlexTestSuite) TestPlex_Validate() {
	defaultCfg := DefaultPlex()

	s.T().Run("Valid default config (disabled) should pass", func(t *testing.T) {
		cfg := DefaultPlex()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Enabled with empty URL should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.Enabled = true
		cfg.Url = ""
		cfg.Token = "a-token"
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "plex.url is required when plex is enabled")
	})

	s.T().Run("Enabled with empty Token should fail", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.Enabled = true
		cfg.Url = "http://localhost"
		cfg.Token = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.EqualError(t, err, "plex.token is required when plex is enabled")
	})

	s.T().Run("Enabled with URL and Token should pass", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.Enabled = true
		cfg.Url = "http://localhost"
		cfg.Token = "a-token"
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	s.T().Run("Disabled with URL and Token should pass", func(t *testing.T) {
		cfg := *defaultCfg
		cfg.Enabled = false
		cfg.Url = "http://localhost"
		cfg.Token = "a-token"
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}
