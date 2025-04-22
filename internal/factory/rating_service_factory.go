package factory

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/constant"
	"github.com/zepollabot/media-rating-overlay/internal/model"
	"github.com/zepollabot/media-rating-overlay/internal/processor/logo"
	"github.com/zepollabot/media-rating-overlay/internal/processor/text"

	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
	logoIMDB "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/imdb/logo"
	logoRottenTomatoes "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/rotten-tomatoes/logo"
	logoTmdb "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/logo"
)

// RatingServiceBaseFactory interface is defined in this file.
type RatingServiceBaseFactory interface {
	BuildTMDBComponents() (ratingModel.RatingService, error)
	BuildRottenTomatoesComponents() (ratingModel.RatingService, error)
	BuildIMDBComponents() (ratingModel.RatingService, error)
}

type RatingPlatformServiceModelFactory struct {
	baseFactory RatingServiceBaseFactory
	logger      *zap.Logger
	visualDebug bool
}

// MODIFIED: Changed ratingServiceBaseFactory parameter to interface type RatingServiceBaseFactory
func NewRatingPlatformServiceModelFactory(logger *zap.Logger, ratingServiceBaseFactory RatingServiceBaseFactory, visualDebug bool) *RatingPlatformServiceModelFactory {
	return &RatingPlatformServiceModelFactory{
		baseFactory: ratingServiceBaseFactory,
		logger:      logger,
		visualDebug: visualDebug,
	}
}

func (f *RatingPlatformServiceModelFactory) Create(serviceName string) (ratingModel.RatingService, error) {
	defaultPosterConfig := model.PosterConfigWithDefaultValues()
	imageCreator := logo.NewLogoImageCreator(f.logger, f.visualDebug)
	textGuesser := text.NewTextGuesser(f.logger)
	textService := text.NewTextService(f.logger, textGuesser)
	logoCreator := logo.NewLogoCreator(f.logger, imageCreator, textService, defaultPosterConfig)

	switch serviceName {
	case constant.RatingServiceTMDB:
		return f.buildTMDBRatingService(logoCreator, defaultPosterConfig)
	case constant.RatingServiceRottenTomatoes:
		return f.buildRottenTomatoesRatingService(logoCreator, defaultPosterConfig)
	case constant.RatingServiceIMDB:
		return f.buildIMDBRatingService(logoCreator, defaultPosterConfig)
	default:
		return ratingModel.RatingService{}, fmt.Errorf("unsupported rating service: %s", serviceName)
	}
}

func (f *RatingPlatformServiceModelFactory) buildTMDBRatingService(logoCreator *logo.LogoCreator, defaultPosterConfig *model.PosterConfig) (ratingModel.RatingService, error) {
	// Get base components
	tmdbService, err := f.baseFactory.BuildTMDBComponents()
	if err != nil {
		return ratingModel.RatingService{}, err
	}

	// Add logo service
	logoService := logoTmdb.NewTMDBLogoService(f.logger, defaultPosterConfig, logoCreator)
	tmdbService.LogoService = logoService

	return tmdbService, nil
}

func (f *RatingPlatformServiceModelFactory) buildRottenTomatoesRatingService(logoCreator *logo.LogoCreator, defaultPosterConfig *model.PosterConfig) (ratingModel.RatingService, error) {
	// Get base components
	rottenTomatoesService, err := f.baseFactory.BuildRottenTomatoesComponents()
	if err != nil {
		return ratingModel.RatingService{}, err
	}

	// Add logo service
	logoService := logoRottenTomatoes.NewRottenTomatoesLogoService(f.logger, defaultPosterConfig, logoCreator)
	rottenTomatoesService.LogoService = logoService

	return rottenTomatoesService, nil
}

func (f *RatingPlatformServiceModelFactory) buildIMDBRatingService(logoCreator *logo.LogoCreator, defaultPosterConfig *model.PosterConfig) (ratingModel.RatingService, error) {
	// Get base components
	imdbService, err := f.baseFactory.BuildIMDBComponents()
	if err != nil {
		return ratingModel.RatingService{}, err
	}

	// Add logo service
	logoService := logoIMDB.NewIMDBLogoService(f.logger, defaultPosterConfig, logoCreator)
	imdbService.LogoService = logoService

	return imdbService, nil
}
