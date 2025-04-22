package logo

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// Test fixtures
const (
	testImageDir = "testdata"
	// Test image dimensions
	horizontalImageWidth  = 800
	horizontalImageHeight = 400
	verticalImageWidth    = 400
	verticalImageHeight   = 800
	squareImageSize       = 500
)

// LogoCreatorTestSuite is a test suite for the LogoCreator
type LogoCreatorTestSuite struct {
	suite.Suite
	creator *LogoImageCreator
	logger  *zap.Logger
}

// SetupTest sets up the test suite
func (s *LogoCreatorTestSuite) SetupTest() {
	s.logger = zap.NewNop()
	s.creator = NewLogoImageCreator(s.logger, false)
}

// TearDownTest cleans up after each test
func (s *LogoCreatorTestSuite) TearDownTest() {
	// Clean up test files
	err := os.RemoveAll(testImageDir)
	require.NoError(s.T(), err)
}

// createTestImage creates a test image with the specified dimensions
func createTestImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	// Fill with a gradient for visual testing
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / width),
				G: uint8(y * 255 / height),
				B: 100,
				A: 255,
			})
		}
	}
	return img
}

// saveTestImage saves a test image to disk
func (s *LogoCreatorTestSuite) saveTestImage(img image.Image, filename string) string {
	// Create test directory if it doesn't exist
	err := os.MkdirAll(testImageDir, 0755)
	require.NoError(s.T(), err)

	// Save the image
	path := filepath.Join(testImageDir, filename)
	err = imaging.Save(img, path)
	require.NoError(s.T(), err)
	return path
}

// TestNewLogoCreator tests the creation of a new LogoCreator
func (s *LogoCreatorTestSuite) TestNewLogoCreator() {
	creator := NewLogoImageCreator(s.logger, false)
	assert.NotNil(s.T(), creator)
	assert.Equal(s.T(), s.logger, creator.logger)
}

// TestCreateContext_HorizontalImage tests creating a context with a horizontal image
func (s *LogoCreatorTestSuite) TestCreateContext_HorizontalImage() {
	// Create and save a horizontal test image
	horizontalImg := createTestImage(horizontalImageWidth, horizontalImageHeight)
	horizontalImgPath := s.saveTestImage(horizontalImg, "horizontal.png")

	// Test parameters
	logoAreaWidth := 1000.0
	logoAreaHeight := 500.0

	// Execute
	context, err := s.creator.CreateContext(logoAreaWidth, logoAreaHeight, horizontalImgPath)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), context)

	// Verify context dimensions
	contextWidth := context.Width()
	contextHeight := context.Height()
	assert.Equal(s.T(), int(logoAreaHeight), contextHeight)
	assert.True(s.T(), contextWidth > 0)
}

// TestCreateContext_VerticalImage tests creating a context with a vertical image
func (s *LogoCreatorTestSuite) TestCreateContext_VerticalImage() {
	// Create and save a vertical test image
	verticalImg := createTestImage(verticalImageWidth, verticalImageHeight)
	verticalImgPath := s.saveTestImage(verticalImg, "vertical.png")

	// Test parameters
	logoAreaWidth := 500.0
	logoAreaHeight := 1000.0

	// Execute
	context, err := s.creator.CreateContext(logoAreaWidth, logoAreaHeight, verticalImgPath)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), context)

	// Verify context dimensions
	contextWidth := context.Width()
	contextHeight := context.Height()
	assert.Equal(s.T(), int(logoAreaHeight), contextHeight)
	assert.True(s.T(), contextWidth > 0)
}

// TestCreateContext_SquareImage tests creating a context with a square image
func (s *LogoCreatorTestSuite) TestCreateContext_SquareImage() {
	// Create and save a square test image
	squareImg := createTestImage(squareImageSize, squareImageSize)
	squareImgPath := s.saveTestImage(squareImg, "square.png")

	// Test parameters
	logoAreaWidth := 800.0
	logoAreaHeight := 800.0

	// Execute
	context, err := s.creator.CreateContext(logoAreaWidth, logoAreaHeight, squareImgPath)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), context)

	// Verify context dimensions
	contextWidth := context.Width()
	contextHeight := context.Height()
	assert.Equal(s.T(), int(logoAreaHeight), contextHeight)
	assert.True(s.T(), contextWidth > 0)
}

