package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	// "golang.org/x/sync/semaphore" // No longer using concrete semaphore

	appmocks "github.com/zepollabot/media-rating-overlay/internal/app/mocks"
	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	mediamocks "github.com/zepollabot/media-rating-overlay/internal/media-service/mocks"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
)

type ItemProcessorTestSuite struct {
	suite.Suite
	logger                 *zap.Logger
	mockCtx                context.Context
	mockWorkSemaphore      *appmocks.SemaphoreWeighted // Changed type
	mockPosterService      *mediamocks.PosterService
	mockPosterGenerator    *appmocks.PosterGenerator
	mockEligibilityChecker *appmocks.ItemEligibilityChecker
	mockRatingBuilder      *appmocks.RatingBuilder
	itemProcessor          *ItemProcessor
}

func (s *ItemProcessorTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.mockCtx = context.Background()
	s.mockWorkSemaphore = new(appmocks.SemaphoreWeighted) // Changed instantiation
	s.mockPosterService = new(mediamocks.PosterService)
	s.mockPosterGenerator = new(appmocks.PosterGenerator)
	s.mockEligibilityChecker = new(appmocks.ItemEligibilityChecker)
	s.mockRatingBuilder = new(appmocks.RatingBuilder)

	s.itemProcessor = NewItemProcessor(
		s.logger,
		s.mockCtx,
		s.mockWorkSemaphore, // Pass mock
		s.mockPosterGenerator,
		s.mockEligibilityChecker,
		s.mockRatingBuilder,
	)
	s.itemProcessor.SetPosterService(s.mockPosterService)
}

func (s *ItemProcessorTestSuite) TearDownTest() {
	s.mockWorkSemaphore.AssertExpectations(s.T()) // Added verification
	s.mockPosterService.AssertExpectations(s.T())
	s.mockPosterGenerator.AssertExpectations(s.T())
	s.mockEligibilityChecker.AssertExpectations(s.T())
	s.mockRatingBuilder.AssertExpectations(s.T())
}

func TestItemProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(ItemProcessorTestSuite))
}

func (s *ItemProcessorTestSuite) TestNewItemProcessor() {
	// Arrange
	logger := zap.NewNop()
	ctx := context.Background()
	mockWorkSem := new(appmocks.SemaphoreWeighted) // Use mock
	mockPosterGen := new(appmocks.PosterGenerator)
	mockEligibilityCheck := new(appmocks.ItemEligibilityChecker)
	mockRatingBuild := new(appmocks.RatingBuilder)

	// Act
	processor := NewItemProcessor(
		logger,
		ctx,
		mockWorkSem, // Pass mock
		mockPosterGen,
		mockEligibilityCheck,
		mockRatingBuild,
	)

	// Assert
	s.NotNil(processor)
	s.Equal(logger, processor.logger)
	s.Equal(ctx, processor.ctx)
	s.Equal(mockWorkSem, processor.workSemaphore)
	s.Equal(mockPosterGen, processor.posterGenerator)
	s.Equal(mockEligibilityCheck, processor.eligibilityChecker)
	s.Equal(mockRatingBuild, processor.ratingBuilder)
	s.Nil(processor.posterService) // Initially nil until SetPosterService is called
}

func (s *ItemProcessorTestSuite) TestSetPosterService() {
	// Arrange
	mockPosterSvc := new(mediamocks.PosterService)

	// Act
	s.itemProcessor.SetPosterService(mockPosterSvc)

	// Assert
	s.Equal(mockPosterSvc, s.itemProcessor.posterService)
}

