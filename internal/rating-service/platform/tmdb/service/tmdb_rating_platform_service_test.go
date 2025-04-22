package tmdb_test

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
	tmdb "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/service"
)

type TMDBRatingPlatformServiceTestSuite struct {
	suite.Suite
	mockSearchService *rating_mocks.SearchService
	service           *tmdb.TMDBRatingPlatformService
	logger            *zap.Logger
}

func (s *TMDBRatingPlatformServiceTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.mockSearchService = rating_mocks.NewSearchService(s.T())
	s.service = tmdb.NewTMDBRatingPlatformService(s.logger, s.mockSearchService)
}

func (s *TMDBRatingPlatformServiceTestSuite) TearDownTest() {
	s.mockSearchService.AssertExpectations(s.T())
}

func TestTMDBRatingPlatformServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TMDBRatingPlatformServiceTestSuite))
}

func (s *TMDBRatingPlatformServiceTestSuite) TestGetRating_Success() {
	// Arrange
	ctx := context.Background()
	item := model.Item{ID: "test-id", Title: "Test Movie", Year: 2023, Type: "movie"}
	expectedRating := model.Rating{
		Name:   constant.RatingServiceTMDB,
		Rating: 8.5,
		Type:   model.RatingServiceTypeAudience,
	}
	searchResults := []model.SearchResult{
		{ID: 1, Title: "Test Movie", Vote: 8.5},
	}

	s.mockSearchService.On("GetResults", ctx, item).Return(searchResults, nil)

	// Act
	rating, err := s.service.GetRating(ctx, item)

	// Assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedRating, rating)
}

func (s *TMDBRatingPlatformServiceTestSuite) TestGetRating_SearchServiceError() {
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

func (s *TMDBRatingPlatformServiceTestSuite) TestGetRating_NoResults() {
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

func (s *TMDBRatingPlatformServiceTestSuite) TestGetRating_FirstResultNoVote() {
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

func (s *TMDBRatingPlatformServiceTestSuite) TestGetRating_MultipleResults_PicksFirst() {
	// Arrange
	ctx := context.Background()
	item := model.Item{ID: "test-id", Title: "Test Movie", Year: 2023, Type: "movie"}
	expectedRating := model.Rating{
		Name:   constant.RatingServiceTMDB,
		Rating: 7.0,
		Type:   model.RatingServiceTypeAudience,
	}
	searchResults := []model.SearchResult{
		{ID: 1, Title: "Test Movie", Vote: 7.0},
		{ID: 2, Title: "Another Movie", Vote: 9.0},
	}

	s.mockSearchService.On("GetResults", ctx, item).Return(searchResults, nil)

	// Act
	rating, err := s.service.GetRating(ctx, item)

	// Assert
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expectedRating, rating)
}

func (s *TMDBRatingPlatformServiceTestSuite) TestGetRating_FirstResultNegativeVote() {
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
