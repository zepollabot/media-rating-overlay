package metacritic

import (
	"context"
	"strconv"

	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type LogoCreator interface {
	CreateLogo(imagePath string, text string, dimensions model.LogoDimensions) (*model.Logo, error)
}

// MetacriticLogoService implements the LogoService interface for Metacritic
type MetacriticLogoService struct {
	logger      *zap.Logger
	config      *model.PosterConfig
	logoCreator LogoCreator
}

// NewMetacriticLogoService creates a new Metacritic logo service
func NewMetacriticLogoService(
	logger *zap.Logger,
	config *model.PosterConfig,
	logoCreator LogoCreator,
) *MetacriticLogoService {
	return &MetacriticLogoService{
		logger:      logger,
		config:      config,
		logoCreator: logoCreator,
	}
}

// GetLogos gets logos for a Metacritic item
func (s *MetacriticLogoService) GetLogos(
	ctx context.Context,
	ratings []model.Rating,
	itemID string,
	dimensions model.LogoDimensions,
) ([]*model.Logo, error) {
	logos := make([]*model.Logo, 0)

	s.logger.Debug("Build Metacritic logos..",
		zap.String("Item ID", itemID),
	)

	for _, rating := range ratings {
		if rating.Rating > 0 {
			logo, err := s.logoCreator.CreateLogo(
				s.config.ImagePaths.Metacritic.Critic.Normal,
				strconv.Itoa(int(rating.Rating)),
				dimensions,
			)

			if err != nil {
				s.logger.Debug("Error creating Metacritic logo", zap.Error(err))
				return nil, err
			}

			logos = append(logos, logo)
		}
	}

	return logos, nil
}
