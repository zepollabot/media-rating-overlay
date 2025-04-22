package text

import (
	"github.com/fogleman/gg"
	"go.uber.org/zap"
)

const textStepReduction = 0.95

const textAreaReduction = 0.85

const textMargin = 0.1

type Guesser struct {
	logger *zap.Logger
}

func NewTextGuesser(logger *zap.Logger) *Guesser {
	return &Guesser{
		logger: logger,
	}
}

func (g *Guesser) FindTextMaxPoints(
	imageWidth float64,
	areaWidth float64,
	areaHeight float64,
	fontPath string,
	numberOfDigits int,
) (float64, float64, float64, error) {
	horizontalMargin := (areaWidth - imageWidth) * textMargin / 2
	areaAvailableForTextWidth := areaWidth - imageWidth - (horizontalMargin * 2)
	estimatedTextPoints := g.estimateTextPointByAvailableWidth(areaAvailableForTextWidth, numberOfDigits)

	fontHeightInPoints, textWidth, err := g.calculateTextMaxPoints(
		areaHeight,
		estimatedTextPoints,
		areaAvailableForTextWidth,
		fontPath,
		numberOfDigits,
	)

	return fontHeightInPoints, textWidth, horizontalMargin, err
}

func (g *Guesser) calculateTextMaxPoints(
	areaHeight float64,
	fontHeightInPoints float64,
	areaAvailableForTextWidth float64,
	fontPath string,
	numberOfDigits int,
) (float64, float64, error) {
	textContext := gg.NewContext(int(areaAvailableForTextWidth), int(areaHeight))
	logoTextAreaWidth := textContext.Width()

	var textWidth float64

	for {
		if errLoad := textContext.LoadFontFace(fontPath, fontHeightInPoints); errLoad != nil {
			return 0.0, 0.0, errLoad
		}

		var text string
		switch numberOfDigits {
		case 1:
			text = "1"
		case 2:
			text = "9,9"
		case 3:
			text = "99%"
		default:
			text = "100%"
		}

		tmpTextWidth, tmpTextHeight := textContext.MeasureString(text) // max text available

		if (tmpTextWidth <= float64(logoTextAreaWidth)) && (tmpTextHeight < (areaHeight * textAreaReduction)) {
			textWidth = tmpTextWidth
			break
		} else {
			// lower fontHeightInPoints by 2.5%
			fontHeightInPoints = fontHeightInPoints * textStepReduction
		}
	}

	g.logger.Debug("Data for text",
		zap.Float64("font height in points", fontHeightInPoints),
		zap.Float64("areaAvailableForTextWidth", areaAvailableForTextWidth),
		zap.Float64("textWidth", textWidth),
	)

	return fontHeightInPoints, textWidth, nil
}

func (g *Guesser) estimateTextPointByAvailableWidth(availableWidth float64, numberOfDigits int) float64 {
	// this is valid for the font BebasNeue-Regular
	var ratio float64
	switch numberOfDigits {
	case 1:
		ratio = 0.4
	case 2:
		ratio = 0.99
	case 3:
		ratio = 1.39
	default: // more than 4 is not supported
		g.logger.Debug("more than 4 digits is not supported", zap.Int("numberOfDigits", numberOfDigits))
		ratio = 1.79
	}

	return availableWidth / ratio
}
