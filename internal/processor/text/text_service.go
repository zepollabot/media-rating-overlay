package text

import (
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type TextGuesser interface {
	FindTextMaxPoints(
		imageWidth float64,
		areaWidth float64,
		areaHeight float64,
		fontPath string,
		numberOfDigits int,
	) (float64, float64, float64, error)
}

// TextService is a service that creates and guesses text
type TextService struct {
	logger   *zap.Logger
	guesser  TextGuesser
	fontPath string
}

func NewTextService(logger *zap.Logger, guesser TextGuesser) *TextService {
	return &TextService{
		guesser:  guesser,
		logger:   logger,
		fontPath: FontPath,
	}
}

func (h *TextService) GetText(
	imageObject model.Image,
	areaWidth float64,
	areaHeight float64,
	text string,
	numberOfDigits int,
) (model.Text, error) {

	fontHeightInPoints, textWidth, horizontalMargin, err := h.guesser.FindTextMaxPoints(
		float64(imageObject.Context.Width()),
		areaWidth,
		areaHeight,
		h.fontPath,
		numberOfDigits,
	)

	if err != nil {
		return model.Text{}, err
	}

	return model.Text{
		Points:           fontHeightInPoints,
		Width:            textWidth,
		Value:            text,
		HorizontalMargin: horizontalMargin,
	}, nil
}
