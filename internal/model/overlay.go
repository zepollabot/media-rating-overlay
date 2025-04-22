package model

import (
	"image"

	"github.com/fogleman/gg"
	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
)

// Overlay represents a poster overlay
type Overlay interface {
	Apply(img image.Image, dc *gg.Context, config *config.Library)
}
