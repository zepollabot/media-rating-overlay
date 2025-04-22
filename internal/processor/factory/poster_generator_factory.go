package factory

import (
	"go.uber.org/zap"

	model "github.com/zepollabot/media-rating-overlay/internal/model"
	"github.com/zepollabot/media-rating-overlay/internal/processor/file"
	"github.com/zepollabot/media-rating-overlay/internal/processor/image"
	"github.com/zepollabot/media-rating-overlay/internal/processor/logo"
	"github.com/zepollabot/media-rating-overlay/internal/processor/overlay"
	poster "github.com/zepollabot/media-rating-overlay/internal/processor/poster"
	"github.com/zepollabot/media-rating-overlay/internal/processor/text"
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

type PosterGeneratorFactory struct {
	Logger                 *zap.Logger
	RatingPlatformServices []ratingModel.RatingService
	VisualDebug            bool
}

func NewPosterGeneratorFactory(logger *zap.Logger, ratingPlatformServices []ratingModel.RatingService, visualDebug bool) *PosterGeneratorFactory {
	return &PosterGeneratorFactory{Logger: logger, RatingPlatformServices: ratingPlatformServices, VisualDebug: visualDebug}
}

func (f *PosterGeneratorFactory) Create() *poster.PosterGenerator {

	defaultPosterConfig := model.PosterConfigWithDefaultValues()

	textCreator := text.NewTextCreator(f.Logger, f.VisualDebug)
	logoService := logo.NewLogoService(f.Logger, textCreator, defaultPosterConfig)
	fileManager := file.NewFileManager(f.Logger)
	imageService := image.NewImageService(f.Logger, defaultPosterConfig, fileManager)
	imageProcessor := image.NewImageService(f.Logger, defaultPosterConfig, fileManager)
	overlayFactory := NewOverlayFactory(f.Logger, defaultPosterConfig)
	overlayService := overlay.NewOverlayService(f.Logger, imageService, overlayFactory, defaultPosterConfig)

	posterGenerator := poster.NewPosterGenerator(f.Logger, imageProcessor, logoService, overlayService, defaultPosterConfig, f.RatingPlatformServices, f.VisualDebug)

	return posterGenerator
}
