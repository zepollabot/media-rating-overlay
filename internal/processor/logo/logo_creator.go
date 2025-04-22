package logo

import (
	"github.com/fogleman/gg"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
)

type LogoImageCreatorInterface interface {
	CreateContext(areaWidth, areaHeight float64, path string) (*gg.Context, error)
}

type TextService interface {
	GetText(imageObject model.Image, areaWidth float64, areaHeight float64, text string, numberOfDigits int) (model.Text, error)
}

type LogoCreator struct {
	logger           *zap.Logger
	logoImageCreator LogoImageCreatorInterface
	textService      TextService
	config           *model.PosterConfig
}

func NewLogoCreator(
	logger *zap.Logger,
	logoImageCreator LogoImageCreatorInterface,
	textService TextService,
	config *model.PosterConfig,
) *LogoCreator {
	return &LogoCreator{
		logger:           logger,
		logoImageCreator: logoImageCreator,
		textService:      textService,
		config:           config,
	}
}

// CreateLogo creates a new logo with the given parameters
func (s *LogoCreator) CreateLogo(
	imagePath string,
	text string,
	dimensions model.LogoDimensions,
) (*model.Logo, error) {
	imageContext, err := s.logoImageCreator.CreateContext(
		dimensions.AreaWidth,
		dimensions.AreaHeight,
		imagePath,
	)

	if err != nil {
		return nil, &model.PosterError{
			Stage: "create_logo_image_context",
			Err:   err,
		}
	}

	imageObject := model.Image{
		Context: imageContext,
	}

	textObject, err := s.textService.GetText(
		imageObject,
		dimensions.AreaWidth,
		dimensions.AreaHeight,
		text,
		len(text),
	)
	if err != nil {
		return nil, &model.PosterError{
			Stage: "create_logo_text",
			Err:   err,
		}
	}

	return &model.Logo{
		Image: imageObject,
		Text:  textObject,
	}, nil
}
