package rotten_tomatoes

import (
	"context"

	"go.uber.org/zap"

	"github.com/shopspring/decimal"
	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type LogoCreator interface {
	CreateLogo(imagePath string, text string, dimensions model.LogoDimensions) (*model.Logo, error)
}

// RottenTomatoesLogoService implements the RatingService interface for Rotten Tomatoes
type RottenTomatoesLogoService struct {
	logger      *zap.Logger
	config      *model.PosterConfig
	logoCreator LogoCreator
}

// NewRottenTomatoesService creates a new Rotten Tomatoes service
func NewRottenTomatoesLogoService(
	logger *zap.Logger,
	config *model.PosterConfig,
	logoCreator LogoCreator,
) *RottenTomatoesLogoService {
	return &RottenTomatoesLogoService{
		logger:      logger,
		config:      config,
		logoCreator: logoCreator,
	}
}

// GetLogos gets logos for a Rotten Tomatoes item
func (s *RottenTomatoesLogoService) GetLogos(
	ctx context.Context,
	ratings []model.Rating,
	itemID string,
	dimensions model.LogoDimensions,
) ([]*model.Logo, error) {
	logos := make([]*model.Logo, 0)

	s.logger.Debug("Build Rotten Tomatoes logos..",
		zap.String("Item ID", itemID),
	)

	logoPath := ""

	for _, rating := range ratings {
		if rating.Rating > 0.0 {

			percentageRating := rating.Rating * 10
			ratingText := decimal.NewFromFloat32(percentageRating).Round(2).StringFixedBank(0) + "%"

			switch rating.Type {
			case model.RatingServiceTypeAudience:
				if percentageRating < 60 {
					logoPath = s.config.ImagePaths.RottenTomatoes.Audience.Low
				} else {
					logoPath = s.config.ImagePaths.RottenTomatoes.Audience.Normal
				}
			case model.RatingServiceTypeCritic:
				if percentageRating < 60 {
					logoPath = s.config.ImagePaths.RottenTomatoes.Critic.Low
				} else {
					logoPath = s.config.ImagePaths.RottenTomatoes.Critic.Normal
				}
			default:
				s.logger.Debug("Unknown rating type", zap.String("type", rating.Type))
				continue
			}

			logo, err := s.logoCreator.CreateLogo(
				logoPath,
				ratingText,
				dimensions,
			)
			if err != nil {
				s.logger.Debug("Error creating Rotten Tomatoes logo", zap.Error(err))
				return nil, err
			}
			logos = append(logos, logo)
		}
	}

	return logos, nil
}
