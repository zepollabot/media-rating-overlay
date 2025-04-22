package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type TextCreatorTestSuite struct {
	suite.Suite
	logger      *zap.Logger
	service     *Creator
	fontPath    string
	visualDebug bool
}

func (suite *TextCreatorTestSuite) SetupTest() {
	suite.logger = zaptest.NewLogger(suite.T())
	suite.visualDebug = false
	suite.service = NewTextCreator(suite.logger, suite.visualDebug)
	suite.fontPath = "testdata/font.ttf" // font is Bebas Neue
}

func (suite *TextCreatorTestSuite) TestCreateContext_Success() {

	testCases := []struct {
		name                  string
		areaWidth             float64
		areaHeight            float64
		horizontalMargin      float64
		fontHeightInPoints    float64
		text                  string
		expectedContextWidth  int
		expectedContextHeight int
	}{
		{
			name:                  "Recalculation needed",
			areaWidth:             100.0,
			areaHeight:            50.0,
			horizontalMargin:      10.0,
			fontHeightInPoints:    20.0,
			text:                  "Test Text",
			expectedContextWidth:  82,
			expectedContextHeight: 50,
		},
		{
			name:                  "No recalculation needed",
			areaWidth:             100.0,
			areaHeight:            50.0,
			horizontalMargin:      10.0,
			fontHeightInPoints:    40.0,
			text:                  "Test Text",
			expectedContextWidth:  120,
			expectedContextHeight: 50,
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			// Arrange
			areaWidth := testCase.areaWidth
			areaHeight := testCase.areaHeight
			horizontalMargin := testCase.horizontalMargin
			fontHeightInPoints := testCase.fontHeightInPoints
			text := testCase.text

			// Act
			context, err := suite.service.CreateContext(
				areaWidth,
				areaHeight,
				horizontalMargin,
				fontHeightInPoints,
				text,
				suite.fontPath,
			)

			// Assert
			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), context)
			assert.Equal(suite.T(), testCase.expectedContextWidth, context.Width())
			assert.Equal(suite.T(), testCase.expectedContextHeight, context.Height())
		})
	}
}

func (suite *TextCreatorTestSuite) TestCreateContext_Error() {
	// Arrange
	areaWidth := 100.0
	areaHeight := 50.0
	horizontalMargin := 10.0
	fontHeightInPoints := 20.0
	text := "Test Text"
	invalidFontPath := "nonexistent.ttf"

	// Act
	context, err := suite.service.CreateContext(
		areaWidth,
		areaHeight,
		horizontalMargin,
		fontHeightInPoints,
		text,
		invalidFontPath,
	)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), context)
}

func (suite *TextCreatorTestSuite) TestPrepareContext_Success() {
	// Arrange
	contextWidth := 100.0
	contextHeight := 50.0
	horizontalMargin := 10.0
	fontHeightInPoints := 20.0
	text := "Test Text"

	// Act
	context, textWidth, err := suite.service.PrepareContext(
		contextWidth,
		contextHeight,
		horizontalMargin,
		suite.fontPath,
		fontHeightInPoints,
		text,
	)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), context)
	assert.Greater(suite.T(), textWidth, 0.0)
	assert.Equal(suite.T(), int(contextWidth), context.Width())
	assert.Equal(suite.T(), int(contextHeight), context.Height())
}

func (suite *TextCreatorTestSuite) TestPrepareContext_Error() {
	// Arrange
	contextWidth := 100.0
	contextHeight := 50.0
	horizontalMargin := 10.0
	fontHeightInPoints := 20.0
	text := "Test Text"
	invalidFontPath := "nonexistent.ttf"

	// Act
	context, textWidth, err := suite.service.PrepareContext(
		contextWidth,
		contextHeight,
		horizontalMargin,
		invalidFontPath,
		fontHeightInPoints,
		text,
	)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), context)
	assert.Equal(suite.T(), 0.0, textWidth)
}

func TestTextCreatorSuite(t *testing.T) {
	suite.Run(t, new(TextCreatorTestSuite))
}
