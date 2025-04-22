package core

import (
	"context"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/constant"
	"github.com/zepollabot/media-rating-overlay/internal/factory"
	mediaFactory "github.com/zepollabot/media-rating-overlay/internal/media-service/factory"
	mediaModel "github.com/zepollabot/media-rating-overlay/internal/media-service/model"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
	processorFactory "github.com/zepollabot/media-rating-overlay/internal/processor/factory"
	"github.com/zepollabot/media-rating-overlay/internal/processor/rating"
	ratingFactory "github.com/zepollabot/media-rating-overlay/internal/rating-service/factory"
	"github.com/zepollabot/media-rating-overlay/internal/rating-service/item"
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

// VisualDebug is a constant that determines whether to show debug visualization
const VisualDebug = false

type MediaServiceModelFactory interface {
	Create(mediaServiceType string) (mediaModel.MediaService, error)
}

// ServiceInitializer handles the initialization of all services
type ServiceInitializer struct {
	logger                   *zap.Logger
	config                   *config.Config
	mediaServices            []mediaModel.MediaService
	ratingPlatformServices   []ratingModel.RatingService
	libraryProcessor         *LibraryProcessor
	itemProcessor            *ItemProcessor
	ctx                      context.Context
	workSemaphore            *semaphore.Weighted
	mediaServiceBaseFactory  factory.MediaServiceBaseFactory
	MediaServiceModelFactory MediaServiceModelFactory
	RatingServiceBaseFactory factory.RatingServiceBaseFactory
}

// NewServiceInitializer creates a new service initializer
func NewServiceInitializer(logger *zap.Logger, config *config.Config, ctx context.Context, workSemaphore *semaphore.Weighted) *ServiceInitializer {
	// Initialize core fields
	si := &ServiceInitializer{
		logger:                 logger,
		config:                 config,
		mediaServices:          make([]mediaModel.MediaService, 0),
		ratingPlatformServices: make([]ratingModel.RatingService, 0),
		ctx:                    ctx,
		workSemaphore:          workSemaphore,
	}

	// Create and assign factories
	clock := model.RealClock{}
	si.mediaServiceBaseFactory = mediaFactory.NewMediaServiceBaseFactory(si.logger, clock, si.config)
	si.MediaServiceModelFactory = factory.NewMediaServiceModelFactory(si.logger, si.mediaServiceBaseFactory)
	si.RatingServiceBaseFactory = ratingFactory.NewRatingServiceBaseFactory(si.logger, si.config)

	return si
}

// InitializeServices initializes all services
func (si *ServiceInitializer) InitializeServices() error {
	si.logger.Info("Initializing services")

	if err := si.buildMediaPlatformServicesArray(); err != nil {
		return err
	}

	if err := si.buildRatingPlatformServicesArray(); err != nil {
		return err
	}

	si.initializeProcessors()
	return nil
}

// GetMediaServices returns the initialized media services
func (si *ServiceInitializer) GetMediaServices() []mediaModel.MediaService {
	return si.mediaServices
}

// GetRatingPlatformServices returns the initialized rating platform services
func (si *ServiceInitializer) GetRatingPlatformServices() []ratingModel.RatingService {
	return si.ratingPlatformServices
}

// GetLibraryProcessor returns the initialized library processor
func (si *ServiceInitializer) GetLibraryProcessor() *LibraryProcessor {
	return si.libraryProcessor
}

// GetItemProcessor returns the initialized item processor
func (si *ServiceInitializer) GetItemProcessor() *ItemProcessor {
	return si.itemProcessor
}

func (si *ServiceInitializer) buildMediaPlatformServicesArray() error {
	si.logger.Info("Evaluating media services configured..")

	if si.config.Plex.Enabled {
		si.logger.Debug("Initializing Plex media service")

		mediaService, err := si.MediaServiceModelFactory.Create(mediaModel.MediaServicePlex)
		if err != nil {
			si.logger.Error("error creating media service", zap.Error(err))
			return err
		}

		si.logger.Debug("Plex media service configured", zap.Any("mediaService", mediaService))
		si.mediaServices = append(si.mediaServices, mediaService)
	}

	// TODO: Add other media services here, like Kodi, Jellyfin, etc.
	// if si.config.Kodi.Enabled {
	// 	si.mediaServices = append(si.mediaServices, models.MediaService{
	// 		Name:      media.MediaServiceKodi,
	// 		Libraries: si.config.Kodi.Libraries,
	// 		Client:    kodiClient,
	// 	})
	// }

	si.logger.Info("Found configuration for the following media services",
		zap.Strings("mediaServices", lo.Map(si.mediaServices, func(ms mediaModel.MediaService, _ int) string { return ms.Name })),
	)

	return nil
}

func (si *ServiceInitializer) buildRatingPlatformServicesArray() error {
	si.logger.Info("Evaluating rating services configured..")

	ratingPlatformServiceModelFactory := factory.NewRatingPlatformServiceModelFactory(si.logger, si.RatingServiceBaseFactory, VisualDebug)

	// Initialize TMDB rating service
	si.logger.Debug("Initializing TMDB rating platform service")

	tmdbRatingService, err := ratingPlatformServiceModelFactory.Create(constant.RatingServiceTMDB)
	if err != nil {
		si.logger.Error("error creating rating service", zap.Error(err))
		return err
	}

	si.logger.Debug("TMDB rating platform service configured", zap.Any("ratingService", tmdbRatingService))
	si.ratingPlatformServices = append(si.ratingPlatformServices, tmdbRatingService)

	// Initialize Rotten Tomatoes rating service
	si.logger.Debug("Initializing Rotten Tomatoes rating platform service")

	rottenTomatoesRatingService, err := ratingPlatformServiceModelFactory.Create(constant.RatingServiceRottenTomatoes)
	if err != nil {
		si.logger.Error("error creating rating service", zap.Error(err))
		return err
	}

	si.logger.Debug("Rotten Tomatoes rating platform service configured", zap.Any("ratingService", rottenTomatoesRatingService))
	si.ratingPlatformServices = append(si.ratingPlatformServices, rottenTomatoesRatingService)

	// Initialize IMDB rating service
	si.logger.Debug("Initializing IMDB rating platform service")

	imdbRatingService, err := ratingPlatformServiceModelFactory.Create(constant.RatingServiceIMDB)
	if err != nil {
		si.logger.Error("error creating rating service", zap.Error(err))
		return err
	}

	si.logger.Debug("IMDB rating platform service configured", zap.Any("ratingService", imdbRatingService))
	si.ratingPlatformServices = append(si.ratingPlatformServices, imdbRatingService)

	si.logger.Info("Found configuration for the following rating platform services",
		zap.Strings("ratingPlatformServices", lo.Map(si.ratingPlatformServices, func(rs ratingModel.RatingService, _ int) string { return rs.Name })),
	)

	return nil
}

func (si *ServiceInitializer) initializeProcessors() {
	si.logger.Info("Initializing processors")

	// Initialize item processor
	eligibilityService := item.NewItemEligibilityService(si.ratingPlatformServices, si.logger)
	ratingBuilderService := rating.NewRatingBuilderService(si.ratingPlatformServices, si.logger)

	// Create poster generator using factory
	posterGeneratorFactory := processorFactory.NewPosterGeneratorFactory(si.logger, si.ratingPlatformServices, VisualDebug)
	posterGenerator := posterGeneratorFactory.Create()

	si.itemProcessor = NewItemProcessor(
		si.logger,
		si.ctx,
		si.workSemaphore,
		posterGenerator,
		eligibilityService,
		ratingBuilderService,
	)

	// Initialize library processor with timeout from config
	libraryConfig := &LibraryProcessorConfig{
		DefaultTimeout: si.config.Performance.LibraryProcessingTimeout,
	}
	si.libraryProcessor = NewLibrariesProcessor(
		si.logger,
		*si.itemProcessor,
		libraryConfig,
	)
	si.logger.Info("Processors initialized successfully")
}
