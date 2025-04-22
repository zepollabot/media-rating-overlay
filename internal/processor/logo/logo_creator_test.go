package logo

import (
	"errors"
	"testing"

	"github.com/fogleman/gg"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	logo_mocks "github.com/zepollabot/media-rating-overlay/internal/processor/logo/mocks"
)

type LogoCreatorSuite struct {
	suite.Suite
	mockLogoImageCreator *logo_mocks.LogoImageCreatorInterface
	mockTextService      *logo_mocks.TextService
	logger               *zap.Logger
	config               *model.PosterConfig
	logoCreator          *LogoCreator
}

func (s *LogoCreatorSuite) SetupTest() {
	s.mockLogoImageCreator = new(logo_mocks.LogoImageCreatorInterface)
	s.mockTextService = new(logo_mocks.TextService)
	s.logger = zap.NewNop()          // Use Noop logger for tests
	s.config = &model.PosterConfig{} // Initialize with a default or specific config if needed

	s.logoCreator = NewLogoCreator(
		s.logger,
		s.mockLogoImageCreator,
		s.mockTextService,
		s.config,
	)
}

func (s *LogoCreatorSuite) TearDownTest() {
	s.mockLogoImageCreator.AssertExpectations(s.T())
	s.mockTextService.AssertExpectations(s.T())
}

func TestLogoCreatorTestSuite(t *testing.T) {
	suite.Run(t, new(LogoCreatorSuite))
}

func (s *LogoCreatorSuite) TestCreateLogo_Success() {
	// Arrange
	imagePath := "path/to/image.png"
	textValue := "Test"
	dimensions := model.LogoDimensions{
		AreaWidth:  100.0,
		AreaHeight: 50.0,
	}
	mockContext := &gg.Context{}
	expectedImageObject := model.Image{Context: mockContext}
	expectedTextObject := model.Text{
		Context:          nil,
		Points:           12.0,
		Width:            30.0,
		Value:            textValue,
		HorizontalMargin: 5.0,
	}

	s.mockLogoImageCreator.On("CreateContext", dimensions.AreaWidth, dimensions.AreaHeight, imagePath).
		Return(mockContext, nil).Once()

	s.mockTextService.On(
		"GetText",
		expectedImageObject,
		dimensions.AreaWidth,
		dimensions.AreaHeight,
		textValue,
		len(textValue),
	).Return(expectedTextObject, nil).Once()

	// Act
	logo, err := s.logoCreator.CreateLogo(imagePath, textValue, dimensions)

	// Assert
	s.NoError(err)
	s.NotNil(logo)
	s.Equal(expectedImageObject, logo.Image)
	s.Equal(expectedTextObject, logo.Text)
}

func (s *LogoCreatorSuite) TestCreateLogo_ErrorOnCreateContext() {
	// Arrange
	imagePath := "path/to/image.png"
	text := "Test"
	dimensions := model.LogoDimensions{
		AreaWidth:  100.0,
		AreaHeight: 50.0,
	}
	expectedError := errors.New("failed to create context")

	s.mockLogoImageCreator.On("CreateContext", dimensions.AreaWidth, dimensions.AreaHeight, imagePath).
		Return(nil, expectedError).Once()

	// Act
	logo, err := s.logoCreator.CreateLogo(imagePath, text, dimensions)

	// Assert
	s.Error(err)
	s.Nil(logo)
	var posterErr *model.PosterError
	s.ErrorAs(err, &posterErr)
	s.Equal("create_logo_image_context", posterErr.Stage)
	s.ErrorIs(posterErr.Err, expectedError)
}

func (s *LogoCreatorSuite) TestCreateLogo_ErrorOnGetText() {
	// Arrange
	imagePath := "path/to/image.png"
	text := "Test"
	dimensions := model.LogoDimensions{
		AreaWidth:  100.0,
		AreaHeight: 50.0,
	}
	mockContext := &gg.Context{}
	expectedImageObject := model.Image{Context: mockContext}
	expectedError := errors.New("failed to get text")

	s.mockLogoImageCreator.On("CreateContext", dimensions.AreaWidth, dimensions.AreaHeight, imagePath).
		Return(mockContext, nil).Once()

	s.mockTextService.On(
		"GetText",
		expectedImageObject,
		dimensions.AreaWidth,
		dimensions.AreaHeight,
		text,
		len(text),
	).Return(model.Text{}, expectedError).Once()

	// Act
	logo, err := s.logoCreator.CreateLogo(imagePath, text, dimensions)

	// Assert
	s.Error(err)
	s.Nil(logo)
	var posterErr *model.PosterError
	s.ErrorAs(err, &posterErr)
	s.Equal("create_logo_text", posterErr.Stage)
	s.ErrorIs(posterErr.Err, expectedError)
}
