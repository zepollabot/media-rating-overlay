package imdb

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	mocks "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/imdb/logo/mocks"
)

type IMDBLogoServiceTestSuite struct {
	suite.Suite
	mockLogoCreator *mocks.LogoCreator
	logger          *zap.Logger
	config          *model.PosterConfig
	testConfigPath  string
	service         *IMDBLogoService
}

func (s *IMDBLogoServiceTestSuite) SetupTest() {
	s.mockLogoCreator = new(mocks.LogoCreator)
	s.logger = zap.NewNop() // Use Noop logger for tests

	s.testConfigPath = "path/to/imdb/audience/normal.png"

	s.config = &model.PosterConfig{
		ImagePaths: struct { // Anonymous struct for ImagePaths
			RottenTomatoes struct {
				Critic struct {
					Normal string
					Low    string
				}
				Audience struct {
					Normal string
					Low    string
				}
			}
			IMDB struct { // Anonymous struct for IMDB
				Audience struct { // Anonymous struct for Audience
					Normal string
				}
			}
			TMDB struct {
				Audience struct {
					Normal string
				}
			}
		}{
			IMDB: struct { // Initialize IMDB
				Audience struct {
					Normal string
				}
			}{
				Audience: struct { // Initialize Audience
					Normal string
				}{
					Normal: s.testConfigPath,
				},
			},
			// Other fields like RottenTomatoes and TMDB can be zero-valued
			// as they are not used by the service under test.
		},
	}

	s.service = NewIMDBLogoService(s.logger, s.config, s.mockLogoCreator)
}

func (s *IMDBLogoServiceTestSuite) TearDownTest() {
	s.mockLogoCreator.AssertExpectations(s.T())
}

func TestIMDBLogoServiceTestSuite(t *testing.T) {
	suite.Run(t, new(IMDBLogoServiceTestSuite))
}

func (s *IMDBLogoServiceTestSuite) TestNewIMDBLogoService() {
	// Arrange
	logger := zap.NewNop()
	config := &model.PosterConfig{}
	mockLogoCreator := new(mocks.LogoCreator)

	// Act
	service := NewIMDBLogoService(logger, config, mockLogoCreator)

	// Assert
	s.NotNil(service)
	s.Equal(logger, service.logger)
	s.Equal(config, service.config)
	s.Equal(mockLogoCreator, service.logoCreator)
}

func (s *IMDBLogoServiceTestSuite) TestGetLogos_Success_SingleRating() {
	// Arrange
	ctx := context.Background()
	ratings := []model.Rating{
		{Rating: 7.8},
	}
	itemID := "tt1234567"
	dimensions := model.LogoDimensions{}
	expectedRatingStr := decimal.NewFromFloat32(7.8).Round(1).StringFixedBank(1)
	expectedLogo := &model.Logo{}

	s.mockLogoCreator.On("CreateLogo", s.testConfigPath, expectedRatingStr, dimensions).Return(expectedLogo, nil).Once()

	// Act
	logos, err := s.service.GetLogos(ctx, ratings, itemID, dimensions)

	// Assert
	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 1)
	s.Equal(expectedLogo, logos[0])
}

func (s *IMDBLogoServiceTestSuite) TestGetLogos_Success_MultipleRatings() {
	// Arrange
	ctx := context.Background()
	ratings := []model.Rating{
		{Rating: 7.8},
		{Rating: 8.2},
	}
	itemID := "tt1234567"
	dimensions := model.LogoDimensions{}

	expectedRatingStr1 := decimal.NewFromFloat32(7.8).Round(1).StringFixedBank(1)
	expectedLogo1 := &model.Logo{}
	s.mockLogoCreator.On("CreateLogo", s.testConfigPath, expectedRatingStr1, dimensions).Return(expectedLogo1, nil).Once()

	expectedRatingStr2 := decimal.NewFromFloat32(8.2).Round(1).StringFixedBank(1)
	expectedLogo2 := &model.Logo{}
	s.mockLogoCreator.On("CreateLogo", s.testConfigPath, expectedRatingStr2, dimensions).Return(expectedLogo2, nil).Once()

	// Act
	logos, err := s.service.GetLogos(ctx, ratings, itemID, dimensions)

	// Assert
	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 2)
	s.Contains(logos, expectedLogo1)
	s.Contains(logos, expectedLogo2)
}

