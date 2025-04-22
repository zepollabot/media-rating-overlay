package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	coremocks "github.com/zepollabot/media-rating-overlay/internal/app/mocks" // Alias for core mocks
	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	mediamocks "github.com/zepollabot/media-rating-overlay/internal/media-service/mocks"
	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type LibraryProcessorTestSuite struct {
	suite.Suite
	logger                 *zap.Logger
	realItemProcessor      *ItemProcessor // Actual ItemProcessor struct
	mockPosterGenerator    *coremocks.PosterGenerator
	mockEligibilityChecker *coremocks.ItemEligibilityChecker
	mockRatingBuilder      *coremocks.RatingBuilder
	mockLibrariesService   *mediamocks.LibraryService
	mockItemsService       *mediamocks.ItemService
	mockPostersService     *mediamocks.PosterService
	processor              *LibraryProcessor

	// Fields for managing ItemProcessor's context and semaphore
	itemProcCtx       context.Context
	itemProcCancel    context.CancelFunc
	realItemSemaphore *semaphore.Weighted
}

func (s *LibraryProcessorTestSuite) SetupTest() {
	s.logger = zap.NewNop()

	// Mocks for ItemProcessor's dependencies
	s.mockPosterGenerator = coremocks.NewPosterGenerator(s.T())
	s.mockEligibilityChecker = coremocks.NewItemEligibilityChecker(s.T())
	s.mockRatingBuilder = coremocks.NewRatingBuilder(s.T())

	// Cancellable context and initialized semaphore for the ItemProcessor instance
	s.itemProcCtx, s.itemProcCancel = context.WithCancel(context.Background())
	s.realItemSemaphore = semaphore.NewWeighted(5) // Initialize semaphore

	s.realItemProcessor = NewItemProcessor(
		s.logger,
		s.itemProcCtx,       // Use the cancellable context
		s.realItemSemaphore, // Use the initialized semaphore
		s.mockPosterGenerator,
		s.mockEligibilityChecker,
		s.mockRatingBuilder,
	)

	s.mockLibrariesService = mediamocks.NewLibraryService(s.T())
	s.mockItemsService = mediamocks.NewItemService(s.T())
	s.mockPostersService = mediamocks.NewPosterService(s.T())

	cfg := DefaultLibraryProcessorConfig()
	s.processor = NewLibrariesProcessor(s.logger, *s.realItemProcessor, cfg)
}

func (s *LibraryProcessorTestSuite) TearDownTest() {
	s.itemProcCancel() // Cancel the ItemProcessor's context after each test
	s.mockPosterGenerator.AssertExpectations(s.T())
	s.mockEligibilityChecker.AssertExpectations(s.T())
	s.mockRatingBuilder.AssertExpectations(s.T())
	s.mockLibrariesService.AssertExpectations(s.T())
	s.mockItemsService.AssertExpectations(s.T())
	s.mockPostersService.AssertExpectations(s.T())
}

func TestLibraryProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(LibraryProcessorTestSuite))
}

// TestNewLibrariesProcessor tests the constructor for LibraryProcessor.
func (s *LibraryProcessorTestSuite) TestNewLibrariesProcessor() {
	s.Run("DefaultConfig", func() {
		logger := zap.NewNop()

		tempMockPosterGen := coremocks.NewPosterGenerator(s.T())
		tempMockEligChecker := coremocks.NewItemEligibilityChecker(s.T())
		tempMockRatingBuilder := coremocks.NewRatingBuilder(s.T())
		dummyItemProcessor := NewItemProcessor(logger, context.Background(), nil, tempMockPosterGen, tempMockEligChecker, tempMockRatingBuilder)

		processor := NewLibrariesProcessor(logger, *dummyItemProcessor, nil)

		s.Assert().NotNil(processor)
		s.Assert().Equal(logger, processor.logger)
		s.Assert().Equal(dummyItemProcessor.logger, processor.itemProcessor.logger)
		s.Assert().Equal(30*time.Second, processor.defaultTimeout)

		tempMockPosterGen.AssertExpectations(s.T())
		tempMockEligChecker.AssertExpectations(s.T())
		tempMockRatingBuilder.AssertExpectations(s.T())
	})

	s.Run("CustomConfig", func() {
		logger := zap.NewNop()
		customTimeout := 15 * time.Second
		cfg := &LibraryProcessorConfig{
			DefaultTimeout: customTimeout,
		}

		tempMockPosterGen := coremocks.NewPosterGenerator(s.T())
		tempMockEligChecker := coremocks.NewItemEligibilityChecker(s.T())
		tempMockRatingBuilder := coremocks.NewRatingBuilder(s.T())
		dummyItemProcessor := NewItemProcessor(logger, context.Background(), nil, tempMockPosterGen, tempMockEligChecker, tempMockRatingBuilder)

		processor := NewLibrariesProcessor(logger, *dummyItemProcessor, cfg)

		s.Assert().NotNil(processor)
		s.Assert().Equal(logger, processor.logger)
		s.Assert().Equal(dummyItemProcessor.logger, processor.itemProcessor.logger)
		s.Assert().Equal(customTimeout, processor.defaultTimeout)

		tempMockPosterGen.AssertExpectations(s.T())
		tempMockEligChecker.AssertExpectations(s.T())
		tempMockRatingBuilder.AssertExpectations(s.T())
	})
}

