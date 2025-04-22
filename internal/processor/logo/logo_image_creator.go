package logo

import (
	"image"
	"image/color"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"go.uber.org/zap"
)

// Constants for logo image processing
const (
	// LogoImageReduction is the initial reduction factor applied to the logo size
	LogoImageReduction = 0.85
	// LogoImageStepReduction is the factor used to iteratively reduce the logo size
	LogoImageStepReduction = 0.95
	// LogoImageMargin is the margin factor applied to the logo
	LogoImageMargin = 0.2
	// LogoAreaWidthFactor is the factor applied to the logo area width
	LogoAreaWidthFactor = 0.35
)

// ImageDimensions represents the dimensions of an image
type ImageDimensions struct {
	Width  int
	Height int
}

// Creator handles the creation and manipulation of logo images
type LogoImageCreator struct {
	logger      *zap.Logger
	visualDebug bool
}

// NewLogoImageCreator creates a new instance of the LogoImageCreator
func NewLogoImageCreator(logger *zap.Logger, visualDebug bool) *LogoImageCreator {
	return &LogoImageCreator{
		logger:      logger,
		visualDebug: visualDebug,
	}
}

// CreateContext creates a new context with the logo image properly sized and positioned
func (c *LogoImageCreator) CreateContext(
	logoAreaWidth float64,
	logoAreaHeight float64,
	logoPath string,
) (*gg.Context, error) {
	// Open and validate the logo image
	logoImage, err := c.openLogoImage(logoPath)
	if err != nil {
		return nil, err
	}

	// Calculate the target width for the logo area
	logoImageAreaWidth := logoAreaWidth * LogoAreaWidthFactor

	// Determine the appropriate resize strategy and resize the image
	logoImageResize := c.determineResizeStrategy(
		logoImage,
		logoImageAreaWidth,
		logoAreaHeight,
	)

	// Create the context with the resized image
	return c.createContextWithImage(
		logoImageResize,
		logoAreaHeight,
		c.visualDebug,
	)
}

// openLogoImage opens and validates the logo image file
func (c *LogoImageCreator) openLogoImage(logoPath string) (image.Image, error) {
	logoImage, err := imaging.Open(logoPath)
	if err != nil {
		c.logger.Error(
			"unable to open logo image",
			zap.String("filePath", logoPath),
			zap.Error(err),
		)
		return nil, err
	}
	return logoImage, nil
}

// determineResizeStrategy determines the appropriate resize strategy based on image dimensions
func (c *LogoImageCreator) determineResizeStrategy(
	logoImage image.Image,
	logoImageAreaWidth float64,
	logoAreaHeight float64,
) *image.NRGBA {
	logoImageBounds := logoImage.Bounds()
	logoImageDimensions := ImageDimensions{
		Width:  logoImageBounds.Max.X,
		Height: logoImageBounds.Max.Y,
	}

	// Determine if the logo area is horizontal or vertical
	isLogoAreaHorizontal := logoImageAreaWidth > logoAreaHeight

	if isLogoAreaHorizontal {
		c.logger.Debug("logo area is a horizontal rectangle")
		return c.resizeForHorizontalArea(logoImage, logoImageDimensions, logoImageAreaWidth, logoAreaHeight)
	} else {
		c.logger.Debug("logo area is a vertical rectangle or a square")
		return c.resizeForVerticalArea(logoImage, logoImageDimensions, logoImageAreaWidth, logoAreaHeight)
	}
}

// resizeForHorizontalArea handles resizing for a horizontal logo area
func (c *LogoImageCreator) resizeForHorizontalArea(
	logoImage image.Image,
	logoImageDimensions ImageDimensions,
	logoImageAreaWidth float64,
	logoAreaHeight float64,
) *image.NRGBA {
	// If logo image height > width, it's a vertical rectangle, resize by height
	if logoImageDimensions.Height > logoImageDimensions.Width {
		c.logger.Debug("image is a vertical rectangle or a square, resize by height")
		newImageHeight := logoAreaHeight * LogoImageReduction
		return c.resizeByHeight(newImageHeight, logoImage, logoImageAreaWidth)
	} else {
		// It's a horizontal rectangle, resize by width
		c.logger.Debug("image is a horizontal rectangle, resize by width")
		newImageWidth := logoImageAreaWidth * LogoImageReduction
		return c.resizeByWidth(newImageWidth, logoImage, logoAreaHeight)
	}
}

