package core

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	media "github.com/zepollabot/media-rating-overlay/internal/media-service"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
)

// ServiceContext holds the services required for processing a library
type ServiceContext struct {
	LibrariesService media.LibraryService
	ItemsService     media.ItemService
	PostersService   media.PosterService
}

// Processor handles the processing of libraries
type LibraryProcessor struct {
	logger        *zap.Logger
	itemProcessor ItemProcessor
	// Default timeout for operations
	defaultTimeout time.Duration
}

// LibraryProcessorConfig holds the configuration for a LibraryProcessor
type LibraryProcessorConfig struct {
	DefaultTimeout time.Duration
}

// DefaultLibraryProcessorConfig returns a default configuration for LibraryProcessor
func DefaultLibraryProcessorConfig() *LibraryProcessorConfig {
	return &LibraryProcessorConfig{
		DefaultTimeout: 30 * time.Second,
	}
}

// NewLibrariesProcessor creates a new library processor
func NewLibrariesProcessor(
	logger *zap.Logger,
	itemProcessor ItemProcessor,
	config *LibraryProcessorConfig,
) *LibraryProcessor {
	if config == nil {
		config = DefaultLibraryProcessorConfig()
	}

	return &LibraryProcessor{
		logger:         logger,
		itemProcessor:  itemProcessor,
		defaultTimeout: config.DefaultTimeout,
	}
}

// ProcessLibrary processes a single library
func (lp *LibraryProcessor) ProcessLibrary(
	ctx context.Context,
	configLibrary *config.Library,
	mediaServiceLibraries *[]model.Library,
	serviceCtx ServiceContext,
) error {
	// Create a context with timeout for this library processing
	ctx, cancel := context.WithTimeout(ctx, lp.defaultTimeout)
	defer cancel()

	// Check for context cancellation before starting
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled before processing: %w", err)
	}

	if err := lp.validateServiceContext(serviceCtx); err != nil {
		return err
	}

	lp.logger.Info("Processing library", zap.String("library", configLibrary.Name))
	// Check if the library is active
	if !configLibrary.Enabled {
		lp.logger.Info("Library not active, skipped", zap.String("library", configLibrary.Name))
		return nil
	}

	// Find the corresponding library in the Media Service libraries list
	library, found := lo.Find(*mediaServiceLibraries, func(lib model.Library) bool {
		return lib.Name == configLibrary.Name
	})

	if !found {
		return fmt.Errorf("library '%s' not found or not active", configLibrary.Name)
	}

	lp.logger.Info("Library found",
		zap.String("ID", library.ID),
		zap.String("Name", library.Name),
	)

	// Check for context cancellation before retrieving items
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled before retrieving items: %w", err)
	}

	// Retrieve the library items
	libraryItems, err := serviceCtx.ItemsService.GetItems(ctx, library, configLibrary)
	if err != nil {
		return fmt.Errorf("unable to retrieve items: %w", err)
	}

	if len(libraryItems) == 0 {
		lp.logger.Info("No items found in the library", zap.String("library", library.Name))
		return nil
	}

	lp.logger.Info("Processing items", zap.Int("count", len(libraryItems)))

	// Check for context cancellation before processing items
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled before processing items: %w", err)
	}

	// set posters service for the item processor
	lp.itemProcessor.SetPosterService(serviceCtx.PostersService)

	// Process the library items
	err = lp.itemProcessor.ProcessItems(libraryItems, configLibrary)
	if err != nil {
		return fmt.Errorf("error processing items: %w", err)
	}

	// Check for context cancellation before refreshing
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled before refreshing library: %w", err)
	}

	// Refresh the library if requested
	if configLibrary.Refresh {
		lp.logger.Info("Refreshing library...", zap.String("library", library.Name))
		if err := serviceCtx.LibrariesService.RefreshLibrary(ctx, library.ID, true); err != nil {
			return fmt.Errorf("unable to refresh library: %w", err)
		}
		lp.logger.Info("Library refreshed successfully", zap.String("library", library.Name))
	}

	return nil
}

func (lp *LibraryProcessor) validateServiceContext(serviceCtx ServiceContext) error {
	if serviceCtx.LibrariesService == nil {
		lp.logger.Error("Libraries service not set")
		return fmt.Errorf("libraries service not set")
	}

	if serviceCtx.ItemsService == nil {
		lp.logger.Error("Items service not set")
		return fmt.Errorf("items service not set")
	}

	if serviceCtx.PostersService == nil {
		lp.logger.Error("Posters service not set")
		return fmt.Errorf("posters service not set")
	}

	return nil
}