// TestProcessLibrary_Success tests the successful processing of a library.
func (s *LibraryProcessorTestSuite) TestProcessLibrary_Success() {
	ctx := context.Background()
	configLibrary := &config.Library{Name: "Movies", Enabled: true, Refresh: true}
	mediaLib := model.Library{ID: "lib1", Name: "Movies"}
	mediaServiceLibraries := &[]model.Library{mediaLib}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}
	libraryItems := []model.Item{{ID: "item1", Title: "Movie 1"}}

	// A. Arrange: Define mock behaviors
	// 1. ItemsService.GetItems successfully returns items
	s.mockItemsService.On("GetItems", mock.Anything, mediaLib, configLibrary).Return(libraryItems, nil).Once()
	s.mockPostersService.On("GetPosterDiskPosition", mock.Anything, mock.AnythingOfType("model.Item"), configLibrary).Return("", nil).Maybe()

	// 2. ItemProcessor.ProcessItems related mocks (indirectly testing ProcessItems success)
	//    ItemProcessor.SetPosterService will be called with s.mockPostersService.
	//    ItemProcessor.ProcessItems will be called. For it to succeed, its dependencies must not error out in a critical way.
	//    These are .Maybe() because their exact invocation count might vary depending on ItemProcessor logic for one item.
	s.mockEligibilityChecker.On("IsEligible", mock.AnythingOfType("*model.Item")).Return(true).Maybe()
	s.mockRatingBuilder.On("BuildRatings", mock.Anything, mock.AnythingOfType("*model.Item")).Return(nil).Maybe()
	//    If ItemProcessor calls its internal posterService (which would be s.mockPostersService after SetPosterService)
	s.mockPostersService.On("EnsurePosterExists", mock.Anything, mock.AnythingOfType("model.Item"), configLibrary).Return(nil).Maybe()
	//    If ItemProcessor calls PosterGenerator
	s.mockPosterGenerator.On("ApplyLogos", mock.Anything, mock.AnythingOfType("string"), configLibrary, mock.AnythingOfType("model.Item")).Return("new/path.jpg", nil).Maybe()

	// 3. LibrariesService.RefreshLibrary successfully refreshes
	s.mockLibrariesService.On("RefreshLibrary", mock.Anything, mediaLib.ID, true).Return(nil).Once()

	// Act
	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)

	// Assert
	s.Assert().NoError(err)
	// Mock expectations are asserted in TearDownTest
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_ContextCancelled_BeforeProcessing() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	configLibrary := &config.Library{Name: "Movies", Enabled: true}
	// Pass a non-nil pointer to an empty slice for mediaServiceLibraries
	mediaServiceLibraries := &[]model.Library{}
	// serviceCtx can be minimal as it might not be reached, but ensure fields are valid if processor accesses them before erroring
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService, // Provide mocks to avoid nil panics if validation is attempted
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}

	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "context cancelled before processing")
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_ContextCancelled_DuringGetItems() { // Renaming for clarity
	ctx := context.Background() // Parent context for ProcessLibrary

	configLibrary := &config.Library{Name: "Movies", Enabled: true}
	mediaLib := model.Library{ID: "lib1", Name: "Movies"}
	mediaServiceLibraries := &[]model.Library{mediaLib}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}

	// Expect GetItems to be called
	// Make it return a context.Canceled error
	contextCancelledError := context.Canceled // Or fmt.Errorf("wrapped: %w", context.Canceled)
	s.mockItemsService.On("GetItems", mock.Anything, mediaLib, configLibrary).
		Return(nil, contextCancelledError).Once()

	// Use the default processor from SetupTest
	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)

	s.Assert().Error(err)
	s.Assert().True(errors.Is(err, context.Canceled), "Error should wrap context.Canceled. Got: %v", err)
	// LibraryProcessor wraps the error from GetItems: "unable to retrieve items: %w"
	s.Assert().Contains(err.Error(), "unable to retrieve items:", "Error message structure mismatch. Got: %v", err)
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_ServiceValidation_LibrariesServiceNil() {
	ctx := context.Background()
	configLibrary := &config.Library{Name: "Movies", Enabled: true}
	mediaServiceLibraries := &[]model.Library{}
	serviceCtx := ServiceContext{
		LibrariesService: nil, // Nil service
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}

	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)
	s.Assert().Error(err)
	s.Assert().EqualError(err, "libraries service not set")
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_ServiceValidation_ItemsServiceNil() {
	ctx := context.Background()
	configLibrary := &config.Library{Name: "Movies", Enabled: true}
	mediaServiceLibraries := &[]model.Library{}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     nil, // Nil service
		PostersService:   s.mockPostersService,
	}

	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)
	s.Assert().Error(err)
	s.Assert().EqualError(err, "items service not set")
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_ServiceValidation_PostersServiceNil() {
	ctx := context.Background()
	configLibrary := &config.Library{Name: "Movies", Enabled: true}
	mediaServiceLibraries := &[]model.Library{}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     s.mockItemsService,
		PostersService:   nil, // Nil service
	}

	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)
	s.Assert().Error(err)
	s.Assert().EqualError(err, "posters service not set")
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_LibraryNotEnabled() {
	ctx := context.Background()
	configLibrary := &config.Library{Name: "Movies", Enabled: false} // Not enabled
	mediaServiceLibraries := &[]model.Library{}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}

	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)
	s.Assert().NoError(err)
	// Ensure no service calls were made that would indicate processing beyond the 'Enabled' check.
	s.mockItemsService.AssertNotCalled(s.T(), "GetItems", mock.Anything, mock.Anything, mock.Anything)
	s.mockLibrariesService.AssertNotCalled(s.T(), "RefreshLibrary", mock.Anything, mock.Anything, mock.Anything)
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_LibraryNotFound() {
	ctx := context.Background()
	configLibrary := &config.Library{Name: "NonExistentShows", Enabled: true}
	mediaLib := model.Library{ID: "lib1", Name: "Movies"} // Media service has 'Movies', config asks for 'NonExistentShows'
	mediaServiceLibraries := &[]model.Library{mediaLib}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}

	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)
	s.Assert().Error(err)
	s.Assert().EqualError(err, "library 'NonExistentShows' not found or not active")
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_GetItemsError() {
	ctx := context.Background()
	configLibrary := &config.Library{Name: "Movies", Enabled: true}
	mediaLib := model.Library{ID: "lib1", Name: "Movies"}
	mediaServiceLibraries := &[]model.Library{mediaLib}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}
	expectedError := errors.New("get items failed")

	s.mockItemsService.On("GetItems", mock.Anything, mediaLib, configLibrary).Return(nil, expectedError).Once()

	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "unable to retrieve items: get items failed")
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_NoItemsFound() {
	ctx := context.Background()
	configLibrary := &config.Library{Name: "Movies", Enabled: true, Refresh: false} // No refresh for simplicity
	mediaLib := model.Library{ID: "lib1", Name: "Movies"}
	mediaServiceLibraries := &[]model.Library{mediaLib}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}
	// GetItems returns empty slice
	s.mockItemsService.On("GetItems", mock.Anything, mediaLib, configLibrary).Return([]model.Item{}, nil).Once()

	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)
	s.Assert().NoError(err)
	// ItemProcessor.ProcessItems should not be called if no items are found.
	// We can check that by ensuring none of its core operational mocks (like RatingBuilder) were called.
	s.mockRatingBuilder.AssertNotCalled(s.T(), "BuildRatings", mock.Anything, mock.Anything)
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_ProcessItemsError() {
	ctx := context.Background() // Outer context for LibraryProcessor
	configLibrary := &config.Library{Name: "Movies", Enabled: true, Refresh: false}
	mediaLib := model.Library{ID: "lib1", Name: "Movies"}
	mediaServiceLibraries := &[]model.Library{mediaLib}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}
	libraryItems := []model.Item{{ID: "item1", Title: "Movie 1"}}

	s.mockItemsService.On("GetItems", mock.Anything, mediaLib, configLibrary).Return(libraryItems, nil).Once()

	// Expect IsEligible to be called, and let the item be eligible.
	// This allows ProcessItems to proceed to a point where the canceled s.itemProcCtx causes an error.
	s.mockEligibilityChecker.On("IsEligible", mock.AnythingOfType("*model.Item")).Return(true).Once()

	// Cancel the ItemProcessor's own context *before* ProcessItems is called by LibraryProcessor.
	// This should cause ItemProcessor.ProcessItems to return an error.
	s.itemProcCancel()

	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)

	s.Assert().Error(err) // Expect an error from LibraryProcessor

	// Check that context.Canceled is part of the error chain.
	s.Assert().True(errors.Is(err, context.Canceled), "Error should wrap context.Canceled. Got: %v", err)

	// Check for the expected wrapping messages.
	s.Assert().Contains(err.Error(), "error processing items: error during parallel processing:", "Error message structure mismatch. Got: %v", err)
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_RefreshLibraryError() {
	ctx := context.Background()
	configLibrary := &config.Library{Name: "Movies", Enabled: true, Refresh: true} // Refresh enabled
	mediaLib := model.Library{ID: "lib1", Name: "Movies"}
	mediaServiceLibraries := &[]model.Library{mediaLib}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}
	libraryItems := []model.Item{{ID: "item1"}}
	expectedError := errors.New("refresh library failed")

	s.mockItemsService.On("GetItems", mock.Anything, mediaLib, configLibrary).Return(libraryItems, nil).Once()

	// Assume ItemProcessor.ProcessItems succeeds
	s.mockRatingBuilder.On("BuildRatings", mock.Anything, mock.AnythingOfType("*model.Item")).Return(nil).Maybe()
	s.mockEligibilityChecker.On("IsEligible", mock.AnythingOfType("*model.Item")).Return(true).Maybe()
	s.mockPosterGenerator.On("ApplyLogos", mock.Anything, mock.AnythingOfType("string"), configLibrary, mock.AnythingOfType("model.Item")).Return("new/path.jpg", nil).Maybe()
	s.mockPostersService.On("EnsurePosterExists", mock.Anything, mock.AnythingOfType("model.Item"), configLibrary).Return(nil).Maybe()
	s.mockPostersService.On("GetPosterDiskPosition", mock.Anything, mock.AnythingOfType("model.Item"), configLibrary).Return("", nil).Maybe()

	s.mockLibrariesService.On("RefreshLibrary", mock.Anything, mediaLib.ID, true).Return(expectedError).Once()

	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "unable to refresh library: refresh library failed")
}