// resizeForVerticalArea handles resizing for a vertical logo area
func (c *LogoImageCreator) resizeForVerticalArea(
	logoImage image.Image,
	logoImageDimensions ImageDimensions,
	logoImageAreaWidth float64,
	logoAreaHeight float64,
) *image.NRGBA {
	// If logo image height <= width, it's a horizontal rectangle, resize by width
	if logoImageDimensions.Height <= logoImageDimensions.Width {
		c.logger.Debug("image is a horizontal rectangle or a square, resize by width")
		newImageWidth := logoImageAreaWidth * LogoImageReduction
		return c.resizeByWidth(newImageWidth, logoImage, logoAreaHeight)
	} else {
		// It's a vertical rectangle, resize by height
		c.logger.Debug("image is a vertical rectangle, resize by height")
		newImageHeight := logoAreaHeight * LogoImageReduction
		return c.resizeByHeight(newImageHeight, logoImage, logoImageAreaWidth)
	}
}

// resizeByHeight resizes the image by height while ensuring width constraints are met
func (c *LogoImageCreator) resizeByHeight(
	newImageHeight float64,
	logoImage image.Image,
	logoImageAreaWidth float64,
) *image.NRGBA {
	var logoImageResize *image.NRGBA
	for {
		logoImageResize = imaging.Resize(logoImage, 0, int(newImageHeight), imaging.Lanczos)
		tmpLogoImageResizeBounds := logoImageResize.Bounds()
		if float64(tmpLogoImageResizeBounds.Max.X) <= logoImageAreaWidth*(1-LogoImageMargin) {
			break
		}
		newImageHeight = newImageHeight * LogoImageStepReduction
	}
	return logoImageResize
}

// resizeByWidth resizes the image by width while ensuring height constraints are met
func (c *LogoImageCreator) resizeByWidth(
	newImageWidth float64,
	logoImage image.Image,
	logoAreaHeight float64,
) *image.NRGBA {
	var logoImageResize *image.NRGBA
	for {
		logoImageResize = imaging.Resize(logoImage, int(newImageWidth), 0, imaging.Lanczos)
		tmpLogoImageResizeBounds := logoImageResize.Bounds()
		if float64(tmpLogoImageResizeBounds.Max.Y) <= logoAreaHeight*(1-LogoImageMargin) {
			break
		}
		newImageWidth = newImageWidth * LogoImageStepReduction
	}
	return logoImageResize
}

// createContextWithImage creates a context with the resized image properly positioned
func (c *LogoImageCreator) createContextWithImage(
	logoImageResize *image.NRGBA,
	logoAreaHeight float64,
	visualDebug bool,
) (*gg.Context, error) {
	logoImageResizeBounds := logoImageResize.Bounds()
	c.logger.Debug("new image size",
		zap.Int("width", logoImageResizeBounds.Max.X),
		zap.Int("height", logoImageResizeBounds.Max.Y),
	)

	// Calculate margins for centering
	marginTop := (logoAreaHeight - float64(logoImageResizeBounds.Max.Y)) / 2
	marginLeft := float64(logoImageResizeBounds.Max.X) * LogoImageMargin / 2

	// Create the context with appropriate dimensions
	contextWidth := int(float64(logoImageResizeBounds.Max.X) + marginLeft*2)
	imageContext := gg.NewContext(contextWidth, int(logoAreaHeight))

	// Add debug visualization if enabled
	if visualDebug {
		c.addDebugVisualization(imageContext, contextWidth, logoAreaHeight)
	}

	// Draw the image in the context
	imageContext.DrawImage(logoImageResize, int(marginLeft), int(marginTop))

	// Save debug image if enabled
	if visualDebug {
		if err := c.saveDebugImage(imageContext); err != nil {
			return nil, err
		}
	}

	return imageContext, nil
}

// addDebugVisualization adds visual debugging elements to the context
func (c *LogoImageCreator) addDebugVisualization(
	imageContext *gg.Context,
	contextWidth int,
	logoAreaHeight float64,
) {
	imageContext.SetColor(color.RGBA{R: 100, G: 150, B: 200, A: 0xFF})
	imageContext.DrawRectangle(0, 0, float64(contextWidth), logoAreaHeight)
	imageContext.Fill()
}

// saveDebugImage saves the debug image to disk
func (c *LogoImageCreator) saveDebugImage(imageContext *gg.Context) error {
	debugPath := filepath.Join("internal", "app", "images", "image.png")
	if err := imageContext.SavePNG(debugPath); err != nil {
		c.logger.Error("failed to save debug image", zap.Error(err))
		return err
	}
	return nil
}
