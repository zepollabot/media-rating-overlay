package rating

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	ratingmocks "github.com/zepollabot/media-rating-overlay/internal/rating-service/mocks"
	ratingmodel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

type RatingBuilderServiceTestSuite struct {
	suite.Suite
	service              *RatingBuilderService
	logger               *zap.Logger
	mockPlatformService1 *ratingmocks.RatingPlatformService
	mockPlatformService2 *ratingmocks.RatingPlatformService
}

func (s *RatingBuilderServiceTestSuite) SetupSuite() {
	s.logger = zap.NewNop()
	s.mockPlatformService1 = ratingmocks.NewRatingPlatformService(s.T())
	s.mockPlatformService2 = ratingmocks.NewRatingPlatformService(s.T())

	ratingServices := []ratingmodel.RatingService{
		{
			Name:            "service1",
			PlatformService: s.mockPlatformService1,
		},
		{
			Name:            "service2",
			PlatformService: s.mockPlatformService2,
		},
	}

	s.service = NewRatingBuilderService(ratingServices, s.logger)
}

func (s *RatingBuilderServiceTestSuite) TearDownTest() {
	s.mockPlatformService1.AssertExpectations(s.T())
	s.mockPlatformService2.AssertExpectations(s.T())
}

func (s *RatingBuilderServiceTestSuite) TestBuildRatings_NewRatings() {
	// Arrange
	item := &model.Item{
		ID:      "test-id",
		Ratings: []model.Rating{},
	}
	expectedRating1 := model.Rating{Name: "service1", Rating: 8.5, Type: model.RatingServiceTypeCritic}
	expectedRating2 := model.Rating{Name: "service2", Rating: 9.0, Type: model.RatingServiceTypeAudience}

	s.mockPlatformService1.EXPECT().
		GetRating(mock.Anything, *item).
		Return(expectedRating1, nil).
		Once()

	updatedItem := &model.Item{
		ID:      "test-id",
		Ratings: []model.Rating{expectedRating1},
	}

	s.mockPlatformService2.EXPECT().
		GetRating(mock.Anything, *updatedItem).
		Return(expectedRating2, nil).
		Once()

	// Act
	err := s.service.BuildRatings(context.Background(), item)

	// Assert
	s.Require().NoError(err)
	s.Len(item.Ratings, 2)
	s.Contains(item.Ratings, expectedRating1)
	s.Contains(item.Ratings, expectedRating2)
}

func (s *RatingBuilderServiceTestSuite) TestBuildRatings_ExistingRating() {
	// Arrange
	existingRating := model.Rating{Name: "service1", Rating: 8.5, Type: model.RatingServiceTypeCritic}
	item := &model.Item{
		ID:      "test-id",
		Ratings: []model.Rating{existingRating},
	}
	expectedRating2 := model.Rating{Name: "service2", Rating: 9.0, Type: model.RatingServiceTypeAudience}

	s.mockPlatformService2.EXPECT().
		GetRating(mock.Anything, *item).
		Return(expectedRating2, nil).
		Once()

	// Act
	err := s.service.BuildRatings(context.Background(), item)

	// Assert
	s.Require().NoError(err)
	s.Len(item.Ratings, 2)
	s.Contains(item.Ratings, existingRating)
	s.Contains(item.Ratings, expectedRating2)
}

func (s *RatingBuilderServiceTestSuite) TestBuildRatings_ServiceError() {
	// Arrange
	item := &model.Item{
		ID:      "test-id",
		Ratings: []model.Rating{},
	}

	s.mockPlatformService1.EXPECT().
		GetRating(mock.Anything, *item).
		Return(model.Rating{}, assert.AnError).
		Once()

	// Act
	err := s.service.BuildRatings(context.Background(), item)

	// Assert
	s.Error(err)
	s.Empty(item.Ratings)
}

func TestRatingBuilderServiceSuite(t *testing.T) {
	suite.Run(t, new(RatingBuilderServiceTestSuite))
}
