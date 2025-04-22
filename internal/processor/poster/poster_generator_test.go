package poster

import (
	"context"
	"testing"

	"github.com/fogleman/gg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
	poster_mocks "github.com/zepollabot/media-rating-overlay/internal/processor/poster/mocks"
	rating_service_mocks "github.com/zepollabot/media-rating-overlay/internal/rating-service/mocks"
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

type PosterGeneratorTestSuite struct {
	suite.Suite
	generator      *PosterGenerator
	logger         *zap.Logger
	imageProcessor *poster_mocks.ImageProcessor
	logoService    *poster_mocks.LogoService
	overlayService *poster_mocks.OverlayService
	posterConfig   *model.PosterConfig
	libraryConfig  *config.Library
	ratingServices []ratingModel.RatingService
}

func (s *PosterGeneratorTestSuite) SetupSuite() {
	s.logger = zap.NewNop()
	s.imageProcessor = poster_mocks.NewImageProcessor(s.T())
	s.logoService = poster_mocks.NewLogoService(s.T())
	s.overlayService = poster_mocks.NewOverlayService(s.T())
	s.posterConfig = model.PosterConfigWithDefaultValues()
	s.libraryConfig = &config.Library{
		Overlay: config.Overlay{
			Height: 0.2,
		},
	}
	s.ratingServices = []ratingModel.RatingService{}

	s.generator = NewPosterGenerator(
		s.logger,
		s.imageProcessor,
		s.logoService,
		s.overlayService,
		s.posterConfig,
		s.ratingServices,
		false,
	)
}

func (s *PosterGeneratorTestSuite) TearDownTest() {
	s.imageProcessor.ExpectedCalls = nil
	s.logoService.ExpectedCalls = nil
	s.overlayService.ExpectedCalls = nil
}

func (s *PosterGeneratorTestSuite) TestApplyLogos() {
	// Arrange
	filePath := "test.jpg"
	item := model.Item{
		ID: "test-id",
		Ratings: []model.Rating{
			{Name: "service1", Rating: 8.5, Type: model.RatingServiceTypeCritic},
			{Name: "service2", Rating: 9.0, Type: model.RatingServiceTypeAudience},
		},
	}

	drawContext := gg.NewContext(100, 100)
	logoContext := gg.NewContext(100, 100)

	s.overlayService.EXPECT().CreateDrawContextWithOverlay(filePath, item, s.libraryConfig).Return(drawContext, nil)
	s.logoService.EXPECT().PositionLogos(mock.Anything, mock.Anything, mock.Anything, false).Return(logoContext, nil)
	s.imageProcessor.EXPECT().SaveImage(mock.Anything, filePath).Return(filePath, nil)

	// Act
	resultPath, err := s.generator.ApplyLogos(context.Background(), filePath, s.libraryConfig, item)

	// Assert
	s.Require().NoError(err)
	s.Equal(filePath, resultPath)
	s.overlayService.AssertExpectations(s.T())
	s.logoService.AssertExpectations(s.T())
	s.imageProcessor.AssertExpectations(s.T())
}

func (s *PosterGeneratorTestSuite) TestApplyLogos_OverlayError() {
	// Arrange
	filePath := "test.jpg"
	item := model.Item{
		ID: "test-id",
		Ratings: []model.Rating{
			{Name: "service1", Rating: 8.5, Type: model.RatingServiceTypeCritic},
		},
	}

	s.overlayService.EXPECT().CreateDrawContextWithOverlay(filePath, item, s.libraryConfig).Return(nil, assert.AnError)

	// Act
	resultPath, err := s.generator.ApplyLogos(context.Background(), filePath, s.libraryConfig, item)

	// Assert
	s.Error(err)
	s.Equal(filePath, resultPath)
	s.overlayService.AssertExpectations(s.T())
}

func (s *PosterGeneratorTestSuite) TestApplyLogos_LogoPositioningError() {
	// Arrange
	filePath := "test.jpg"
	item := model.Item{
		ID: "test-id",
		Ratings: []model.Rating{
			{Name: "service1", Rating: 8.5, Type: model.RatingServiceTypeCritic},
		},
	}

	drawContext := gg.NewContext(100, 100)

	s.overlayService.EXPECT().CreateDrawContextWithOverlay(filePath, item, s.libraryConfig).Return(drawContext, nil)
	s.logoService.EXPECT().PositionLogos(mock.Anything, mock.Anything, mock.Anything, false).Return(nil, assert.AnError)

	// Act
	resultPath, err := s.generator.ApplyLogos(context.Background(), filePath, s.libraryConfig, item)

	// Assert
	s.Error(err)
	s.Equal(filePath, resultPath)
	s.overlayService.AssertExpectations(s.T())
	s.logoService.AssertExpectations(s.T())
}

func (s *PosterGeneratorTestSuite) TestApplyLogos_SaveImageError() {
	// Arrange
	filePath := "test.jpg"
	item := model.Item{
		ID: "test-id",
		Ratings: []model.Rating{
			{Name: "service1", Rating: 8.5, Type: model.RatingServiceTypeCritic},
		},
	}

	drawContext := gg.NewContext(100, 100)
	logoContext := gg.NewContext(100, 100)

	s.overlayService.EXPECT().CreateDrawContextWithOverlay(filePath, item, s.libraryConfig).Return(drawContext, nil)
	s.logoService.EXPECT().PositionLogos(mock.Anything, mock.Anything, mock.Anything, false).Return(logoContext, nil)
	s.imageProcessor.EXPECT().SaveImage(mock.Anything, filePath).Return("", assert.AnError)

	// Act
	resultPath, err := s.generator.ApplyLogos(context.Background(), filePath, s.libraryConfig, item)

	// Assert
	s.Error(err)
	s.Equal(filePath, resultPath) // Expect original filepath on error
	s.overlayService.AssertExpectations(s.T())
	s.logoService.AssertExpectations(s.T())
	s.imageProcessor.AssertExpectations(s.T())
}

func (s *PosterGeneratorTestSuite) TestBuildLogos_RatingServiceFound() {
	// Arrange
	ctx := context.Background()
	item := model.Item{
		ID: "test-id",
		Ratings: []model.Rating{
			{Name: "mockService", Rating: 7.5, Type: model.RatingServiceTypeUser},
		},
	}
	logoDimensions := model.LogoDimensions{AreaWidth: 50, AreaHeight: 20}

	mockRatingLogoService := rating_service_mocks.NewLogoService(s.T())
	mockRatingLogoService.EXPECT().GetLogos(ctx, []model.Rating{item.Ratings[0]}, "mockService", logoDimensions).Return([]*model.Logo{{Text: model.Text{Value: "logo_text_1"}}}, nil)

	// Create a concrete RatingService struct and assign the mock LogoService to it
	ratingServiceInstance := ratingModel.RatingService{
		Name:        "mockService",
		LogoService: mockRatingLogoService,
	}

	s.generator.ratingPlatformServices = []ratingModel.RatingService{ratingServiceInstance}

	// Act
	logos := s.generator.buildLogos(ctx, item, logoDimensions)

	// Assert
	s.Len(logos, 1)
	s.Equal("logo_text_1", logos[0].Text.Value)
	// No AssertExpectations on ratingServiceInstance as it's not a mock
	// mockRatingLogoService.AssertExpectations(s.T()) // Handled by t.Cleanup in NewLogoService
}

func (s *PosterGeneratorTestSuite) TestBuildLogos_GetLogosError() {
	// Arrange
	ctx := context.Background()
	item := model.Item{
		ID: "test-id",
		Ratings: []model.Rating{
			{Name: "mockServiceError", Rating: 7.5, Type: model.RatingServiceTypeUser},
			{Name: "mockServiceOK", Rating: 8.0, Type: model.RatingServiceTypeCritic},
		},
	}
	logoDimensions := model.LogoDimensions{AreaWidth: 50, AreaHeight: 20}

	// Service 1 (returns error)
	mockRatingLogoService1 := rating_service_mocks.NewLogoService(s.T())
	mockRatingLogoService1.EXPECT().GetLogos(ctx, []model.Rating{item.Ratings[0]}, "mockServiceError", logoDimensions).Return(nil, assert.AnError)
	ratingServiceInstance1 := ratingModel.RatingService{
		Name:        "mockServiceError",
		LogoService: mockRatingLogoService1,
	}

	// Service 2 (returns success)
	mockRatingLogoService2 := rating_service_mocks.NewLogoService(s.T())
	mockRatingLogoService2.EXPECT().GetLogos(ctx, []model.Rating{item.Ratings[1]}, "mockServiceOK", logoDimensions).Return([]*model.Logo{{Text: model.Text{Value: "logo_text_2"}}}, nil)
	ratingServiceInstance2 := ratingModel.RatingService{
		Name:        "mockServiceOK",
		LogoService: mockRatingLogoService2,
	}

	s.generator.ratingPlatformServices = []ratingModel.RatingService{ratingServiceInstance1, ratingServiceInstance2}

	// Act
	logos := s.generator.buildLogos(ctx, item, logoDimensions)

	// Assert
	s.Len(logos, 1) // Expecting one logo as the first call to GetLogos fails
	s.Equal("logo_text_2", logos[0].Text.Value)
	// mockRatingLogoService1.AssertExpectations(s.T()) // Handled by t.Cleanup
	// mockRatingLogoService2.AssertExpectations(s.T()) // Handled by t.Cleanup
}

func TestPosterGeneratorSuite(t *testing.T) {
	suite.Run(t, new(PosterGeneratorTestSuite))
}