func (s *IMDBLogoServiceTestSuite) TestGetLogos_NoRatings() {
	// Arrange
	ctx := context.Background()
	var ratings []model.Rating
	itemID := "tt1234567"
	dimensions := model.LogoDimensions{}

	// Act
	logos, err := s.service.GetLogos(ctx, ratings, itemID, dimensions)

	// Assert
	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 0)
}

func (s *IMDBLogoServiceTestSuite) TestGetLogos_RatingIsZero() {
	// Arrange
	ctx := context.Background()
	ratings := []model.Rating{
		{Rating: 0.0},
	}
	itemID := "tt1234567"
	dimensions := model.LogoDimensions{}

	// Act
	logos, err := s.service.GetLogos(ctx, ratings, itemID, dimensions)

	// Assert
	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 0)
}

func (s *IMDBLogoServiceTestSuite) TestGetLogos_CreateLogoError() {
	// Arrange
	ctx := context.Background()
	ratings := []model.Rating{
		{Rating: 7.8},
	}
	itemID := "tt1234567"
	dimensions := model.LogoDimensions{}
	expectedRatingStr := decimal.NewFromFloat32(7.8).Round(1).StringFixedBank(1)
	expectedError := errors.New("failed to create logo")

	s.mockLogoCreator.On("CreateLogo", s.testConfigPath, expectedRatingStr, dimensions).Return(nil, expectedError).Once()

	// Act
	logos, err := s.service.GetLogos(ctx, ratings, itemID, dimensions)

	// Assert
	s.Error(err)
	s.Equal(expectedError, err)
	s.Nil(logos)
}

func (s *IMDBLogoServiceTestSuite) TestGetLogos_Success_MultipleRatings_OneZero() {
	// Arrange
	ctx := context.Background()
	ratings := []model.Rating{
		{Rating: 7.8},
		{Rating: 0.0},
		{Rating: 8.2},
	}
	itemID := "tt1234567"
	dimensions := model.LogoDimensions{}

	expectedRatingStr1 := decimal.NewFromFloat32(7.8).Round(1).StringFixedBank(1)
	expectedLogo1 := &model.Logo{}
	s.mockLogoCreator.On("CreateLogo", s.testConfigPath, expectedRatingStr1, dimensions).Return(expectedLogo1, nil).Once()

	expectedRatingStr2 := decimal.NewFromFloat32(8.2).Round(1).StringFixedBank(1)
	expectedLogo2 := &model.Logo{}
	s.mockLogoCreator.On("CreateLogo", s.testConfigPath, expectedRatingStr2, dimensions).Return(expectedLogo2, nil).Once()

	// Act
	logos, err := s.service.GetLogos(ctx, ratings, itemID, dimensions)

	// Assert
	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 2)
	s.Contains(logos, expectedLogo1)
	s.Contains(logos, expectedLogo2)
}

func (s *IMDBLogoServiceTestSuite) TestGetLogos_Success_RatingRounded() {
	// Arrange
	ctx := context.Background()
	ratings := []model.Rating{
		{Rating: 7.85}, // Will be rounded to 7.9
		{Rating: 7.84}, // Will be rounded to 7.8
	}
	itemID := "tt1234567"
	dimensions := model.LogoDimensions{}

	expectedRatingStr1 := "7.9"
	expectedLogo1 := &model.Logo{}
	s.mockLogoCreator.On("CreateLogo", s.testConfigPath, expectedRatingStr1, dimensions).Return(expectedLogo1, nil).Once()

	expectedRatingStr2 := "7.8"
	expectedLogo2 := &model.Logo{}
	s.mockLogoCreator.On("CreateLogo", s.testConfigPath, expectedRatingStr2, dimensions).Return(expectedLogo2, nil).Once()

	// Act
	logos, err := s.service.GetLogos(ctx, ratings, itemID, dimensions)

	// Assert
	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 2)
	s.Contains(logos, expectedLogo1)
	s.Contains(logos, expectedLogo2)
}

func (s *IMDBLogoServiceTestSuite) TestGetLogos_ContextCancelled() {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	ratings := []model.Rating{
		{Rating: 7.8},
	}
	itemID := "tt1234567"
	dimensions := model.LogoDimensions{}
	expectedRatingStr := decimal.NewFromFloat32(7.8).Round(1).StringFixedBank(1)

	cancel()

	s.mockLogoCreator.On("CreateLogo", s.testConfigPath, expectedRatingStr, dimensions).Return(nil, context.Canceled).Once()

	// Act
	logos, err := s.service.GetLogos(ctx, ratings, itemID, dimensions)

	// Assert
	s.ErrorIs(err, context.Canceled)
	s.Nil(logos)
}
