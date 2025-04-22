package model

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PosterConfigTestSuite struct {
	suite.Suite
}

func TestPosterConfigTestSuite(t *testing.T) {
	suite.Run(t, new(PosterConfigTestSuite))
}

func (s *PosterConfigTestSuite) TestPosterConfigWithDefaultValues() {
	cfg := PosterConfigWithDefaultValues()

	s.T().Run("Should return a non-nil config", func(t *testing.T) {
		assert.NotNil(t, cfg)
	})

	s.T().Run("Dimensions should have default values", func(t *testing.T) {
		assert.Equal(t, 1200, cfg.Dimensions.Width)
		assert.Equal(t, 1800, cfg.Dimensions.Height)
	})

	s.T().Run("Margins should have default values", func(t *testing.T) {
		assert.Equal(t, 20, cfg.Margins.Left)
		assert.Equal(t, 20, cfg.Margins.Right)
	})

	s.T().Run("ImagePaths for RottenTomatoes Critic should have default values", func(t *testing.T) {
		expectedNormal := filepath.Join("internal", "processor", "image", "data", "RT_critic.png")
		expectedLow := filepath.Join("internal", "processor", "image", "data", "RT_critic_low.png")
		assert.Equal(t, expectedNormal, cfg.ImagePaths.RottenTomatoes.Critic.Normal)
		assert.Equal(t, expectedLow, cfg.ImagePaths.RottenTomatoes.Critic.Low)
	})

	s.T().Run("ImagePaths for RottenTomatoes Audience should have default values", func(t *testing.T) {
		expectedNormal := filepath.Join("internal", "processor", "image", "data", "RT_audience.png")
		expectedLow := filepath.Join("internal", "processor", "image", "data", "RT_audience_low.png")
		assert.Equal(t, expectedNormal, cfg.ImagePaths.RottenTomatoes.Audience.Normal)
		assert.Equal(t, expectedLow, cfg.ImagePaths.RottenTomatoes.Audience.Low)
	})

	s.T().Run("ImagePaths for IMDB Audience should have default values", func(t *testing.T) {
		expectedNormal := filepath.Join("internal", "processor", "image", "data", "IMDb.png")
		assert.Equal(t, expectedNormal, cfg.ImagePaths.IMDB.Audience.Normal)
	})

	s.T().Run("ImagePaths for TMDB Audience should have default values", func(t *testing.T) {
		expectedNormal := filepath.Join("internal", "processor", "image", "data", "TMDB.png")
		assert.Equal(t, expectedNormal, cfg.ImagePaths.TMDB.Audience.Normal)
	})

	s.T().Run("VisualDebug should be false by default", func(t *testing.T) {
		assert.False(t, cfg.VisualDebug)
	})
}
