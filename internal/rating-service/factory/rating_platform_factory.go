// internal/rating-service/factory/base_factory.go
package factory

import (
	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/constant"
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
	clientTmdb "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/client"
	filterTmdb "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/filter"
	tmdb "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/search"
	serviceTmdb "github.com/zepollabot/media-rating-overlay/internal/rating-service/platform/tmdb/service"
)

type RatingServiceBaseFactory struct {
	Logger *zap.Logger
	Config *config.Config
}

func NewRatingServiceBaseFactory(logger *zap.Logger, config *config.Config) *RatingServiceBaseFactory {
	return &RatingServiceBaseFactory{
		Logger: logger,
		Config: config,
	}
}

// BuildTMDBComponents returns TMDB-specific components without the logo service
func (f *RatingServiceBaseFactory) BuildTMDBComponents() (ratingModel.RatingService, error) {
	f.Logger.Info("Building TMDB rating service components")

	tmdbService := ratingModel.RatingService{
		Name: constant.RatingServiceTMDB,
	}

	if f.Config.TMDB.Enabled {
		ratingClient, err := clientTmdb.NewTMDBClient(&f.Config.TMDB, &f.Config.HTTPClient, f.Config.Logger.LogFilePath, f.Logger)
		if err != nil {
			f.Logger.Error("error creating TMDB client", zap.Error(err))
			return ratingModel.RatingService{}, err
		}
		filtersService := filterTmdb.NewTMDBFilterService(f.Logger)
		searchService := tmdb.NewTMDBSearchService(ratingClient, filtersService, f.Logger)
		ratingPlatformService := serviceTmdb.NewTMDBRatingPlatformService(f.Logger, searchService)

		f.Logger.Info("TMDB rating service initialized")
		tmdbService.PlatformService = ratingPlatformService
	}

	return tmdbService, nil
}

func (f *RatingServiceBaseFactory) BuildRottenTomatoesComponents() (ratingModel.RatingService, error) {
	f.Logger.Info("Building Rotten Tomatoes rating service")

	rottenTomatoesService := ratingModel.RatingService{
		Name: constant.RatingServiceRottenTomatoes,
	}

	// Rotten Tomatoes Platform Service is not yet implemented
	f.Logger.Info("Rotten Tomatoes rating service initialized")

	return rottenTomatoesService, nil
}

func (f *RatingServiceBaseFactory) BuildIMDBComponents() (ratingModel.RatingService, error) {
	f.Logger.Info("Building IMDB rating service")

	imdbService := ratingModel.RatingService{
		Name: constant.RatingServiceIMDB,
	}

	// IMDB Platform Service is not yet implemented
	f.Logger.Info("IMDB rating service initialized")

	return imdbService, nil
}
