package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/fogleman/gg"
	"github.com/zepollabot/media-rating-overlay/internal/model"
	text_mocks "github.com/zepollabot/media-rating-overlay/internal/processor/text/mocks"
)

type TextServiceTestSuite struct {
	suite.Suite
	logger   *zap.Logger
	service  *TextService
	guesser  *text_mocks.TextGuesser
	fontPath string
}

func (suite *TextServiceTestSuite) SetupTest() {
	suite.logger = zaptest.NewLogger(suite.T())
	suite.guesser = text_mocks.NewTextGuesser(suite.T())
	suite.service = NewTextService(suite.logger, suite.guesser)
	suite.fontPath = "internal/processor/text/fonts/Bebas-Neue/BebasNeue-Regular.ttf"
}

func (suite *TextServiceTestSuite) TestGetText_Success() {
	// Arrange
	expectedPoints := 20.0
	expectedWidth := 100.0
	expectedMargin := 10.0
	imageObject := model.Image{
		Context: gg.NewContext(180, 100), // Create a context with width 100
	}
	areaWidth := 200.0
	areaHeight := 100.0
	text := "Test Text"
	numberOfDigits := 2

	suite.guesser.EXPECT().
		FindTextMaxPoints(
			float64(imageObject.Context.Width()),
			areaWidth,
			areaHeight,
			suite.fontPath,
			numberOfDigits,
		).
		Return(expectedPoints, expectedWidth, expectedMargin, nil)

	// Act
	result, err := suite.service.GetText(imageObject, areaWidth, areaHeight, text, numberOfDigits)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedPoints, result.Points)
	assert.Equal(suite.T(), expectedWidth, result.Width)
	assert.Equal(suite.T(), text, result.Value)
	assert.Equal(suite.T(), expectedMargin, result.HorizontalMargin)
}

func (suite *TextServiceTestSuite) TestGetText_Error() {
	// Arrange
	imageObject := model.Image{
		Context: gg.NewContext(180, 100), // Create a context with width 100
	}
	areaWidth := 200.0
	areaHeight := 100.0
	text := "Test Text"
	numberOfDigits := 2

	suite.guesser.EXPECT().
		FindTextMaxPoints(
			float64(imageObject.Context.Width()),
			areaWidth,
			areaHeight,
			suite.fontPath,
			numberOfDigits,
		).
		Return(0.0, 0.0, 0.0, assert.AnError)

	// Act
	result, err := suite.service.GetText(imageObject, areaWidth, areaHeight, text, numberOfDigits)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), model.Text{}, result)
}

func TestTextServiceSuite(t *testing.T) {
	suite.Run(t, new(TextServiceTestSuite))
}
