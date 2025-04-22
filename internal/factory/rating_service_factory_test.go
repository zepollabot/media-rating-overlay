package factory

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/zepollabot/media-rating-overlay/internal/constant"
	factory_mocks "github.com/zepollabot/media-rating-overlay/internal/factory/mocks"
	rating_service_model "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
	"go.uber.org/zap"
)

type RatingPlatformServiceModelFactorySuite struct {
	suite.Suite
	mockBaseFactory *factory_mocks.RatingServiceBaseFactory
	logger          *zap.Logger
	factory         *RatingPlatformServiceModelFactory
}

func (s *RatingPlatformServiceModelFactorySuite) SetupTest() {
	s.mockBaseFactory = factory_mocks.NewRatingServiceBaseFactory(s.T())
	s.logger = zap.NewNop()
	s.factory = NewRatingPlatformServiceModelFactory(s.logger, s.mockBaseFactory, false)
}

func (s *RatingPlatformServiceModelFactorySuite) TearDownTest() {
	s.mockBaseFactory.AssertExpectations(s.T())
}

func TestRatingPlatformServiceModelFactorySuite(t *testing.T) {
	suite.Run(t, new(RatingPlatformServiceModelFactorySuite))
}

func (s *RatingPlatformServiceModelFactorySuite) TestNewRatingPlatformServiceModelFactory() {
	// Arrange
	logger := zap.NewNop()
	mockBase := factory_mocks.NewRatingServiceBaseFactory(s.T())
	visualDebug := false

	// Act
	f := NewRatingPlatformServiceModelFactory(logger, mockBase, visualDebug)

	// Assert
	s.NotNil(f)
}

// TestCreate_TMDBSuccess verifies that the TMDB rating service is created correctly.
func (s *RatingPlatformServiceModelFactorySuite) TestCreate_TMDBSuccess() {
	// Arrange
	baseTMDBService := rating_service_model.RatingService{Name: constant.RatingServiceTMDB}
	s.mockBaseFactory.On("BuildTMDBComponents").Return(baseTMDBService, nil).Once()

	// Act
	ratingService, err := s.factory.Create(constant.RatingServiceTMDB)

	// Assert
	s.NoError(err)
	s.NotNil(ratingService, "RatingService should not be nil")
	s.Equal(constant.RatingServiceTMDB, ratingService.Name, "Service name should be TMDB")
	// Verify that the factory has indeed assigned the LogoService to the service returned by the base factory.
	s.NotNil(ratingService.LogoService, "LogoService should be initialized for TMDB")
}

// TestCreate_RottenTomatoesSuccess verifies that the Rotten Tomatoes rating service is created correctly.
func (s *RatingPlatformServiceModelFactorySuite) TestCreate_RottenTomatoesSuccess() {
	// Arrange
	baseRTService := rating_service_model.RatingService{Name: constant.RatingServiceRottenTomatoes}
	s.mockBaseFactory.On("BuildRottenTomatoesComponents").Return(baseRTService, nil).Once()

	// Act
	ratingService, err := s.factory.Create(constant.RatingServiceRottenTomatoes)

	// Assert
	s.NoError(err)
	s.NotNil(ratingService, "RatingService should not be nil")
	s.Equal(constant.RatingServiceRottenTomatoes, ratingService.Name, "Service name should be Rotten Tomatoes")
	s.NotNil(ratingService.LogoService, "LogoService should be initialized for Rotten Tomatoes")
}

// TestCreate_IMDBSuccess verifies that the IMDB rating service is created correctly.
func (s *RatingPlatformServiceModelFactorySuite) TestCreate_IMDBSuccess() {
	// Arrange
	baseIMDBService := rating_service_model.RatingService{Name: constant.RatingServiceIMDB}
	s.mockBaseFactory.On("BuildIMDBComponents").Return(baseIMDBService, nil).Once()

	// Act
	ratingService, err := s.factory.Create(constant.RatingServiceIMDB)

	// Assert
	s.NoError(err)
	s.NotNil(ratingService, "RatingService should not be nil")
	s.Equal(constant.RatingServiceIMDB, ratingService.Name, "Service name should be IMDB")
	s.NotNil(ratingService.LogoService, "LogoService should be initialized for IMDB")
}

// TestCreate_TMDBBuildError verifies error handling when TMDB component building fails.
func (s *RatingPlatformServiceModelFactorySuite) TestCreate_TMDBBuildError() {
	// Arrange
	expectedErr := errors.New("tmdb build error")
	// Return an empty RatingService struct on error, as per typical Go error handling.
	s.mockBaseFactory.On("BuildTMDBComponents").Return(rating_service_model.RatingService{}, expectedErr).Once()

	// Act
	ratingService, err := s.factory.Create(constant.RatingServiceTMDB)

	// Assert
	s.Error(err)
	s.Equal(expectedErr, err)
	s.Equal(rating_service_model.RatingService{}, ratingService, "RatingService should be zero value on error")
}

// TestCreate_RottenTomatoesBuildError verifies error handling when Rotten Tomatoes component building fails.
func (s *RatingPlatformServiceModelFactorySuite) TestCreate_RottenTomatoesBuildError() {
	// Arrange
	expectedErr := errors.New("rt build error")
	s.mockBaseFactory.On("BuildRottenTomatoesComponents").Return(rating_service_model.RatingService{}, expectedErr).Once()

	// Act
	ratingService, err := s.factory.Create(constant.RatingServiceRottenTomatoes)

	// Assert
	s.Error(err)
	s.Equal(expectedErr, err)
	s.Equal(rating_service_model.RatingService{}, ratingService, "RatingService should be zero value on error")
}

// TestCreate_IMDBBuildError verifies error handling when IMDB component building fails.
func (s *RatingPlatformServiceModelFactorySuite) TestCreate_IMDBBuildError() {
	// Arrange
	expectedErr := errors.New("imdb build error")
	s.mockBaseFactory.On("BuildIMDBComponents").Return(rating_service_model.RatingService{}, expectedErr).Once()

	// Act
	ratingService, err := s.factory.Create(constant.RatingServiceIMDB)

	// Assert
	s.Error(err)
	s.Equal(expectedErr, err)
	s.Equal(rating_service_model.RatingService{}, ratingService, "RatingService should be zero value on error")
}

// TestCreate_UnsupportedService verifies error handling for unsupported service names.
func (s *RatingPlatformServiceModelFactorySuite) TestCreate_UnsupportedService() {
	// Arrange
	unsupportedServiceName := "unsupported"

	// Act
	ratingService, err := s.factory.Create(unsupportedServiceName)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), fmt.Sprintf("unsupported rating service: %s", unsupportedServiceName))
	s.Equal(rating_service_model.RatingService{}, ratingService, "RatingService should be zero value on error")
}
