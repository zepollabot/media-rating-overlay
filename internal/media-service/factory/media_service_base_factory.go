package factory

import (
	"time"

	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	media "github.com/zepollabot/media-rating-overlay/internal/media-service"
	plexClient "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/client"
	plexFilters "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/filter"
	plexItem "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/item"
	plex "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/library"
)

// Clock interface defines the time-related operations needed
type Clock interface {
	Now() time.Time
}

// BaseMediaServiceFactory handles only media-service specific components
type MediaServiceBaseFactory struct {
	logger *zap.Logger
	clock  Clock
	config *config.Config
}

func NewMediaServiceBaseFactory(
	logger *zap.Logger,
	clock Clock,
	config *config.Config,
) *MediaServiceBaseFactory {
	return &MediaServiceBaseFactory{
		logger: logger,
		clock:  clock,
		config: config,
	}
}

// BuildPlexComponents returns all the Plex-specific components without the poster service
func (f *MediaServiceBaseFactory) BuildPlexComponents() (
	media.MediaClient,
	media.LibraryService,
	media.ItemService,
	error,
) {
	f.logger.Info("Building Plex media service components")

	plexClient, err := plexClient.NewPlexClient(&f.config.Plex, &f.config.HTTPClient, f.config.Logger.LogFilePath, f.logger)
	if err != nil {
		f.logger.Error("error creating plex client", zap.Error(err))
		return nil, nil, nil, err
	}

	libraryService := plex.NewPlexLibraryService(plexClient, f.logger)
	filtersService := plexFilters.NewPlexFiltersService(f.clock)
	itemService := plexItem.NewPlexItemService(plexClient, f.logger, filtersService)

	return plexClient, libraryService, itemService, nil
}

func (f *MediaServiceBaseFactory) GetLibraries() []config.Library {
	return f.config.Plex.Libraries
}
