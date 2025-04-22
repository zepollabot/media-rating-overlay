package factory

import (
	"errors"

	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	overlay "github.com/zepollabot/media-rating-overlay/internal/processor/overlay"
)

// DefaultOverlayFactory creates overlay instances
type OverlayFactory struct {
	logger *zap.Logger
	config *model.PosterConfig
}

// NewOverlayFactory creates a new overlay factory
func NewOverlayFactory(logger *zap.Logger, config *model.PosterConfig) *OverlayFactory {
	return &OverlayFactory{
		logger: logger,
		config: config,
	}
}

// CreateOverlay creates an overlay for the given overlay type
func (f *OverlayFactory) CreateOverlay(overlayType string) (model.Overlay, error) {
	switch overlayType {
	case "frame":
		return overlay.NewFrameOverlay(f.logger, f.config), nil
	case "bar":
		return overlay.NewBarOverlay(f.logger, f.config), nil
	default:
		return nil, &model.PosterError{
			Stage: "create_overlay",
			Err:   errors.New("invalid overlay type: " + overlayType),
		}
	}
}
