package item

import (
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

// ItemEligibilityService implements ItemEligibilityChecker interface
type ItemEligibilityService struct {
	ratingPlatformServices []ratingModel.RatingService
	logger                 *zap.Logger
}

func NewItemEligibilityService(ratingPlatformServices []ratingModel.RatingService, logger *zap.Logger) *ItemEligibilityService {
	return &ItemEligibilityService{
		ratingPlatformServices: ratingPlatformServices,
		logger:                 logger,
	}
}

func (s *ItemEligibilityService) IsEligible(item *model.Item) bool {
	s.logger.Debug("Checking if item is eligible..",
		zap.String("Item ID", item.ID),
	)
	if !item.IsEligible {
		s.logger.Debug("Item is not eligible, skipping",
			zap.String("Item ID", item.ID),
			zap.String("Item Title", item.Title),
		)
		return false
	}

	if len(item.Ratings) == 0 && len(s.ratingPlatformServices) == 0 {
		s.logger.Debug("Item doens't have any ratings, skipping",
			zap.String("Item ID", item.ID),
			zap.String("Item Title", item.Title),
		)
		return false
	}

	return true
}
