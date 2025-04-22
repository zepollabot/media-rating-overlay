package rating

import (
	"context"

	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	rating "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

// RatingBuilderService implements RatingBuilder interface
type RatingBuilderService struct {
	ratingPlatformServices []rating.RatingService
	logger                 *zap.Logger
}

func NewRatingBuilderService(ratingPlatformServices []rating.RatingService, logger *zap.Logger) *RatingBuilderService {
	return &RatingBuilderService{
		ratingPlatformServices: ratingPlatformServices,
		logger:                 logger,
	}
}

func (s *RatingBuilderService) BuildRatings(ctx context.Context, item *model.Item) error {
	s.logger.Debug("Building ratings..",
		zap.String("Item ID", item.ID),
	)

	for _, ratingService := range s.ratingPlatformServices {
		if lo.ContainsBy(item.Ratings, func(rating model.Rating) bool {
			return rating.Name == ratingService.Name
		}) {
			s.logger.Debug("Rating already exists",
				zap.String("Item ID", item.ID),
				zap.String("Rating Service", ratingService.Name),
			)
			continue
		} else if ratingService.PlatformService != nil {
			// Rating service is configured in the config file, so we need to get the rating from the service
			s.logger.Debug("Getting rating from service",
				zap.String("Item ID", item.ID),
				zap.String("Rating Service", ratingService.Name),
			)

			rating, err := ratingService.PlatformService.GetRating(ctx, *item)
			if err != nil {
				return err
			}
			item.Ratings = append(item.Ratings, rating)
		}
	}

	return nil
}
