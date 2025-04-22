package text

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TextGuesserTestSuite struct {
	suite.Suite
	logger *zap.Logger
}

func TestTextGuesserSuite(t *testing.T) {
	suite.Run(t, new(TextGuesserTestSuite))
}

func (s *TextGuesserTestSuite) SetupTest() {
	s.logger = zap.NewNop()
}

func (s *TextGuesserTestSuite) TestNewTextGuesser() {
	// Arrange
	// Act
	guesser := NewTextGuesser(s.logger)

	// Assert
	s.NotNil(guesser)
	s.Equal(s.logger, guesser.logger)
}

func (s *TextGuesserTestSuite) TestEstimateTextPointByAvailableWidth() {
	tests := []struct {
		name           string
		availableWidth float64
		numberOfDigits int
		expected       float64
	}{
		{
			name:           "single digit",
			availableWidth: 100,
			numberOfDigits: 1,
			expected:       250.0, // 100/0.4
		},
		{
			name:           "two digits",
			availableWidth: 100,
			numberOfDigits: 2,
			expected:       101.01, // 100/0.99
		},
		{
			name:           "three digits",
			availableWidth: 100,
			numberOfDigits: 3,
			expected:       71.94, // 100/1.39
		},
		{
			name:           "more than three digits",
			availableWidth: 100,
			numberOfDigits: 4,
			expected:       55.87, // 100/1.79
		},
	}

	guesser := NewTextGuesser(s.logger)

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			// Act
			result := guesser.estimateTextPointByAvailableWidth(tt.availableWidth, tt.numberOfDigits)

			// Assert
			s.InDelta(tt.expected, result, 0.01)
		})
	}
}

func (s *TextGuesserTestSuite) TestFindTextMaxPoints() {
	tests := []struct {
		name           string
		imageWidth     float64
		areaWidth      float64
		areaHeight     float64
		fontPath       string
		numberOfDigits int
		expectError    bool
	}{
		{
			name:           "valid single digit",
			imageWidth:     180,
			areaWidth:      200,
			areaHeight:     100,
			fontPath:       "testdata/font.ttf",
			numberOfDigits: 1,
			expectError:    false,
		},
		{
			name:           "valid two digits",
			imageWidth:     180,
			areaWidth:      200,
			areaHeight:     100,
			fontPath:       "testdata/font.ttf",
			numberOfDigits: 2,
			expectError:    false,
		},
		{
			name:           "invalid font path",
			imageWidth:     180,
			areaWidth:      200,
			areaHeight:     50,
			fontPath:       "nonexistent/font.ttf",
			numberOfDigits: 1,
			expectError:    true,
		},
	}

	guesser := NewTextGuesser(s.logger)

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Arrange
			// Act
			fontHeight, textWidth, margin, err := guesser.FindTextMaxPoints(
				tt.imageWidth,
				tt.areaWidth,
				tt.areaHeight,
				tt.fontPath,
				tt.numberOfDigits,
			)

			// Assert
			if tt.expectError {
				s.Error(err)
				return
			}

			s.NoError(err)
			s.Greater(fontHeight, 0.0)
			s.Greater(textWidth, 0.0)
			s.Greater(margin, 0.0)
			s.Less(textWidth, tt.areaWidth-tt.imageWidth)
			s.Less(fontHeight, tt.areaHeight)
		})
	}
}
