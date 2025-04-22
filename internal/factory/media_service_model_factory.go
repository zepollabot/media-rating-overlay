package factory

import (
	"fmt"

	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/media-service"
	mediaModel "github.com/zepollabot/media-rating-overlay/internal/media-service/model"
	plexPoster "github.com/zepollabot/media-rating-overlay/internal/media-service/platform/plex/poster"
	"github.com/zepollabot/media-rating-overlay/internal/processor/file"
)

type MediaServiceBaseFactory interface {
	BuildPlexComponents() (media.MediaClient, media.LibraryService, media.ItemService, error)
	GetLibraries() []config.Library
}

// MediaServiceFactory composes all components together
type MediaServiceModelFactory struct {
	baseFactory MediaServiceBaseFactory
	logger      *zap.Logger
}

func NewMediaServiceModelFactory(
	logger *zap.Logger,
	mediaFactory MediaServiceBaseFactory,
) *MediaServiceModelFactory {
	return &MediaServiceModelFactory{
		baseFactory: mediaFactory,
		logger:      logger,
	}
}

func (f *MediaServiceModelFactory) Create(serviceName string) (mediaModel.MediaService, error) {
	switch serviceName {
	case mediaModel.MediaServicePlex:
		return f.buildPlexMediaService()
	default:
		return mediaModel.MediaService{}, fmt.Errorf("unsupported media service: %s", serviceName)
	}
}

func (f *MediaServiceModelFactory) buildPlexMediaService() (mediaModel.MediaService, error) {
	// Get base components from the base factory
	plexClient, libraryService, itemService, err := f.baseFactory.BuildPlexComponents()
	if err != nil {
		return mediaModel.MediaService{}, err
	}

	// Create processor-specific components
	fileManager := file.NewFileManager(f.logger)

	// Create the poster service with the file manager
	posterService := plexPoster.NewPlexPosterService(plexClient, f.logger, fileManager)

	f.logger.Info("Plex media service initialized")
	return mediaModel.MediaService{
		Name:           mediaModel.MediaServicePlex,
		Libraries:      f.baseFactory.GetLibraries(),
		Client:         plexClient,
		LibraryService: libraryService,
		ItemService:    itemService,
		PosterService:  posterService,
	}, nil
}
