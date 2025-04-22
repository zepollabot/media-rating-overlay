package image

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/zepollabot/media-rating-overlay/internal/model"
	image_mocks "github.com/zepollabot/media-rating-overlay/internal/processor/image/mocks"
)

type ImageServiceTestSuite struct {
	suite.Suite
	service     *ImageService
	logger      *zap.Logger
	fileManager *image_mocks.FileManager
	config      *model.PosterConfig
	tempDir     string
}

func (s *ImageServiceTestSuite) SetupSuite() {
	s.logger = zap.NewNop()
	s.fileManager = image_mocks.NewFileManager(s.T())
	s.config = model.PosterConfigWithDefaultValues()
	s.service = NewImageService(s.logger, s.config, s.fileManager)
}

func (s *ImageServiceTestSuite) SetupTest() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "image-service-test-*")
	s.Require().NoError(err)
}

func (s *ImageServiceTestSuite) TearDownTest() {
	err := os.RemoveAll(s.tempDir)
	s.Require().NoError(err)
	s.fileManager.ExpectedCalls = nil
}

func (s *ImageServiceTestSuite) TestOpenImage() {
	// Arrange
	testImage := image.NewRGBA(image.Rect(0, 0, 100, 100))
	testImage.Set(0, 0, color.RGBA{255, 0, 0, 255})
	testFilePath := filepath.Join(s.tempDir, "test.png")
	err := s.service.SaveImageObject(testImage, testFilePath)
	s.Require().NoError(err)

	// Act
	img, err := s.service.OpenImage(testFilePath)

	// Assert
	s.Require().NoError(err)
	s.NotNil(img)
	s.Equal(100, img.Bounds().Dx())
	s.Equal(100, img.Bounds().Dy())
}

func (s *ImageServiceTestSuite) TestResizeImage() {
	// Arrange
	testImage := image.NewRGBA(image.Rect(0, 0, 100, 100))
	testImage.Set(0, 0, color.RGBA{255, 0, 0, 255})

	// Act
	resizedImage := s.service.ResizeImage(testImage, 50, 50)

	// Assert
	s.Equal(50, resizedImage.Bounds().Dx())
	s.Equal(50, resizedImage.Bounds().Dy())
}

func (s *ImageServiceTestSuite) TestSaveImage() {
	// Arrange
	testImage := image.NewRGBA(image.Rect(0, 0, 100, 100))
	originalPath := filepath.Join(s.tempDir, "original.jpg")
	expectedPosterPath := filepath.Join(s.tempDir, "poster.png")

	s.fileManager.On("GeneratePosterFilePath", originalPath, ".png").Return(expectedPosterPath)
	s.fileManager.On("BackupExistingPoster", expectedPosterPath).Return(nil)

	// Act
	savedPath, err := s.service.SaveImage(testImage, originalPath)

	// Assert
	s.Require().NoError(err)
	s.Equal(expectedPosterPath, savedPath)
	s.fileManager.AssertExpectations(s.T())

	// Verify file exists and has correct permissions
	fileInfo, err := os.Stat(savedPath)
	s.Require().NoError(err)
	s.Equal(os.FileMode(0644), fileInfo.Mode().Perm())
}

func (s *ImageServiceTestSuite) TestCreateContext() {
	// Act
	dc := s.service.CreateContext(100, 100)

	// Assert
	s.NotNil(dc)
	s.Equal(100, dc.Width())
	s.Equal(100, dc.Height())
}

func TestImageServiceSuite(t *testing.T) {
	suite.Run(t, new(ImageServiceTestSuite))
}