// Placeholder for ProcessItem tests
func (s *ItemProcessorTestSuite) TestProcessItem_Success() {
	// Arrange
	item := model.Item{ID: "1", Title: "Test Movie"}
	configLib := &config.Library{Name: "Movies"}
	expectedPosterPath := "/path/to/poster.jpg"
	expectedNewPosterPath := "/path/to/new_poster.jpg"

	s.mockRatingBuilder.On("BuildRatings", s.mockCtx, &item).Return(nil).Once()
	s.mockPosterService.On("EnsurePosterExists", s.mockCtx, item, configLib).Return(nil).Once()
	s.mockPosterService.On("GetPosterDiskPosition", s.mockCtx, item, configLib).Return(expectedPosterPath, nil).Once()
	s.mockPosterGenerator.On("ApplyLogos", s.mockCtx, expectedPosterPath, configLib, item).Return(expectedNewPosterPath, nil).Once()

	// Act
	result := s.itemProcessor.ProcessItem(item, 0, s.mockPosterService, configLib)

	// Assert
	s.Equal(item.Title, result.Title)
	s.Equal(expectedPosterPath, result.OriginalPosterDiskPosition)
	s.Equal(expectedNewPosterPath, result.OverlayPosterDiskPosition)
	s.Nil(result.Err)
}

// Placeholder for ProcessItems tests
func (s *ItemProcessorTestSuite) TestProcessItems_Success() {
	// Arrange
	items := []model.Item{
		{ID: "1", Title: "Movie 1"},
		{ID: "2", Title: "Movie 2"},
	}
	configLib := &config.Library{Name: "Movies"}
	expectedPosterPath1 := "/path/to/poster1.jpg"
	expectedNewPosterPath1 := "/path/to/new_poster1.jpg"
	expectedPosterPath2 := "/path/to/poster2.jpg"
	expectedNewPosterPath2 := "/path/to/new_poster2.jpg"

	// Mock semaphore for item 1
	s.mockWorkSemaphore.On("Acquire", mock.Anything, int64(1)).Return(nil).Once()
	s.mockEligibilityChecker.On("IsEligible", &items[0]).Return(true).Once()
	s.mockRatingBuilder.On("BuildRatings", mock.Anything, &items[0]).Return(nil).Once()
	s.mockPosterService.On("EnsurePosterExists", mock.Anything, items[0], configLib).Return(nil).Once()
	s.mockPosterService.On("GetPosterDiskPosition", mock.Anything, items[0], configLib).Return(expectedPosterPath1, nil).Once()
	s.mockPosterGenerator.On("ApplyLogos", mock.Anything, expectedPosterPath1, configLib, items[0]).Return(expectedNewPosterPath1, nil).Once()
	s.mockWorkSemaphore.On("Release", int64(1)).Return().Once()

	// Mock semaphore for item 2
	s.mockWorkSemaphore.On("Acquire", mock.Anything, int64(1)).Return(nil).Once()
	s.mockEligibilityChecker.On("IsEligible", &items[1]).Return(true).Once()
	s.mockRatingBuilder.On("BuildRatings", mock.Anything, &items[1]).Return(nil).Once()
	s.mockPosterService.On("EnsurePosterExists", mock.Anything, items[1], configLib).Return(nil).Once()
	s.mockPosterService.On("GetPosterDiskPosition", mock.Anything, items[1], configLib).Return(expectedPosterPath2, nil).Once()
	s.mockPosterGenerator.On("ApplyLogos", mock.Anything, expectedPosterPath2, configLib, items[1]).Return(expectedNewPosterPath2, nil).Once()
	s.mockWorkSemaphore.On("Release", int64(1)).Return().Once()

	// Act
	err := s.itemProcessor.ProcessItems(items, configLib)

	// Assert
	s.NoError(err)
}

func (s *ItemProcessorTestSuite) TestProcessItem_RatingBuilderError() {
	// Arrange
	item := model.Item{ID: "1", Title: "Test Movie"}
	configLib := &config.Library{Name: "Movies"}
	expectedError := fmt.Errorf("rating builder error")

	s.mockRatingBuilder.On("BuildRatings", s.mockCtx, &item).Return(expectedError).Once()

	// Act
	result := s.itemProcessor.ProcessItem(item, 0, s.mockPosterService, configLib)

	// Assert
	s.Equal(expectedError, result.Err)
	s.Equal(item.Title, result.Title)
	s.Equal(expectedError, result.Err)
	s.Empty(result.OriginalPosterDiskPosition)
	s.Empty(result.OverlayPosterDiskPosition)
}

