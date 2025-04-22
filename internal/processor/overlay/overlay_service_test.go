package overlay

import (
	"errors"
	"image"
	"testing"

	"github.com/fogleman/gg"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	config_model "github.com/zepollabot/media-rating-overlay/internal/config/model"
	"github.com/zepollabot/media-rating-overlay/internal/model"
	model_mocks "github.com/zepollabot/media-rating-overlay/internal/model/mocks" // For model.Overlay mock
	overlay_mocks "github.com/zepollabot/media-rating-overlay/internal/processor/overlay/mocks"
)

type OverlayServiceSuite struct {
	suite.Suite
	mockImageService   *overlay_mocks.ImageService
	mockOverlayFactory *overlay_mocks.OverlayFactory
	mockOverlay        *model_mocks.Overlay // Updated to use mock from model/mocks
	logger             *zap.Logger
	posterConfig       *model.PosterConfig
	overlayService     *OverlayService
}

func (s *OverlayServiceSuite) SetupTest() {
	s.mockImageService = new(overlay_mocks.ImageService)
	s.mockOverlayFactory = new(overlay_mocks.OverlayFactory)
	s.mockOverlay = new(model_mocks.Overlay) // Updated instantiation
	s.logger = zap.NewNop()
	s.posterConfig = &model.PosterConfig{}
	s.posterConfig.Dimensions.Width = 1920
	s.posterConfig.Dimensions.Height = 1080

	s.overlayService = NewOverlayService(
		s.logger,
		s.mockImageService,
		s.mockOverlayFactory,
		s.posterConfig,
	)
}

func (s *OverlayServiceSuite) TearDownTest() {
	s.mockImageService.AssertExpectations(s.T())
	s.mockOverlayFactory.AssertExpectations(s.T())
	s.mockOverlay.AssertExpectations(s.T())
}

func TestOverlayServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OverlayServiceSuite))
}

func (s *OverlayServiceSuite) TestCreateDrawContextWithOverlay_FrameSuccess() {
	// Arrange
	filePath := "path/to/image.jpg"
	item := model.Item{ID: "item1"}
	overlayTypeFrame := "frame"
	libConfig := &config_model.Library{
		Overlay: config_model.Overlay{Type: overlayTypeFrame},
	}
	mockImage := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	mockDrawContext := gg.NewContext(s.posterConfig.Dimensions.Width, s.posterConfig.Dimensions.Height)

	s.mockImageService.On("OpenImage", filePath).Return(mockImage, nil).Once()
	s.mockImageService.On("CreateContext", s.posterConfig.Dimensions.Width, s.posterConfig.Dimensions.Height).Return(mockDrawContext).Once()

	// Assuming OverlayFactory.CreateOverlay now returns model.Overlay, error
	s.mockOverlayFactory.On("CreateOverlay", overlayTypeFrame).Return(s.mockOverlay, nil).Once()

	// s.mockOverlay is now *model_mocks.Overlay, which mocks model.Overlay
	s.mockOverlay.On("Apply", mockImage, mockDrawContext, libConfig).Once()

	// Act
	drawContext, err := s.overlayService.CreateDrawContextWithOverlay(filePath, item, libConfig)

	// Assert
	s.NoError(err)
	s.NotNil(drawContext)
	s.Equal(mockDrawContext, drawContext)
}

func (s *OverlayServiceSuite) TestCreateDrawContextWithOverlay_BarSuccess() {
	// Arrange
	filePath := "path/to/image.jpg"
	item := model.Item{ID: "item1"}
	overlayTypeBar := "bar"
	libConfig := &config_model.Library{
		Overlay: config_model.Overlay{Type: overlayTypeBar},
	}
	mockImage := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	mockDrawContext := gg.NewContext(s.posterConfig.Dimensions.Width, s.posterConfig.Dimensions.Height)

	s.mockImageService.On("OpenImage", filePath).Return(mockImage, nil).Once()
	s.mockImageService.On("CreateContext", s.posterConfig.Dimensions.Width, s.posterConfig.Dimensions.Height).Return(mockDrawContext).Once()

	s.mockOverlayFactory.On("CreateOverlay", overlayTypeBar).Return(s.mockOverlay, nil).Once()
	s.mockOverlay.On("Apply", mockImage, mockDrawContext, libConfig).Once()

	// Act
	drawContext, err := s.overlayService.CreateDrawContextWithOverlay(filePath, item, libConfig)

	// Assert
	s.NoError(err)
	s.NotNil(drawContext)
	s.Equal(mockDrawContext, drawContext)
}

func (s *OverlayServiceSuite) TestCreateDrawContextWithOverlay_OpenImageError() {
	// Arrange
	filePath := "path/to/image.jpg"
	item := model.Item{ID: "item1"}
	libConfig := &config_model.Library{
		Overlay: config_model.Overlay{Type: "frame"},
	}
	expectedError := errors.New("failed to open image")

	s.mockImageService.On("OpenImage", filePath).Return(nil, expectedError).Once()

	// Act
	drawContext, err := s.overlayService.CreateDrawContextWithOverlay(filePath, item, libConfig)

	// Assert
	s.Error(err)
	s.Nil(drawContext)
	s.Equal(expectedError, err)
}

func (s *OverlayServiceSuite) TestCreateDrawContextWithOverlay_FactoryError() {
	// Arrange
	filePath := "path/to/image.jpg"
	item := model.Item{ID: "item1"}
	overlayType := "frame"
	libConfig := &config_model.Library{
		Overlay: config_model.Overlay{Type: overlayType},
	}
	mockImage := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	mockDrawContext := gg.NewContext(s.posterConfig.Dimensions.Width, s.posterConfig.Dimensions.Height)
	expectedError := errors.New("factory failed to create overlay")

	s.mockImageService.On("OpenImage", filePath).Return(mockImage, nil).Once()
	s.mockImageService.On("CreateContext", s.posterConfig.Dimensions.Width, s.posterConfig.Dimensions.Height).Return(mockDrawContext).Once()

	s.mockOverlayFactory.On("CreateOverlay", overlayType).Return(nil, expectedError).Once()

	// Act
	drawContext, err := s.overlayService.CreateDrawContextWithOverlay(filePath, item, libConfig)

	// Assert
	s.Error(err)
	s.Nil(drawContext)
	s.Equal(expectedError, err)
}

func (s *OverlayServiceSuite) TestCreateDrawContextWithOverlay_InvalidOverlayType() {
	// Arrange
	filePath := "path/to/image.jpg"
	item := model.Item{ID: "item1"}
	overlayTypeInvalid := "invalid_type"
	libConfig := &config_model.Library{
		Overlay: config_model.Overlay{Type: overlayTypeInvalid},
	}
	mockImage := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	mockDrawContext := gg.NewContext(s.posterConfig.Dimensions.Width, s.posterConfig.Dimensions.Height)
	factorySpecificError := errors.New("invalid overlay type from factory: " + overlayTypeInvalid)

	s.mockImageService.On("OpenImage", filePath).Return(mockImage, nil).Once()
	s.mockImageService.On("CreateContext", s.posterConfig.Dimensions.Width, s.posterConfig.Dimensions.Height).Return(mockDrawContext).Once()
	s.mockOverlayFactory.On("CreateOverlay", overlayTypeInvalid).Return(nil, factorySpecificError).Once()

	// Act
	drawContext, err := s.overlayService.CreateDrawContextWithOverlay(filePath, item, libConfig)

	// Assert
	s.Error(err)
	s.Nil(drawContext)
	s.Equal(factorySpecificError, err)
}
