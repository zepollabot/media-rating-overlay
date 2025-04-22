package rotten_tomatoes

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	mocks "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/rotten-tomatoes/logo/mocks"
)

type RottenTomatoesLogoServiceTestSuite struct {
	suite.Suite
	mockLogoCreator              *mocks.LogoCreator
	logger                       *zap.Logger
	config                       *model.PosterConfig
	service                      *RottenTomatoesLogoService
	testAudienceNormalConfigPath string
	testAudienceLowConfigPath    string
	testCriticNormalConfigPath   string
	testCriticLowConfigPath      string
}

func (s *RottenTomatoesLogoServiceTestSuite) SetupTest() {
	s.mockLogoCreator = new(mocks.LogoCreator)
	s.logger = zap.NewNop()

	s.testAudienceNormalConfigPath = "path/to/rt/audience/normal.png"
	s.testAudienceLowConfigPath = "path/to/rt/audience/low.png"
	s.testCriticNormalConfigPath = "path/to/rt/critic/normal.png"
	s.testCriticLowConfigPath = "path/to/rt/critic/low.png"

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
			IMDB struct {
				Audience struct{ Normal string }
			}
			TMDB struct {
				Audience struct{ Normal string }
			}
		}{
			RottenTomatoes: struct {
				Critic struct {
					Normal string
					Low    string
				}
				Audience struct {
					Normal string
					Low    string
				}
			}{
				Audience: struct {
					Normal string
					Low    string
				}{
					Normal: s.testAudienceNormalConfigPath,
					Low:    s.testAudienceLowConfigPath,
				},
				Critic: struct {
					Normal string
					Low    string
				}{
					Normal: s.testCriticNormalConfigPath,
					Low:    s.testCriticLowConfigPath,
				},
			},
		},
	}

	s.service = NewRottenTomatoesLogoService(s.logger, s.config, s.mockLogoCreator)
}

func (s *RottenTomatoesLogoServiceTestSuite) TearDownTest() {
	s.mockLogoCreator.AssertExpectations(s.T())
}

func TestRottenTomatoesLogoServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RottenTomatoesLogoServiceTestSuite))
}

