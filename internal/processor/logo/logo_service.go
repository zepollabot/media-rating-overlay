package logo

import (
	"errors"
	"image/color"

	"github.com/fogleman/gg"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	text "github.com/zepollabot/media-rating-overlay/internal/processor/text"
)

type TextCreatorInterface interface {
	CreateContext(
		areaWidth float64,
		areaHeight float64,
		horizontalMargin float64,
		fontHeightInPoints float64,
		text string,
		fontPath string,
	) (*gg.Context, error)
}

// LogoService implements the LogoService interface
type LogoService struct {
	logger      *zap.Logger
	textCreator TextCreatorInterface
	config      *model.PosterConfig
}

// NewLogoService creates a new logo service
func NewLogoService(
	logger *zap.Logger,
	textCreator TextCreatorInterface,
	config *model.PosterConfig,
) *LogoService {
	return &LogoService{
		logger:      logger,
		textCreator: textCreator,
		config:      config,
	}
}

// PositionLogos positions logos in the given area
func (s *LogoService) PositionLogos(
	logos []*model.Logo,
	areaWidth,
	areaHeight float64,
	visualDebug bool,
) (*gg.Context, error) {
	if len(logos) == 0 {
		return nil, &model.PosterError{
			Stage: "position_logos",
			Err:   errors.New("no logos to position"),
		}
	}

	// Calculate standard font size and margin
	standardFontSize := s.chooseFontSize(logos)
	standardTextHorizontalMargin := s.chooseTextHorizontalMargin(logos)

	// Apply standard font size and margin to all logos
	var logosSumWidth int
	for _, logo := range logos {
		logo.Text.Points = standardFontSize
		logo.Text.HorizontalMargin = standardTextHorizontalMargin

		textContext, err := s.textCreator.CreateContext(areaWidth, areaHeight, standardTextHorizontalMargin, standardFontSize, logo.Text.Value, text.FontPath)
		if err != nil {
			return nil, &model.PosterError{
				Stage: "get_text_context",
				Err:   err,
			}
		}

		logo.Text.Context = textContext
		logo.SumWidth = logo.Image.Context.Width() + textContext.Width()
		logosSumWidth += logo.SumWidth
	}

	// Create context for logo area
	logoAreaContext := gg.NewContext(int(areaWidth), int(areaHeight))

	// Debug visualization if enabled
	if visualDebug {
		logoAreaContext.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 0xFF})
		logoAreaContext.DrawRectangle(0, 0, areaWidth, areaHeight)
		logoAreaContext.Fill()
	}

	// Calculate centering margin
	centeringMargin := (int(areaWidth) - logosSumWidth) / (len(logos) + 1)
	var startMargin int

	// Position each logo
	for _, logo := range logos {
		marginToApplyLeftToText := centeringMargin + logo.Image.Context.Width()

		singleLogoAreaContext := gg.NewContext(logo.SumWidth+centeringMargin, int(areaHeight))
		singleLogoAreaContext.DrawImage(logo.Image.Context.Image(), centeringMargin, 0)
		singleLogoAreaContext.DrawImage(logo.Text.Context.Image(), marginToApplyLeftToText, 0)

		logoAreaContext.DrawImage(
			singleLogoAreaContext.Image(),
			startMargin,
			0,
		)

		startMargin += singleLogoAreaContext.Width()
	}

	return logoAreaContext, nil
}

// chooseFontSize chooses the smallest font size among logos
func (s *LogoService) chooseFontSize(logos []*model.Logo) float64 {
	fontSize := 0.0
	for _, logo := range logos {
		if fontSize == 0.0 {
			fontSize = logo.Text.Points
		} else if logo.Text.Points < fontSize {
			fontSize = logo.Text.Points
		}
	}
	return fontSize
}

// chooseTextHorizontalMargin chooses the smallest horizontal margin among logos
func (s *LogoService) chooseTextHorizontalMargin(logos []*model.Logo) float64 {
	horizontalMargin := 0.0
	for _, logo := range logos {
		if horizontalMargin == 0.0 {
			horizontalMargin = logo.Text.HorizontalMargin
		} else if logo.Text.HorizontalMargin < horizontalMargin {
			horizontalMargin = logo.Text.HorizontalMargin
		}
	}
	return horizontalMargin
}
