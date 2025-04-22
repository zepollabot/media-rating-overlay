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

// FrameOverlay implements the Overlay interface for frame overlays
type FrameOverlay struct {
	logger *zap.Logger
	config *model.PosterConfig
}

// NewFrameOverlay creates a new frame overlay
func NewFrameOverlay(logger *zap.Logger, config *model.PosterConfig) *FrameOverlay {
	return &FrameOverlay{
		logger: logger,
		config: config,
	}
}

// Apply applies the frame overlay to the image
func (o *FrameOverlay) Apply(
	img image.Image,
	dc *gg.Context,
	config *config.Library,
) {
	// set a white background
	//dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 0xFF})
	// set a black background
	dc.SetColor(color.RGBA{A: 0xFF})
	dc.DrawRectangle(0, 0, float64(o.config.Dimensions.Width), float64(o.config.Dimensions.Height))
	dc.Fill()

	diffWidth := float64(o.config.Dimensions.Width) * config.Overlay.Height
	resizeWidth := float64(o.config.Dimensions.Width) - diffWidth

	// Resize srcImage to width = resizeWidth preserving the aspect ratio.
	imageResized := imaging.Resize(img, int(resizeWidth), 0, imaging.Lanczos)

	dc.DrawImage(imageResized, int(diffWidth/2), 0)
}