func (s *LibraryProcessorTestSuite) TestProcessLibrary_RefreshDisabled() {
	ctx := context.Background()
	configLibrary := &config.Library{Name: "Movies", Enabled: true, Refresh: false} // Refresh disabled
	mediaLib := model.Library{ID: "lib1", Name: "Movies"}
	mediaServiceLibraries := &[]model.Library{mediaLib}
	serviceCtx := ServiceContext{
		LibrariesService: s.mockLibrariesService,
		ItemsService:     s.mockItemsService,
		PostersService:   s.mockPostersService,
	}
	libraryItems := []model.Item{{ID: "item1"}}

	s.mockItemsService.On("GetItems", mock.Anything, mediaLib, configLibrary).Return(libraryItems, nil).Once()

	// Assume ItemProcessor.ProcessItems succeeds
	s.mockRatingBuilder.On("BuildRatings", mock.Anything, mock.AnythingOfType("*model.Item")).Return(nil).Maybe()
	s.mockEligibilityChecker.On("IsEligible", mock.AnythingOfType("*model.Item")).Return(true).Maybe()
	s.mockPosterGenerator.On("ApplyLogos", mock.Anything, mock.AnythingOfType("string"), configLibrary, mock.AnythingOfType("model.Item")).Return("new/path.jpg", nil).Maybe()
	s.mockPostersService.On("EnsurePosterExists", mock.Anything, mock.AnythingOfType("model.Item"), configLibrary).Return(nil).Maybe()
	s.mockPostersService.On("GetPosterDiskPosition", mock.Anything, mock.AnythingOfType("model.Item"), configLibrary).Return("", nil).Maybe()

	// RefreshLibrary should NOT be called
	err := s.processor.ProcessLibrary(ctx, configLibrary, mediaServiceLibraries, serviceCtx)
	s.Assert().NoError(err)
	s.mockLibrariesService.AssertNotCalled(s.T(), "RefreshLibrary", mock.Anything, mock.Anything, mock.Anything)
}
