package factory

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	factory_mocks "github.com/zepollabot/media-rating-overlay/internal/factory/mocks"
	media_service_mocks "github.com/zepollabot/media-rating-overlay/internal/media-service/mocks"
	mediaModel "github.com/zepollabot/media-rating-overlay/internal/media-service/model"
)

type MediaServiceModelFactorySuite struct {
	suite.Suite
	mockBaseFactory *factory_mocks.MediaServiceBaseFactory
	logger          *zap.Logger
	factory         *MediaServiceModelFactory
}

func (s *MediaServiceModelFactorySuite) SetupTest() {
	s.mockBaseFactory = factory_mocks.NewMediaServiceBaseFactory(s.T())
	s.logger = zap.NewNop()
	s.factory = NewMediaServiceModelFactory(s.logger, s.mockBaseFactory)
}

func (s *MediaServiceModelFactorySuite) TearDownTest() {
	s.mockBaseFactory.AssertExpectations(s.T())
}

func TestMediaServiceModelFactorySuite(t *testing.T) {
	suite.Run(t, new(MediaServiceModelFactorySuite))
}

func (s *MediaServiceModelFactorySuite) TestNewMediaServiceModelFactory() {
	// Arrange
	logger := zap.NewNop()
	mockBaseFactory := factory_mocks.NewMediaServiceBaseFactory(s.T())

	// Act
	f := NewMediaServiceModelFactory(logger, mockBaseFactory)

	// Assert
	s.NotNil(f, "Factory should not be nil")
}

func (s *MediaServiceModelFactorySuite) TestCreate_PlexSuccess() {
	// Arrange
	mockPlexClient := media_service_mocks.NewMediaClient(s.T())
	mockLibraryService := media_service_mocks.NewLibraryService(s.T())
	mockItemService := media_service_mocks.NewItemService(s.T())
	expectedLibraries := []config.Library{{Name: "Movies", Path: "/movies"}}

	s.mockBaseFactory.On("BuildPlexComponents").Return(mockPlexClient, mockLibraryService, mockItemService, nil).Once()
	s.mockBaseFactory.On("GetLibraries").Return(expectedLibraries).Once()

	// Act
	mediaService, err := s.factory.Create(mediaModel.MediaServicePlex)

	// Assert
	s.NoError(err, "Should not return an error for Plex service")
	s.NotNil(mediaService, "MediaService should not be nil")
	s.Equal(mediaModel.MediaServicePlex, mediaService.Name, "MediaService name should be Plex")
	s.Equal(expectedLibraries, mediaService.Libraries, "Libraries should match")
	s.NotNil(mediaService.Client, "Client should not be nil")
	s.NotNil(mediaService.LibraryService, "LibraryService should not be nil")
	s.NotNil(mediaService.ItemService, "ItemService should not be nil")
	s.NotNil(mediaService.PosterService, "PosterService should not be nil")
}

func (s *MediaServiceModelFactorySuite) TestCreate_PlexBuildError() {
	// Arrange
	expectedError := errors.New("plex build error")
	s.mockBaseFactory.On("BuildPlexComponents").Return(nil, nil, nil, expectedError).Once()

	// Act
	mediaService, err := s.factory.Create(mediaModel.MediaServicePlex)

	// Assert
	s.Error(err, "Should return an error when BuildPlexComponents fails")
	s.Equal(expectedError, err, "Error should be the one returned by BuildPlexComponents")
	s.Equal(mediaModel.MediaService{}, mediaService, "MediaService should be empty on error")
}

func (s *MediaServiceModelFactorySuite) TestCreate_UnsupportedService() {
	// Arrange
	unsupportedServiceName := "unsupported"

	// Act
	mediaService, err := s.factory.Create(unsupportedServiceName)

	// Assert
	s.Error(err, "Should return an error for an unsupported service")
	s.Contains(err.Error(), "unsupported media service: unsupported", "Error message should indicate unsupported service")
	s.Equal(mediaModel.MediaService{}, mediaService, "MediaService should be empty on error")
}

func (s *MediaServiceModelFactorySuite) TestCreate_PlexSuccess_NilPosterServiceDependencies() {
	// Arrange
	// We are testing the factory, not the poster service itself.
	// The poster service constructor NewPlexPosterService(plexClient, f.logger, fileManager)
	// can handle nil client for its own internal checks if any, but here we ensure
	// that the MediaServiceModelFactory correctly passes along what it receives
	// or constructs. The critical part is that BuildPlexComponents provides the client.
	mockPlexClient := media_service_mocks.NewMediaClient(s.T()) // This mock is for plexClient passed to NewPlexPosterService
	mockLibraryService := media_service_mocks.NewLibraryService(s.T())
	mockItemService := media_service_mocks.NewItemService(s.T())
	expectedLibraries := []config.Library{{Name: "TV Shows", Path: "/tv"}}

	s.mockBaseFactory.On("BuildPlexComponents").Return(mockPlexClient, mockLibraryService, mockItemService, nil).Once()
	s.mockBaseFactory.On("GetLibraries").Return(expectedLibraries).Once()

	// Act
	mediaService, err := s.factory.Create(mediaModel.MediaServicePlex)

	// Assert
	s.NoError(err, "Should not return an error for Plex service")
	s.NotNil(mediaService, "MediaService should not be nil")
	s.Equal(mediaModel.MediaServicePlex, mediaService.Name)
	s.Equal(expectedLibraries, mediaService.Libraries)
	s.NotNil(mediaService.Client, "Plex client in MediaService should not be nil")
	s.NotNil(mediaService.PosterService, "PosterService in MediaService should not be nil")
}
