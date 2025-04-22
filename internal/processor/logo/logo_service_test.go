package logo

import (
	"errors"
	"image/color"
	"testing"

	"github.com/fogleman/gg"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	logo_mocks "github.com/zepollabot/media-rating-overlay/internal/processor/logo/mocks"
	text_processor "github.com/zepollabot/media-rating-overlay/internal/processor/text"
)

type LogoServiceSuite struct {
	suite.Suite
	mockTextCreator *logo_mocks.TextCreatorInterface
	logger          *zap.Logger
	config          *model.PosterConfig
	logoService     *LogoService
}

func (s *LogoServiceSuite) SetupTest() {
	s.mockTextCreator = new(logo_mocks.TextCreatorInterface)
	s.logger = zap.NewNop()
	s.config = &model.PosterConfig{} // Initialize with a default or specific config

	s.logoService = NewLogoService(
		s.logger,
		s.mockTextCreator,
		s.config,
	)
}

func (s *LogoServiceSuite) TearDownTest() {
	s.mockTextCreator.AssertExpectations(s.T())
}

func TestLogoServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LogoServiceSuite))
}

func (s *LogoServiceSuite) TestPositionLogos_Success() {
	// Arrange
	areaWidth := 600.0
	areaHeight := 100.0
	visualDebug := false

	mockImageContext1 := gg.NewContext(50, 50)
	mockImageContext1.SetColor(color.NRGBA{R: 255, G: 0, B: 0, A: 255}) // Red
	mockImageContext1.Clear()
	mockImageContext2 := gg.NewContext(60, 60)
	mockImageContext2.SetColor(color.NRGBA{R: 0, G: 255, B: 0, A: 255}) // Green
	mockImageContext2.Clear()

	mockTextContext1 := gg.NewContext(100, 50)
	mockTextContext1.SetColor(color.NRGBA{R: 0, G: 0, B: 255, A: 255}) // Blue
	mockTextContext1.Clear()
	mockTextContext2 := gg.NewContext(120, 60)
	mockTextContext2.SetColor(color.NRGBA{R: 255, G: 255, B: 0, A: 255}) // Yellow
	mockTextContext2.Clear()

	logos := []*model.Logo{
		{
			Image: model.Image{Context: mockImageContext1},
			Text: model.Text{
				Value:            "TMDB",
				Points:           12.0,
				HorizontalMargin: 5.0,
			},
		},
		{
			Image: model.Image{Context: mockImageContext2},
			Text: model.Text{
				Value:            "IMDB",
				Points:           10.0, // Different to test chooseFontSize
				HorizontalMargin: 4.0,  // Different to test chooseTextHorizontalMargin
			},
		},
	}

	// Expectations for chooseFontSize and chooseTextHorizontalMargin (smallest values)
	expectedStandardFontSize := 10.0
	expectedStandardTextHorizontalMargin := 4.0

	s.mockTextCreator.On("CreateContext", areaWidth, areaHeight, expectedStandardTextHorizontalMargin, expectedStandardFontSize, logos[0].Text.Value, text_processor.FontPath).
		Return(mockTextContext1, nil).Once()
	s.mockTextCreator.On("CreateContext", areaWidth, areaHeight, expectedStandardTextHorizontalMargin, expectedStandardFontSize, logos[1].Text.Value, text_processor.FontPath).
		Return(mockTextContext2, nil).Once()

	// Act
	resultContext, err := s.logoService.PositionLogos(logos, areaWidth, areaHeight, visualDebug)

	// Assert
	s.NoError(err)
	s.NotNil(resultContext)
	s.Equal(int(areaWidth), resultContext.Width())
	s.Equal(int(areaHeight), resultContext.Height())
	// Further assertions could involve checking pixel data if exact positioning is critical
	// or mocking gg.Context drawing methods if we want to verify those calls.
}

func (s *LogoServiceSuite) TestPositionLogos_Success_VisualDebug() {
	// Arrange
	areaWidth := 600.0
	areaHeight := 100.0
	visualDebug := true // Enable visual debug

	mockImageContext1 := gg.NewContext(50, 50)
	mockTextContext1 := gg.NewContext(100, 50)

	logos := []*model.Logo{
		{
			Image: model.Image{Context: mockImageContext1},
			Text: model.Text{
				Value:            "TMDB",
				Points:           12.0,
				HorizontalMargin: 5.0,
			},
		},
	}

	expectedStandardFontSize := 12.0
	expectedStandardTextHorizontalMargin := 5.0

	s.mockTextCreator.On("CreateContext", areaWidth, areaHeight, expectedStandardTextHorizontalMargin, expectedStandardFontSize, logos[0].Text.Value, text_processor.FontPath).
		Return(mockTextContext1, nil).Once()

	// Act
	resultContext, err := s.logoService.PositionLogos(logos, areaWidth, areaHeight, visualDebug)

	// Assert
	s.NoError(err)
	s.NotNil(resultContext)
	s.Equal(int(areaWidth), resultContext.Width())
	s.Equal(int(areaHeight), resultContext.Height())
	// Specific check for visual debug background (e.g., check a corner pixel)
	// This is a simplified check. A more robust check would involve examining more pixels or using an image comparison library.
	img := resultContext.Image()
	at := img.At(0, 0) // Check top-left corner
	expectedColor := color.RGBA{R: 255, G: 255, B: 255, A: 0xFF}
	r, g, b, a := at.RGBA()
	s.Equal(uint32(expectedColor.R)*0x101, r, "Red component does not match")
	s.Equal(uint32(expectedColor.G)*0x101, g, "Green component does not match")
	s.Equal(uint32(expectedColor.B)*0x101, b, "Blue component does not match")
	s.Equal(uint32(expectedColor.A)*0x101, a, "Alpha component does not match")
}

func (s *LogoServiceSuite) TestPositionLogos_NoLogos() {
	// Arrange
	logos := []*model.Logo{}
	areaWidth := 200.0
	areaHeight := 100.0
	visualDebug := false

	// Act
	resultContext, err := s.logoService.PositionLogos(logos, areaWidth, areaHeight, visualDebug)

	// Assert
	s.Error(err)
	s.Nil(resultContext)
	var posterErr *model.PosterError
	s.ErrorAs(err, &posterErr)
	s.Equal("position_logos", posterErr.Stage)
	s.EqualError(posterErr.Err, "no logos to position")
}

func (s *LogoServiceSuite) TestPositionLogos_ErrorOnTextCreator() {
	// Arrange
	areaWidth := 200.0
	areaHeight := 100.0
	visualDebug := false
	expectedError := errors.New("text creator failed")

	mockImageContext1 := gg.NewContext(50, 50)
	logos := []*model.Logo{
		{
			Image: model.Image{Context: mockImageContext1},
			Text: model.Text{
				Value:            "TMDB",
				Points:           12.0,
				HorizontalMargin: 5.0,
			},
		},
	}

	expectedStandardFontSize := 12.0
	expectedStandardTextHorizontalMargin := 5.0

	s.mockTextCreator.On("CreateContext", areaWidth, areaHeight, expectedStandardTextHorizontalMargin, expectedStandardFontSize, logos[0].Text.Value, text_processor.FontPath).
		Return(nil, expectedError).Once()

	// Act
	resultContext, err := s.logoService.PositionLogos(logos, areaWidth, areaHeight, visualDebug)

	// Assert
	s.Error(err)
	s.Nil(resultContext)
	var posterErr *model.PosterError
	s.ErrorAs(err, &posterErr)
	s.Equal("get_text_context", posterErr.Stage)
	s.ErrorIs(posterErr.Err, expectedError)
}
