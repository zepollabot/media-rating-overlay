package item_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	"github.com/zepollabot/media-rating-overlay/internal/rating-service/item"
	mocks "github.com/zepollabot/media-rating-overlay/internal/rating-service/mocks" // Mocks for RatingPlatformService and LogoService
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

type ItemEligibilityServiceTestSuite struct {
	suite.Suite
	logger                 *zap.Logger
	mockPlatformService    *mocks.RatingPlatformService
	mockLogoService        *mocks.LogoService
	itemEligibilityService *item.ItemEligibilityService
}

func TestItemEligibilityServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ItemEligibilityServiceTestSuite))
}

func (s *ItemEligibilityServiceTestSuite) SetupTest() {
	s.logger = zaptest.NewLogger(s.T())
	s.mockPlatformService = mocks.NewRatingPlatformService(s.T())
	s.mockLogoService = mocks.NewLogoService(s.T())
	// Default to one rating service for basic eligible cases
	// Specific tests can override ratingPlatformServices
	ratingServices := []ratingModel.RatingService{
		{
			Name:            "TestPlatform",
			PlatformService: s.mockPlatformService,
			LogoService:     s.mockLogoService,
		},
	}
	s.itemEligibilityService = item.NewItemEligibilityService(ratingServices, s.logger)
}

func (s *ItemEligibilityServiceTestSuite) TearDownTest() {
	// Verify mock expectations
	s.mockPlatformService.AssertExpectations(s.T())
	s.mockLogoService.AssertExpectations(s.T())
}

func (s *ItemEligibilityServiceTestSuite) TestIsEligible_WhenItemIsNotEligible() {
	// Arrange
	testItem := &model.Item{
		ID:         "test-id-1",
		Title:      "Test Item Not Eligible",
		IsEligible: false, // Explicitly not eligible
		Ratings:    []model.Rating{},
	}

	// Act
	isEligible := s.itemEligibilityService.IsEligible(testItem)

	// Assert
	assert.False(s.T(), isEligible, "Expected item to be not eligible because IsEligible flag is false")
}

func (s *ItemEligibilityServiceTestSuite) TestIsEligible_WhenNoRatingsAndNoRatingServices() {
	// Arrange
	// Reconfigure service with no rating platform services
	s.itemEligibilityService = item.NewItemEligibilityService([]ratingModel.RatingService{}, s.logger)

	testItem := &model.Item{
		ID:         "test-id-2",
		Title:      "Test Item No Ratings No Services",
		IsEligible: true,
		Ratings:    []model.Rating{}, // No ratings
	}

	// Act
	isEligible := s.itemEligibilityService.IsEligible(testItem)

	// Assert
	assert.False(s.T(), isEligible, "Expected item to be not eligible due to no ratings and no rating services")
}

func (s *ItemEligibilityServiceTestSuite) TestIsEligible_WhenHasRatingsAndNoRatingServices() {
	// Arrange
	// Reconfigure service with no rating platform services
	s.itemEligibilityService = item.NewItemEligibilityService([]ratingModel.RatingService{}, s.logger)

	testItem := &model.Item{
		ID:         "test-id-3",
		Title:      "Test Item Has Ratings No Services",
		IsEligible: true,
		Ratings: []model.Rating{ // Has ratings
			{Name: "User", Rating: 5.0, Type: model.RatingServiceTypeUser},
		},
	}

	// Act
	isEligible := s.itemEligibilityService.IsEligible(testItem)

	// Assert
	// This should be true because it has ratings, even if no platform services are configured
	// for fetching *new* ratings. The eligibility check seems to care if *any* ratings exist OR if services exist.
	assert.True(s.T(), isEligible, "Expected item to be eligible because it has ratings")
}

func (s *ItemEligibilityServiceTestSuite) TestIsEligible_WhenNoRatingsButHasRatingServices() {
	// Arrange
	// Service is already configured with one rating service in SetupTest
	testItem := &model.Item{
		ID:         "test-id-4",
		Title:      "Test Item No Ratings Has Services",
		IsEligible: true,
		Ratings:    []model.Rating{}, // No ratings
	}

	// Act
	isEligible := s.itemEligibilityService.IsEligible(testItem)

	// Assert
	assert.True(s.T(), isEligible, "Expected item to be eligible because rating services are configured")
}

func (s *ItemEligibilityServiceTestSuite) TestIsEligible_WhenHasRatingsAndHasRatingServices() {
	// Arrange
	// Service is already configured with one rating service in SetupTest
	testItem := &model.Item{
		ID:         "test-id-5",
		Title:      "Test Item Has Ratings Has Services",
		IsEligible: true,
		Ratings: []model.Rating{ // Has ratings
			{Name: "User", Rating: 4.5, Type: model.RatingServiceTypeUser},
		},
	}

	// Act
	isEligible := s.itemEligibilityService.IsEligible(testItem)

	// Assert
	assert.True(s.T(), isEligible, "Expected item to be eligible because it has ratings and services are configured")
}

func (s *ItemEligibilityServiceTestSuite) TestIsEligible_WhenItemIsEligibleAndNoServicesButHasEmptyRatingsSlice() {
	// Arrange
	s.itemEligibilityService = item.NewItemEligibilityService([]ratingModel.RatingService{}, s.logger)
	testItem := &model.Item{
		ID:         "test-id-6",
		Title:      "Test Item Eligible, No Services, Empty Ratings Slice",
		IsEligible: true,
		Ratings:    make([]model.Rating, 0), // Empty but non-nil slice
	}

	// Act
	isEligible := s.itemEligibilityService.IsEligible(testItem)

	// Assert
	assert.False(s.T(), isEligible, "Expected item to be not eligible with empty ratings and no services")
}

func (s *ItemEligibilityServiceTestSuite) TestIsEligible_WhenItemIsEligibleAndNoRatingsButHasEmptyServicesSlice() {
	// Arrange
	s.itemEligibilityService = item.NewItemEligibilityService(make([]ratingModel.RatingService, 0), s.logger)
	testItem := &model.Item{
		ID:         "test-id-7",
		Title:      "Test Item Eligible, No Ratings, Empty Services Slice",
		IsEligible: true,
		Ratings:    nil, // No ratings
	}

	// Act
	isEligible := s.itemEligibilityService.IsEligible(testItem)

	// Assert
	assert.False(s.T(), isEligible, "Expected item to be not eligible with no ratings and empty services slice")
}
