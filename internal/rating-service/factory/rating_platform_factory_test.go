package factory_test

import (
	"testing"

	// "github.com/stretchr/testify/assert" // Removed unused import
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	configModel "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/constant"
	"github.com/zepollabot/media-rating-overlay/internal/rating-service/factory"
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

type RatingServiceBaseFactorySuite struct {
	suite.Suite
	logger      *zap.Logger
	config      *configModel.Config
	baseFactory *factory.RatingServiceBaseFactory
}

func (s *RatingServiceBaseFactorySuite) SetupTest() {
	s.logger = zaptest.NewLogger(s.T())
	s.config = &configModel.Config{
		TMDB: configModel.TMDB{
			Enabled: false,        // Default to disabled, enable in specific tests
			ApiKey:  "testapikey", // Corrected: ApiKey (capital K)
		},
		HTTPClient: configModel.HTTPClient{},
		Logger:     configModel.Logger{LogFilePath: "test.log"}, // Add a dummy log path
	}
	s.baseFactory = factory.NewRatingServiceBaseFactory(s.logger, s.config)
}

func TestRatingServiceBaseFactorySuite(t *testing.T) {
	suite.Run(t, new(RatingServiceBaseFactorySuite))
}

func (s *RatingServiceBaseFactorySuite) TestNewRatingServiceBaseFactory() {
	// Arrange
	logger := s.logger
	config := s.config

	// Act
	f := factory.NewRatingServiceBaseFactory(logger, config)

	// Assert
	s.NotNil(f, "Factory should not be nil")
	s.Equal(logger, f.Logger, "Logger should be set correctly")
	s.Equal(config, f.Config, "Config should be set correctly")
}

func (s *RatingServiceBaseFactorySuite) TestBuildTMDBComponents_WhenDisabled() {
	// Arrange
	s.config.TMDB.Enabled = false // Ensure TMDB is disabled
	// Re-initialize factory with updated config if necessary, or ensure SetupTest covers this
	f := factory.NewRatingServiceBaseFactory(s.logger, s.config)

	// Act
	tmdbService, err := f.BuildTMDBComponents()

	// Assert
	s.NoError(err, "BuildTMDBComponents should not return an error when TMDB is disabled")
	s.Equal(constant.RatingServiceTMDB, tmdbService.Name, "Service name should be TMDB")
	s.Nil(tmdbService.PlatformService, "PlatformService should be nil when TMDB is disabled")
}

func (s *RatingServiceBaseFactorySuite) TestBuildTMDBComponents_WhenEnabled() {
	// Arrange
	s.config.TMDB.Enabled = true
	// s.config.TMDB.ApiKey is already "testapikey" from SetupTest, which is valid for this path
	factoryInstance := factory.NewRatingServiceBaseFactory(s.logger, s.config)

	// Act
	tmdbService, err := factoryInstance.BuildTMDBComponents()

	// Assert
	s.NoError(err, "BuildTMDBComponents should not return an error when TMDB is enabled with valid config")
	s.Equal(constant.RatingServiceTMDB, tmdbService.Name, "Service name should be TMDB")
	s.NotNil(tmdbService.PlatformService, "PlatformService should not be nil when TMDB is enabled")
}

func (s *RatingServiceBaseFactorySuite) TestBuildTMDBComponents_WhenEnabled_ClientCreationError() {
	// Arrange
	s.config.TMDB.Enabled = true
	// s.config.TMDB.ApiKey can remain as "testapikey", it's not the cause of client creation error
	s.config.Logger.LogFilePath = "" // Induce an error in common.SetupLogging by providing an empty LogFilePath
	factoryInstance := factory.NewRatingServiceBaseFactory(s.logger, s.config)

	// Act
	tmdbService, err := factoryInstance.BuildTMDBComponents()

	// Assert
	s.Error(err, "BuildTMDBComponents should return an error when TMDB client creation fails due to logging setup")
	s.Equal(ratingModel.RatingService{}, tmdbService, "Returned service should be empty on client creation error")
	s.Empty(tmdbService.Name, "Service name should be empty on error")
	s.Nil(tmdbService.PlatformService, "PlatformService should be nil on error")
}

func (s *RatingServiceBaseFactorySuite) TestBuildRottenTomatoesComponents() {
	// Arrange
	f := s.baseFactory // Use the factory initialized in SetupTest

	// Act
	rottenTomatoesService, err := f.BuildRottenTomatoesComponents()

	// Assert
	s.NoError(err, "BuildRottenTomatoesComponents should not return an error")
	s.Equal(constant.RatingServiceRottenTomatoes, rottenTomatoesService.Name, "Service name should be RottenTomatoes")
	s.Nil(rottenTomatoesService.PlatformService, "PlatformService should be nil for RottenTomatoes")
}

func (s *RatingServiceBaseFactorySuite) TestBuildIMDBComponents() {
	// Arrange
	f := s.baseFactory // Use the factory initialized in SetupTest

	// Act
	imdbService, err := f.BuildIMDBComponents()

	// Assert
	s.NoError(err, "BuildIMDBComponents should not return an error")
	s.Equal(constant.RatingServiceIMDB, imdbService.Name, "Service name should be IMDB")
	s.Nil(imdbService.PlatformService, "PlatformService should be nil for IMDB")
}
