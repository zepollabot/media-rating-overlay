package tmdb

import (
	"context"

	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/constant"
	"github.com/zepollabot/media-rating-overlay/internal/model"
	rating "github.com/zepollabot/media-rating-overlay/internal/rating-service"
)

type TMDBRatingPlatformService struct {
	logger        *zap.Logger
	searchService rating.SearchService
}

func NewTMDBRatingPlatformService(logger *zap.Logger, searchService rating.SearchService) *TMDBRatingPlatformService {
	return &TMDBRatingPlatformService{
		logger:        logger,
		searchService: searchService,
	}
}

func (s *TMDBRatingPlatformService) GetRating(ctx context.Context, item model.Item) (model.Rating, error) {
	s.logger.Debug("Retrieving TMDB rating..",
		zap.String("Item ID", item.ID),
	)
	results, err := s.searchService.GetResults(ctx, item)
	if err != nil {
		s.logger.Error("unable to get TMDB results",
			zap.String("Item ID", item.ID),
			zap.Error(err),
		)
		return model.Rating{}, err
	}

	if len(results) == 0 {
		s.logger.Debug("no results found",
			zap.String("Item ID", item.ID),
		)
		return model.Rating{}, nil
	} else {
		if len(results) > 1 {
			s.logger.Debug("multiple results found",
				zap.String("Item ID", item.ID),
				zap.Any("results", results),
			)
		}

		// even if there are multiple results, we take the first one as the best one
		result := results[0]
		if result.Vote > 0 {
			s.logger.Debug("TMDB rating found",
				zap.String("Item ID", item.ID),
				zap.Float32("Rating", float32(result.Vote)),
			)
			return model.Rating{
				Name:   constant.RatingServiceTMDB,
				Rating: float32(result.Vote),
				Type:   model.RatingServiceTypeAudience,
			}, nil
		}
	}

	return model.Rating{}, nil
}
