package poster

import (
	"context"
	"image"

	"github.com/fogleman/gg"
	"github.com/samber/lo"
	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

type OverlayService interface {
	CreateDrawContextWithOverlay(filePath string, item model.Item, config *config.Library) (*gg.Context, error)
}

type LogoService interface {
	PositionLogos(logos []*model.Logo, areaWidth float64, areaHeight float64, visualDebug bool) (*gg.Context, error)
}

type ImageProcessor interface {
	OpenImage(filePath string) (image.Image, error)
	ResizeImage(img image.Image, width int, height int) image.Image
	SaveImage(img image.Image, filePath string) (string, error)
	CreateContext(width, height int) *gg.Context
}

// PostersModifier is the main orchestrator for poster modification
type PosterGenerator struct {
	logger                 *zap.Logger
	imageProcessor         ImageProcessor
	logoService            LogoService
	overlayService         OverlayService
	posterConfig           *model.PosterConfig
	ratingPlatformServices []ratingModel.RatingService
	visualDebug            bool
}

// NewPostersModifier creates a new poster modifier
func NewPosterGenerator(
	logger *zap.Logger,
	imageProcessor ImageProcessor,
	logoService LogoService,
	overlayService OverlayService,
	posterConfig *model.PosterConfig,
	ratingPlatformServices []ratingModel.RatingService,
	visualDebug bool,
) *PosterGenerator {
	return &PosterGenerator{
		logger:                 logger,
		imageProcessor:         imageProcessor,
		logoService:            logoService,
		overlayService:         overlayService,
		posterConfig:           posterConfig,
		ratingPlatformServices: ratingPlatformServices,
		visualDebug:            visualDebug,
	}
}

// ApplyLogos applies logos to a poster
func (m *PosterGenerator) ApplyLogos(
	ctx context.Context,
	filePath string,
	config *config.Library,
	item model.Item,
) (string, error) {

	// Calculate number of logos
	estimatedNumberOfLogos := len(item.Ratings)
	m.logger.Debug("estimated number of logos to be applied",
		zap.Int("estimatedNumberOfLogos", estimatedNumberOfLogos),
	)

	drawContext, err := m.overlayService.CreateDrawContextWithOverlay(filePath, item, config)
	if err != nil {
		m.logger.Debug("Unable to create draw context with overlay",
			zap.String("Item ID", item.ID),
			zap.Error(err),
		)
		return filePath, err
	}
	// Calculate logo area dimensions
	logoAreaHeight := float64(m.posterConfig.Dimensions.Height) * config.Overlay.Height
	logoAreaWidth := float64(m.posterConfig.Dimensions.Width - m.posterConfig.Margins.Left*2)

	m.logger.Debug("logo area",
		zap.Float64("logoAreaHeight", logoAreaHeight),
		zap.Float64("logoAreaWidth", logoAreaWidth),
	)

	// Calculate dimensions for each logo
	singleLogoAreaWidth := logoAreaWidth / float64(estimatedNumberOfLogos)
	logoDimensions := model.LogoDimensions{
		AreaWidth:  singleLogoAreaWidth,
		AreaHeight: logoAreaHeight,
	}

	m.logger.Debug("logo dimensions",
		zap.Any("logoDimensions", logoDimensions),
	)

	logos := m.buildLogos(ctx, item, logoDimensions)

	// Position logos
	logoAreaContext, err := m.logoService.PositionLogos(logos, logoAreaWidth, logoAreaHeight, m.visualDebug)
	if err != nil {
		m.logger.Debug("Unable to position logos",
			zap.String("Item ID", item.ID),
			zap.Error(err),
		)
		return filePath, err
	}

	// Draw logos on the poster
	drawContext.DrawImage(
		logoAreaContext.Image(),
		m.posterConfig.Margins.Left,
		int(float64(m.posterConfig.Dimensions.Height)-logoAreaHeight),
	)

	// Save the poster
	posterFilePath, err := m.imageProcessor.SaveImage(drawContext.Image(), filePath)
	if err != nil {
		m.logger.Debug("Unable to save poster",
			zap.String("Item ID", item.ID),
			zap.String("File Path", filePath),
			zap.Error(err),
		)
		return filePath, err
	}

	return posterFilePath, nil
}

func (m *PosterGenerator) buildLogos(ctx context.Context, item model.Item, logoDimensions model.LogoDimensions) []*model.Logo {
	logos := make([]*model.Logo, 0)

	for _, rating := range item.Ratings {
		ratingService, ok := lo.Find(m.ratingPlatformServices, func(rs ratingModel.RatingService) bool {
			return rs.Name == rating.Name
		})

		if !ok {
			m.logger.Debug("Rating service not found",
				zap.String("Item ID", item.ID),
				zap.String("Rating Name", rating.Name),
			)
			// TODO: do we want to stop the execution if a rating service is not found?
			continue
		}

		serviceLogos, err := ratingService.LogoService.GetLogos(ctx, []model.Rating{rating}, rating.Name, logoDimensions)
		if err != nil {
			m.logger.Debug("Unable to get logos from rating service",
				zap.String("Item ID", item.ID),
				zap.Error(err),
			)
			// TODO: do we want to stop the execution if we can't get logos from a rating service?
			continue
		}

		logos = append(logos, serviceLogos...)
	}

	return logos
}
