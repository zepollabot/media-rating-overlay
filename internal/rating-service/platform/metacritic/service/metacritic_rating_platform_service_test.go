package metacritic_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/constant"
	"github.com/zepollabot/media-rating-overlay/internal/model"
	rating_mocks "github.com/zepollabot/media-rating-overlay/internal/rating-service/mocks"
	metacritic "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/metacritic/service"
)

type MetacriticRatingPlatformServiceTestSuite struct {
	suite.Suite
	mockSearchService *rating_mocks.SearchService
	service           *metacritic.MetacriticRatingPlatformService
	logger            *zap.Logger
}

func (s *MetacriticRatingPlatformServiceTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.mockSearchService = rating_mocks.NewSearchService(s.T())
	s.service = metacritic.NewMetacriticRatingPlatformService(s.logger, s.mockSearchService)
}

func (s *MetacriticRatingPlatformServiceTestSuite) TearDownTest() {
	s.mockSearchService.AssertExpectations(s.T())
}

func TestMetacriticRatingPlatformServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MetacriticRatingPlatformServiceTestSuite))
}

func (s *MetacriticRatingPlatformServiceTestSuite) TestGetRating_Success() {
	// Arrange
	ctx := context.Background()
	item := model.Item{ID: "test-id", Title: "Test Movie", Year: 2023, Type: "movie"}
	expectedRating := model.Rating{
		Name:   constant.RatingServiceMetacritic,
		Rating: 85,
		Type:   model.RatingServiceTypeCritic,
	}
	searchResults := []model.SearchResult{
		{ID: 1, Title: "Test Movie", Vote: 85},
	}

	s.mockSearchService.On("GetResults", ctx, item).Return(searchResults, nil)

	// Act
	rating, err := s.service.GetRating(ctx, item)

	// Assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedRating, rating)
}

func (s *MetacriticRatingPlatformServiceTestSuite) TestGetRating_SearchServiceError() {
	// Arrange
	ctx := context.Background()
	item := model.Item{ID: "test-id", Title: "Test Movie", Year: 2023, Type: "movie"}
	expectedError := errors.New("search service error")

	s.mockSearchService.On("GetResults", ctx, item).Return(nil, expectedError)

	// Act
	rating, err := s.service.GetRating(ctx, item)

	// Assert
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expectedError, err)
	assert.Empty(s.T(), rating)
}

func (s *MetacriticRatingPlatformServiceTestSuite) TestGetRating_NoResults() {
	// Arrange
	ctx := context.Background()
	item := model.Item{ID: "test-id", Title: "Test Movie", Year: 2023, Type: "movie"}

	s.mockSearchService.On("GetResults", ctx, item).Return([]model.SearchResult{}, nil)

	// Act
	rating, err := s.service.GetRating(ctx, item)

	// Assert
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), rating)
}

func (s *MetacriticRatingPlatformServiceTestSuite) TestGetRating_FirstResultNoVote() {
	// Arrange
	ctx := context.Background()
	item := model.Item{ID: "test-id", Title: "Test Movie", Year: 2023, Type: "movie"}
	searchResults := []model.SearchResult{
		{ID: 1, Title: "Test Movie", Vote: 0},
	}

	s.mockSearchService.On("GetResults", ctx, item).Return(searchResults, nil)

	// Act
	rating, err := s.service.GetRating(ctx, item)

	// Assert
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), rating)
}

func (s *MetacriticRatingPlatformServiceTestSuite) TestGetRating_MultipleResults_PicksFirst() {
	// Arrange
	ctx := context.Background()
	item := model.Item{ID: "test-id", Title: "Test Movie", Year: 2023, Type: "movie"}
	expectedRating := model.Rating{
		Name:   constant.RatingServiceMetacritic,
		Rating: 70,
		Type:   model.RatingServiceTypeCritic,
	}
	searchResults := []model.SearchResult{
		{ID: 1, Title: "Test Movie", Vote: 70},
		{ID: 2, Title: "Another Movie", Vote: 90},
	}

	s.mockSearchService.On("GetResults", ctx, item).Return(searchResults, nil)

	// Act
	rating, err := s.service.GetRating(ctx, item)

	// Assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedRating, rating)
}

func (s *MetacriticRatingPlatformServiceTestSuite) TestGetRating_FirstResultNegativeVote() {
	// Arrange
	ctx := context.Background()
	item := model.Item{ID: "test-id", Title: "Test Movie", Year: 2023, Type: "movie"}
	searchResults := []model.SearchResult{
		{ID: 1, Title: "Test Movie", Vote: -1},
	}

	s.mockSearchService.On("GetResults", ctx, item).Return(searchResults, nil)

	// Act
	rating, err := s.service.GetRating(ctx, item)

	// Assert
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), rating)
}