// TestCreateContext_InvalidPath tests creating a context with an invalid image path
func (s *LogoCreatorTestSuite) TestCreateContext_InvalidPath() {
	// Test parameters
	logoAreaWidth := 1000.0
	logoAreaHeight := 500.0
	invalidPath := "nonexistent.png"

	// Execute
	context, err := s.creator.CreateContext(logoAreaWidth, logoAreaHeight, invalidPath)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), context)
}

// TestCreateContext_WithVisualDebug tests creating a context with visual debug enabled
func (s *LogoCreatorTestSuite) TestCreateContext_WithVisualDebug() {
	// Create and save a test image
	testImg := createTestImage(horizontalImageWidth, horizontalImageHeight)
	testImgPath := s.saveTestImage(testImg, "debug.png")

	// Note: We can't modify the VisualDebug constant directly, so we'll just test
	// that the code doesn't panic when VisualDebug is false (default)

	// Test parameters
	logoAreaWidth := 1000.0
	logoAreaHeight := 500.0

	// Execute
	context, err := s.creator.CreateContext(logoAreaWidth, logoAreaHeight, testImgPath)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), context)

	// Verify context dimensions
	contextWidth := context.Width()
	contextHeight := context.Height()
	assert.Equal(s.T(), int(logoAreaHeight), contextHeight)
	assert.True(s.T(), contextWidth > 0)
}

// TestResizeByHeight tests the resizeByHeight method
func (s *LogoCreatorTestSuite) TestResizeByHeight() {
	// Create a test image
	testImg := createTestImage(verticalImageWidth, verticalImageHeight)
	newImageHeight := 400.0
	logoImageAreaWidth := 500.0

	// Execute
	resizedImg := s.creator.resizeByHeight(newImageHeight, testImg, logoImageAreaWidth)
	assert.NotNil(s.T(), resizedImg)

	// Verify dimensions
	bounds := resizedImg.Bounds()
	assert.Equal(s.T(), int(newImageHeight), bounds.Max.Y)
	assert.True(s.T(), bounds.Max.X <= int(logoImageAreaWidth*(1-LogoImageMargin)))
}

// TestResizeByWidth tests the resizeByWidth method
func (s *LogoCreatorTestSuite) TestResizeByWidth() {
	// Create a test image
	testImg := createTestImage(horizontalImageWidth, horizontalImageHeight)
	newImageWidth := 400.0
	logoAreaHeight := 500.0

	// Execute
	resizedImg := s.creator.resizeByWidth(newImageWidth, testImg, logoAreaHeight)
	assert.NotNil(s.T(), resizedImg)

	// Verify dimensions
	bounds := resizedImg.Bounds()
	assert.Equal(s.T(), int(newImageWidth), bounds.Max.X)
	assert.True(s.T(), bounds.Max.Y <= int(logoAreaHeight*(1-LogoImageMargin)))
}

