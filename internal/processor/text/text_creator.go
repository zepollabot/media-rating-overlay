package text

import (
	"image/color"
	"path/filepath"

	"github.com/fogleman/gg"
	"go.uber.org/zap"
)

type Creator struct {
	logger      *zap.Logger
	visualDebug bool
}

func NewTextCreator(logger *zap.Logger, visualDebug bool) *Creator {
	return &Creator{
		logger:      logger,
		visualDebug: visualDebug,
	}
}

func (c *Creator) CreateContext(
	areaWidth float64,
	areaHeight float64,
	horizontalMargin float64,
	fontHeightInPoints float64,
	text string,
	fontPath string,
) (*gg.Context, error) {

	textAreaWidth := areaWidth + horizontalMargin*2
	context, textWidth, err := c.PrepareContext(textAreaWidth, areaHeight, horizontalMargin, fontPath, fontHeightInPoints, text)
	if err != nil {
		return context, err
	}

	if textWidth < areaWidth {
		newTextWidth := textWidth + horizontalMargin*2
		context, _, err = c.PrepareContext(newTextWidth, areaHeight, horizontalMargin, fontPath, fontHeightInPoints, text)
		if err != nil {
			return context, err
		}
	}

	return context, nil
}

func (c *Creator) PrepareContext(
	contextWidth float64,
	contextHeight float64,
	horizontalMargin float64,
	fontPath string,
	fontHeightInPoints float64,
	text string,
) (*gg.Context, float64, error) {
	textShadowColor := color.Black
	textColor := color.White

	textContext := gg.NewContext(int(contextWidth), int(contextHeight))

	if c.visualDebug {
		textContext.SetColor(color.RGBA{R: 50, G: 150, B: 50, A: 0xFF})
		textContext.DrawRectangle(0, 0, contextWidth, contextHeight)
		textContext.Fill()
	}

	if errLoad := textContext.LoadFontFace(fontPath, fontHeightInPoints); errLoad != nil {
		c.logger.Error(
			"unable to load font",
			zap.Error(errLoad),
		)
		return nil, 0.0, errLoad
	}

	textWidth, textHeight := textContext.MeasureString(text)

	x := horizontalMargin
	y := textHeight + (float64(textContext.Height())-textHeight)/2

	textContext.SetColor(textShadowColor)
	textContext.DrawString(text, x+1, y+1)
	textContext.SetColor(textColor)
	textContext.DrawString(text, x, y)

	if c.visualDebug {
		if err := textContext.SavePNG(filepath.Join("internal", "app", "images", "text.png")); err != nil {
			panic("fail")
		}
	}

	return textContext, textWidth, nil
}