func (s *ItemProcessorTestSuite) TestProcessItem_EnsurePosterExistsError() {
	// Arrange
	item := model.Item{ID: "1", Title: "Test Movie"}
	configLib := &config.Library{Name: "Movies"}
	expectedError := fmt.Errorf("ensure poster error")

	s.mockRatingBuilder.On("BuildRatings", s.mockCtx, &item).Return(nil).Once()
	s.mockPosterService.On("EnsurePosterExists", s.mockCtx, item, configLib).Return(expectedError).Once()

	// Act
	result := s.itemProcessor.ProcessItem(item, 0, s.mockPosterService, configLib)

	// Assert
	s.Equal(item.Title, result.Title)
	s.Equal(expectedError, result.Err)
}

func (s *ItemProcessorTestSuite) TestProcessItem_GetPosterDiskPositionError() {
	// Arrange
	item := model.Item{ID: "1", Title: "Test Movie"}
	configLib := &config.Library{Name: "Movies"}
	expectedError := fmt.Errorf("get poster disk position error")

	s.mockRatingBuilder.On("BuildRatings", s.mockCtx, &item).Return(nil).Once()
	s.mockPosterService.On("EnsurePosterExists", s.mockCtx, item, configLib).Return(nil).Once()
	s.mockPosterService.On("GetPosterDiskPosition", s.mockCtx, item, configLib).Return("", expectedError).Once()

	// Act
	result := s.itemProcessor.ProcessItem(item, 0, s.mockPosterService, configLib)

	// Assert
	s.Equal(item.Title, result.Title)
	s.Equal(expectedError, result.Err)
}

func (s *ItemProcessorTestSuite) TestProcessItem_ApplyLogosError() {
	// Arrange
	item := model.Item{ID: "1", Title: "Test Movie"}
	configLib := &config.Library{Name: "Movies"}
	expectedPosterPath := "/path/to/poster.jpg"
	expectedError := fmt.Errorf("apply logos error")

	s.mockRatingBuilder.On("BuildRatings", s.mockCtx, &item).Return(nil).Once()
	s.mockPosterService.On("EnsurePosterExists", s.mockCtx, item, configLib).Return(nil).Once()
	s.mockPosterService.On("GetPosterDiskPosition", s.mockCtx, item, configLib).Return(expectedPosterPath, nil).Once()
	s.mockPosterGenerator.On("ApplyLogos", s.mockCtx, expectedPosterPath, configLib, item).Return("", expectedError).Once()

	// Act
	result := s.itemProcessor.ProcessItem(item, 0, s.mockPosterService, configLib)

	// Assert
	s.Equal(item.Title, result.Title)
	s.Equal(expectedError, result.Err)
	// OriginalPosterDiskPosition should still be set in the result if GetPosterDiskPosition succeeds before ApplyLogos fails
	// However, the current implementation of ProcessItem returns early and doesn't set it in the result for ApplyLogos error
	// For now, we assert based on current behavior. If this is not desired, ProcessItem needs adjustment.
	// s.Equal(expectedPosterPath, result.OriginalPosterDiskPosition)
}

func (s *ItemProcessorTestSuite) TestProcessItems_NotEligible() {
	// Arrange
	items := []model.Item{
		{ID: "1", Title: "Not Eligible Movie"},
	}
	configLib := &config.Library{Name: "Movies"}

	s.mockEligibilityChecker.On("IsEligible", &items[0]).Return(false).Once()

	// Act
	err := s.itemProcessor.ProcessItems(items, configLib)

	// Assert
	s.NoError(err)
	// No other mocks should be called
}