func (s *RottenTomatoesLogoServiceTestSuite) TestNewRottenTomatoesLogoService() {
	logger := zap.NewNop()
	config := &model.PosterConfig{}
	mockLogoCreator := new(mocks.LogoCreator)

	service := NewRottenTomatoesLogoService(logger, config, mockLogoCreator)

	s.NotNil(service)
	s.Equal(logger, service.logger)
	s.Equal(config, service.config)
	s.Equal(mockLogoCreator, service.logoCreator)
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_Success_Audience_Normal() {
	ctx := context.Background()
	ratings := []model.Rating{{Type: model.RatingServiceTypeAudience, Rating: 7.5}} // 75%
	dimensions := model.LogoDimensions{}
	expectedLogo := &model.Logo{}
	expectedRatingText := "75%"

	s.mockLogoCreator.On("CreateLogo", s.testAudienceNormalConfigPath, expectedRatingText, dimensions).Return(expectedLogo, nil).Once()

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 1)
	s.Equal(expectedLogo, logos[0])
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_Success_Audience_Low() {
	ctx := context.Background()
	ratings := []model.Rating{{Type: model.RatingServiceTypeAudience, Rating: 5.9}} // 59%
	dimensions := model.LogoDimensions{}
	expectedLogo := &model.Logo{}
	expectedRatingText := "59%"

	s.mockLogoCreator.On("CreateLogo", s.testAudienceLowConfigPath, expectedRatingText, dimensions).Return(expectedLogo, nil).Once()

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 1)
	s.Equal(expectedLogo, logos[0])
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_Success_Critic_Normal() {
	ctx := context.Background()
	ratings := []model.Rating{{Type: model.RatingServiceTypeCritic, Rating: 6.0}} // 60%
	dimensions := model.LogoDimensions{}
	expectedLogo := &model.Logo{}
	expectedRatingText := "60%"

	s.mockLogoCreator.On("CreateLogo", s.testCriticNormalConfigPath, expectedRatingText, dimensions).Return(expectedLogo, nil).Once()

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 1)
	s.Equal(expectedLogo, logos[0])
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_Success_Critic_Low() {
	ctx := context.Background()
	ratings := []model.Rating{{Type: model.RatingServiceTypeCritic, Rating: 0.1}} // 1%
	dimensions := model.LogoDimensions{}
	expectedLogo := &model.Logo{}
	expectedRatingText := "1%"

	s.mockLogoCreator.On("CreateLogo", s.testCriticLowConfigPath, expectedRatingText, dimensions).Return(expectedLogo, nil).Once()

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 1)
	s.Equal(expectedLogo, logos[0])
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_Success_MultipleRatings() {
	ctx := context.Background()
	ratings := []model.Rating{
		{Type: model.RatingServiceTypeAudience, Rating: 8.0}, // 80% Audience Normal
		{Type: model.RatingServiceTypeCritic, Rating: 4.5},   // 45% Critic Low
	}
	dimensions := model.LogoDimensions{}
	expectedLogo1 := &model.Logo{}
	expectedRatingText1 := "80%"
	s.mockLogoCreator.On("CreateLogo", s.testAudienceNormalConfigPath, expectedRatingText1, dimensions).Return(expectedLogo1, nil).Once()

	expectedLogo2 := &model.Logo{}
	expectedRatingText2 := "45%"
	s.mockLogoCreator.On("CreateLogo", s.testCriticLowConfigPath, expectedRatingText2, dimensions).Return(expectedLogo2, nil).Once()

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 2)
	s.Contains(logos, expectedLogo1)
	s.Contains(logos, expectedLogo2)
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_NoRatings() {
	ctx := context.Background()
	var ratings []model.Rating
	dimensions := model.LogoDimensions{}

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 0)
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_RatingIsZero() {
	ctx := context.Background()
	ratings := []model.Rating{{Type: model.RatingServiceTypeAudience, Rating: 0.0}}
	dimensions := model.LogoDimensions{}

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 0)
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_UnknownRatingType() {
	ctx := context.Background()
	ratings := []model.Rating{{Type: "UnknownType", Rating: 7.0}}
	dimensions := model.LogoDimensions{}

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 0) // Expect no logo for unknown type
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_CreateLogoError() {
	ctx := context.Background()
	ratings := []model.Rating{{Type: model.RatingServiceTypeAudience, Rating: 9.0}} // 90%
	dimensions := model.LogoDimensions{}
	expectedError := errors.New("logo creation failed")
	expectedRatingText := "90%"

	s.mockLogoCreator.On("CreateLogo", s.testAudienceNormalConfigPath, expectedRatingText, dimensions).Return(nil, expectedError).Once()

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.Error(err)
	s.Equal(expectedError, err)
	s.Nil(logos) // Important: On error, the service returns nil for the logos slice
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_RatingCalculation_Rounding() {
	ctx := context.Background()
	// Test case for rounding: 5.95 * 10 = 59.5, rounds to "60%"
	ratings := []model.Rating{{Type: model.RatingServiceTypeAudience, Rating: 5.95}}
	dimensions := model.LogoDimensions{}
	expectedLogo := &model.Logo{}
	expectedRatingText := "60%" // 5.95 * 10 = 59.5, decimal.Round(0) makes it 60

	// This should use low path because 60% is the text but the image choice is made based on the percentage
	s.mockLogoCreator.On("CreateLogo", s.testAudienceLowConfigPath, expectedRatingText, dimensions).Return(expectedLogo, nil).Once()

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 1)
	s.Equal(expectedLogo, logos[0])
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_RatingCalculation_ExactZero() {
	ctx := context.Background()
	ratings := []model.Rating{{Type: model.RatingServiceTypeCritic, Rating: 0.001}} // 0.01 * 10 = 0.01 => "0%"
	dimensions := model.LogoDimensions{}
	expectedLogo := &model.Logo{}
	expectedRatingText := "0%" // 0.001 * 10 = 0.01, decimal.StringFixedBank(0) makes it "0"

	s.mockLogoCreator.On("CreateLogo", s.testCriticLowConfigPath, expectedRatingText, dimensions).Return(expectedLogo, nil).Once()

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.NoError(err)
	s.NotNil(logos)
	s.Len(logos, 1)
	s.Equal(expectedLogo, logos[0])
}

func (s *RottenTomatoesLogoServiceTestSuite) TestGetLogos_ContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	ratings := []model.Rating{{Type: model.RatingServiceTypeAudience, Rating: 8.0}}
	dimensions := model.LogoDimensions{}
	expectedRatingText := "80%"

	cancel()

	s.mockLogoCreator.On("CreateLogo", s.testAudienceNormalConfigPath, expectedRatingText, dimensions).Return(nil, context.Canceled).Once()

	logos, err := s.service.GetLogos(ctx, ratings, "tt123", dimensions)

	s.ErrorIs(err, context.Canceled)
	s.Nil(logos)
}