// TestDetermineResizeStrategy tests the determineResizeStrategy method
func (s *LogoCreatorTestSuite) TestDetermineResizeStrategy() {
	// Test cases
	testCases := []struct {
		name               string
		imgWidth           int
		imgHeight          int
		logoImageAreaWidth float64
		logoAreaHeight     float64
		expectedStrategy   string // "width" or "height"
	}{
		{
			name:               "Horizontal area with vertical image",
			imgWidth:           verticalImageWidth,
			imgHeight:          verticalImageHeight,
			logoImageAreaWidth: 1000.0,
			logoAreaHeight:     500.0,
			expectedStrategy:   "height",
		},
		{
			name:               "Horizontal area with horizontal image",
			imgWidth:           horizontalImageWidth,
			imgHeight:          horizontalImageHeight,
			logoImageAreaWidth: 1000.0,
			logoAreaHeight:     500.0,
			expectedStrategy:   "width",
		},
		{
			name:               "Vertical area with horizontal image",
			imgWidth:           horizontalImageWidth,
			imgHeight:          horizontalImageHeight,
			logoImageAreaWidth: 500.0,
			logoAreaHeight:     1000.0,
			expectedStrategy:   "width",
		},
		{
			name:               "Vertical area with vertical image",
			imgWidth:           verticalImageWidth,
			imgHeight:          verticalImageHeight,
			logoImageAreaWidth: 500.0,
			logoAreaHeight:     1000.0,
			expectedStrategy:   "height",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create test image
			testImg := createTestImage(tc.imgWidth, tc.imgHeight)

			// Execute
			resizedImg := s.creator.determineResizeStrategy(
				testImg,
				tc.logoImageAreaWidth,
				tc.logoAreaHeight,
			)
			assert.NotNil(s.T(), resizedImg)

			// Verify strategy based on the actual implementation
			bounds := resizedImg.Bounds()

			// Check if the image was resized by width or height
			// For width-based resizing, the width should be approximately logoImageAreaWidth * LogoImageReduction
			// For height-based resizing, the height should be approximately logoAreaHeight * LogoImageReduction
			if tc.expectedStrategy == "width" {
				// Width-based resizing
				expectedWidth := int(tc.logoImageAreaWidth * LogoImageReduction)
				// Allow for variation due to the iterative reduction
				assert.InDelta(s.T(), float64(expectedWidth), float64(bounds.Max.X), 100.0,
					"Width should be approximately %v, got %v", expectedWidth, bounds.Max.X)
			} else {
				// Height-based resizing
				expectedHeight := int(tc.logoAreaHeight * LogoImageReduction)
				// Allow for variation due to the iterative reduction
				assert.InDelta(s.T(), float64(expectedHeight), float64(bounds.Max.Y), 100.0,
					"Height should be approximately %v, got %v", expectedHeight, bounds.Max.Y)
			}
		})
	}
}

// TestCreateContextWithImage tests the createContextWithImage method
func (s *LogoCreatorTestSuite) TestCreateContextWithImage() {
	// Create a test image
	testImg := createTestImage(horizontalImageWidth, horizontalImageHeight)
	logoAreaHeight := 500.0

	// Execute
	context, err := s.creator.createContextWithImage(testImg, logoAreaHeight, false)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), context)

	// Verify context dimensions
	contextWidth := context.Width()
	contextHeight := context.Height()
	assert.Equal(s.T(), int(logoAreaHeight), contextHeight)
	assert.True(s.T(), contextWidth > 0)
}

// TestAddDebugVisualization tests the addDebugVisualization method
func (s *LogoCreatorTestSuite) TestAddDebugVisualization() {
	// Create a test context
	contextWidth := 800
	logoAreaHeight := 500.0
	imageContext := gg.NewContext(contextWidth, int(logoAreaHeight))

	// Execute
	s.creator.addDebugVisualization(imageContext, contextWidth, logoAreaHeight)

	// Verify (this is a visual test, so we can only check that it doesn't panic)
	assert.NotNil(s.T(), imageContext)
}

// TestSaveDebugImage tests the saveDebugImage method
func (s *LogoCreatorTestSuite) TestSaveDebugImage() {
	// Create a test context
	imageContext := gg.NewContext(800, 500)
	imageContext.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	imageContext.Clear()

	// Create the debug directory if it doesn't exist
	debugDir := filepath.Join("internal", "app", "images")
	err := os.MkdirAll(debugDir, 0755)
	require.NoError(s.T(), err)

	// Execute
	err = s.creator.saveDebugImage(imageContext)
	require.NoError(s.T(), err)

	// Verify debug image was created
	debugPath := filepath.Join(debugDir, "image.png")
	_, err = os.Stat(debugPath)
	assert.NoError(s.T(), err)

	// Clean up the debug image
	err = os.Remove(debugPath)
	require.NoError(s.T(), err)
}

// TestLogoCreatorSuite runs the test suite
func TestLogoCreatorSuite(t *testing.T) {
	suite.Run(t, new(LogoCreatorTestSuite))
}
