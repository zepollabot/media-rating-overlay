package imdb

import (
	"context"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type LogoCreator interface {
	CreateLogo(imagePath string, text string, dimensions model.LogoDimensions) (*model.Logo, error)
}

// IMDBLogoService implements the RatingService interface for IMDb
type IMDBLogoService struct {
	logger      *zap.Logger
	config      *model.PosterConfig
	logoCreator LogoCreator
}

// NewIMDBLogoService creates a new IMDb service
func NewIMDBLogoService(
	logger *zap.Logger,
	config *model.PosterConfig,
	logoCreator LogoCreator,
) *IMDBLogoService {
	return &IMDBLogoService{
		logger:      logger,
		config:      config,
		logoCreator: logoCreator,
	}
}

// GetLogos gets logos for an IMDb item
func (s *IMDBLogoService) GetLogos(
	ctx context.Context,
	ratings []model.Rating,
	itemID string,
	dimensions model.LogoDimensions,
) ([]*model.Logo, error) {
	logos := make([]*model.Logo, 0)

	s.logger.Debug("Build IMDb logo..",
		zap.String("Item ID", itemID),
	)

	for _, rating := range ratings {
		if rating.Rating > 0.0 {
			rating := decimal.NewFromFloat32(rating.Rating).Round(1).StringFixedBank(1)

			logo, err := s.logoCreator.CreateLogo(
				s.config.ImagePaths.IMDB.Audience.Normal,
				rating,
				dimensions,
			)

			if err != nil {
				s.logger.Debug("Error creating IMDb logo", zap.Error(err))
				return nil, err
			}

			logos = append(logos, logo)
		}
	}

	return logos, nil
}
