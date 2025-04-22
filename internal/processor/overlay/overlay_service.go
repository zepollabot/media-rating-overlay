package overlay

import (
	"image"

	"github.com/fogleman/gg"
	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type ImageService interface {
	OpenImage(filePath string) (image.Image, error)
	CreateContext(width, height int) *gg.Context
}

type OverlayFactory interface {
	CreateOverlay(overlayType string) (model.Overlay, error)
}

type OverlayService struct {
	logger         *zap.Logger
	imageService   ImageService
	overlayFactory OverlayFactory
	posterConfig   *model.PosterConfig
}

func NewOverlayService(logger *zap.Logger, imageService ImageService, overlayFactory OverlayFactory, posterConfig *model.PosterConfig) *OverlayService {
	return &OverlayService{
		logger:         logger,
		imageService:   imageService,
		overlayFactory: overlayFactory,
		posterConfig:   posterConfig,
	}
}

func (s *OverlayService) CreateDrawContextWithOverlay(filePath string, item model.Item, config *config.Library) (*gg.Context, error) {
	// Open the image
	img, err := s.imageService.OpenImage(filePath)
	if err != nil {
		s.logger.Debug("Unable to open image",
			zap.String("Item ID", item.ID),
			zap.String("File Path", filePath),
			zap.Error(err),
		)
		return nil, err
	}

	// Create drawing context
	drawContext := s.imageService.CreateContext(
		s.posterConfig.Dimensions.Width,
		s.posterConfig.Dimensions.Height,
	)

	// Create overlay
	overlay, err := s.overlayFactory.CreateOverlay(config.Overlay.Type)
	if err != nil {
		s.logger.Debug("Unable to create overlay",
			zap.String("Overlay Type", config.Overlay.Type),
			zap.String("Item ID", item.ID),
			zap.Error(err),
		)
		return nil, err
	}

	// Apply overlay
	overlay.Apply(img, drawContext, config)

	return drawContext, nil
}
