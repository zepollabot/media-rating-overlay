package tmdb

import (
	"context"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type LogoCreator interface {
	CreateLogo(imagePath string, text string, dimensions model.LogoDimensions) (*model.Logo, error)
}

// TMDBLogoService implements the LogoService interface for TMDB
type TMDBLogoService struct {
	logger      *zap.Logger
	config      *model.PosterConfig
	logoCreator LogoCreator
}

// NewTMDBLogoService creates a new TMDB logo service
func NewTMDBLogoService(
	logger *zap.Logger,
	config *model.PosterConfig,
	logoCreator LogoCreator,
) *TMDBLogoService {
	return &TMDBLogoService{
		logger:      logger,
		config:      config,
		logoCreator: logoCreator,
	}
}

// GetLogos gets logos for a TMDB item
func (s *TMDBLogoService) GetLogos(
	ctx context.Context,
	ratings []model.Rating,
	itemID string,
	dimensions model.LogoDimensions,
) ([]*model.Logo, error) {
	logos := make([]*model.Logo, 0)

	s.logger.Debug("Build TMDB logos..",
		zap.String("Item ID", itemID),
	)

	for _, rating := range ratings {
		if rating.Rating > 0.0 {
			rating := decimal.NewFromFloat32(rating.Rating).Round(1).StringFixedBank(1)

			logo, err := s.logoCreator.CreateLogo(
				s.config.ImagePaths.TMDB.Audience.Normal,
				rating,
				dimensions,
			)

			if err != nil {
				s.logger.Debug("Error creating TMDB logo", zap.Error(err))
				return nil, err
			}

			logos = append(logos, logo)
		}
	}

	return logos, nil
}