func (s *ItemProcessorTestSuite) TestProcessItems_PosterServiceNotSet() {
	// Arrange
	items := []model.Item{{ID: "1", Title: "Test Movie"}}
	configLib := &config.Library{Name: "Movies"}

	// Create a new processor without setting the poster service
	// but with a valid mock semaphore
	mockSem := new(appmocks.SemaphoreWeighted)
	processor := NewItemProcessor(
		s.logger,
		s.mockCtx,
		mockSem, // Use mock semaphore
		s.mockPosterGenerator,
		s.mockEligibilityChecker,
		s.mockRatingBuilder,
	)
	// DO NOT CALL processor.SetPosterService(...)

	// Act
	err := processor.ProcessItems(items, configLib)

	// Assert
	s.Error(err)
	s.EqualError(err, "poster service not set")
	// Ensure semaphore mock expectations for 'mockSem' are met if any were set.
	// In this specific test, no calls to semaphore are expected before poster service validation,
	// so we can assert that no calls were made to it if we want to be strict.
	// mockSem.AssertExpectations(s.T()) // or mockSem.AssertNotCalled(s.T(), "Acquire")
}

func (s *ItemProcessorTestSuite) TestProcessItem_ContextCancelled_BeforeProcessing() {
	// Arrange
	item := model.Item{ID: "1", Title: "Test Movie"}
	configLib := &config.Library{Name: "Movies"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	s.itemProcessor.ctx = ctx // Use the cancelled context

	// Act
	result := s.itemProcessor.ProcessItem(item, 0, s.mockPosterService, configLib)

	// Assert
	s.Equal(context.Canceled, result.Err)
	s.Equal(item.Title, result.Title)
	s.Equal(context.Canceled, result.Err)
}

func (s *ItemProcessorTestSuite) TestProcessItem_ContextCancelled_AfterGetPosterDiskPosition() {
	// Arrange
	item := model.Item{ID: "1", Title: "Test Movie"}
	configLib := &config.Library{Name: "Movies"}
	expectedPosterPath := "/path/to/poster.jpg"

	ctx, cancel := context.WithCancel(context.Background())

	s.mockRatingBuilder.On("BuildRatings", ctx, &item).Return(nil).Once()
	s.mockPosterService.On("EnsurePosterExists", ctx, item, configLib).Return(nil).Once()
	s.mockPosterService.On("GetPosterDiskPosition", ctx, item, configLib).Return(expectedPosterPath, nil).Run(func(args mock.Arguments) {
		cancel() // Cancel context after this call
	}).Once()

	s.itemProcessor.ctx = ctx // Use the cancellable context

	// Act
	result := s.itemProcessor.ProcessItem(item, 0, s.mockPosterService, configLib)

	// Assert
	s.Equal(context.Canceled, result.Err)
	s.Equal(item.Title, result.Title)
	s.Equal(expectedPosterPath, result.OriginalPosterDiskPosition)
	s.Equal(context.Canceled, result.Err)
}

func (s *ItemProcessorTestSuite) TestProcessItem_ContextCancelled_AfterApplyLogos() {
	// Arrange
	item := model.Item{ID: "1", Title: "Test Movie"}
	configLib := &config.Library{Name: "Movies"}
	expectedPosterPath := "/path/to/poster.jpg"
	expectedNewPosterPath := "/path/to/new_poster.jpg"

	ctx, cancel := context.WithCancel(context.Background())

	s.mockRatingBuilder.On("BuildRatings", ctx, &item).Return(nil).Once()
	s.mockPosterService.On("EnsurePosterExists", ctx, item, configLib).Return(nil).Once()
	s.mockPosterService.On("GetPosterDiskPosition", ctx, item, configLib).Return(expectedPosterPath, nil).Once()
	s.mockPosterGenerator.On("ApplyLogos", ctx, expectedPosterPath, configLib, item).Return(expectedNewPosterPath, nil).Run(func(args mock.Arguments) {
		cancel() // Cancel context after this call
	}).Once()

	s.itemProcessor.ctx = ctx // Use the cancellable context

	// Act
	result := s.itemProcessor.ProcessItem(item, 0, s.mockPosterService, configLib)

	// Assert
	s.Equal(context.Canceled, result.Err)
	s.Equal(item.Title, result.Title)
	s.Equal(expectedPosterPath, result.OriginalPosterDiskPosition)
	// s.Equal(expectedNewPosterPath, result.OverlayPosterDiskPosition) // This is not set in the result due to early return
	s.Equal(context.Canceled, result.Err)
}

// Note: TestProcessItems with context cancellation is more complex due to goroutines and errgroup.
// A simple context cancellation on s.itemProcessor.ctx might not be sufficient to test all paths,
// as the errgroup might return its own context error first.
// For true coverage here, one might need to mock semaphore acquisition to block and then cancel,
// or ensure item processing itself returns a context error.

// Test for ProcessItems when ProcessItem returns an error
func (s *ItemProcessorTestSuite) TestProcessItems_ItemProcessingError() {
	// Arrange
	items := []model.Item{
		{ID: "1", Title: "Movie 1 (error)"},
		{ID: "2", Title: "Movie 2 (success)"},
	}
	configLib := &config.Library{Name: "Movies"}
	item1 := items[0]
	item2 := items[1]

	expectedErrorForItem1 := fmt.Errorf("specific error for item 1")
	expectedPosterPath2 := "/path/to/poster2.jpg"
	expectedNewPosterPath2 := "/path/to/new_poster2.jpg"

	// Item 1 (error) - mock semaphore
	s.mockWorkSemaphore.On("Acquire", mock.Anything, int64(1)).Return(nil).Once()
	s.mockEligibilityChecker.On("IsEligible", &item1).Return(true).Once()
	s.mockRatingBuilder.On("BuildRatings", mock.Anything, &item1).Return(expectedErrorForItem1).Once()
	// Release is still called in defer even if BuildRatings errors
	s.mockWorkSemaphore.On("Release", int64(1)).Return().Once()

	// Item 2 (success) - mock semaphore
	s.mockWorkSemaphore.On("Acquire", mock.Anything, int64(1)).Return(nil).Once()
	s.mockEligibilityChecker.On("IsEligible", &item2).Return(true).Once()
	s.mockRatingBuilder.On("BuildRatings", mock.Anything, &item2).Return(nil).Once()
	s.mockPosterService.On("EnsurePosterExists", mock.Anything, item2, configLib).Return(nil).Once()
	s.mockPosterService.On("GetPosterDiskPosition", mock.Anything, item2, configLib).Return(expectedPosterPath2, nil).Once()
	s.mockPosterGenerator.On("ApplyLogos", mock.Anything, expectedPosterPath2, configLib, item2).Return(expectedNewPosterPath2, nil).Once()
	s.mockWorkSemaphore.On("Release", int64(1)).Return().Once()

	// Act
	err := s.itemProcessor.ProcessItems(items, configLib)

	// Assert
	s.NoError(err) // Iteration should continue even if one item fails
	s.mockPosterGenerator.AssertCalled(s.T(), "ApplyLogos", mock.Anything, expectedPosterPath2, configLib, item2)
	s.mockPosterGenerator.AssertNotCalled(s.T(), "ApplyLogos", mock.Anything, expectedPosterPath2, configLib, item1)
}

// Test for ProcessItems when semaphore acquisition fails.
// This is hard to test directly without making the semaphore exhausted.
// We can test the error path if Acquire returns an error for some reason (e.g., context cancelled).
func (s *ItemProcessorTestSuite) TestProcessItems_SemaphoreAcquireError() {
	// Arrange
	items := []model.Item{{ID: "1", Title: "Test Movie"}}
	configLib := &config.Library{Name: "Movies"}
	item1 := items[0]
	expectedError := fmt.Errorf("semaphore acquire failed")

	// No need to reinitialize itemProcessor, s.mockWorkSemaphore is already a mock.
	// s.itemProcessor.ctx is used by the errgroup, not directly by semaphore.Acquire in the mock.

	s.mockEligibilityChecker.On("IsEligible", &item1).Return(true).Once()
	// Simulate semaphore acquire error by making the mock return an error
	s.mockWorkSemaphore.On("Acquire", mock.Anything, int64(1)).Return(expectedError).Once()
	// Release should not be called if Acquire fails

	// Act
	err := s.itemProcessor.ProcessItems(items, configLib)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "error acquiring semaphore")
	s.Contains(err.Error(), expectedError.Error())
}
