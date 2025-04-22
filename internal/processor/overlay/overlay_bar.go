package overlay

import (
	"image"
	"image/color"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/model"
)

// BarOverlay implements the Overlay interface for bar overlays
type BarOverlay struct {
	logger *zap.Logger
	config *model.PosterConfig
}

// NewBarOverlay creates a new bar overlay
func NewBarOverlay(logger *zap.Logger, config *model.PosterConfig) *BarOverlay {
	return &BarOverlay{
		logger: logger,
		config: config,
	}
}

// Apply applies the bar overlay to the image
func (o *BarOverlay) Apply(
	img image.Image,
	dc *gg.Context,
	config *config.Library,
) {
	// Resize srcImage to width = POSTER_WIDTH preserving the aspect ratio.
	imageResized := imaging.Resize(img, o.config.Dimensions.Width, 0, imaging.Lanczos)
	dc.DrawImage(imageResized, 0, 0)

	barHeight := float64(o.config.Dimensions.Height) * config.Overlay.Height
	startHeight := float64(o.config.Dimensions.Height) - barHeight
	transparency := config.Overlay.Transparency * float64(255)

	x := 0.0
	y := startHeight
	w := float64(dc.Width())
	h := barHeight
	dc.SetColor(color.RGBA{A: uint8(transparency)})
	dc.DrawRectangle(x, y, w, h)
	dc.Fill()
}
